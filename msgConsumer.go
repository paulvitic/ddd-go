package ddd

import (
	"context"
	"errors"

	"sync"
	"sync/atomic"
	"time"
)

// MessageSourceKey is the type for the message source context key
type MessageSourceKey struct{}

// MessageConsumer represents a component that consumes messages from an external source
// and translates them into domain events for processing by the application.
type MessageConsumer interface {
	// Target returns the name/identifier of the source this consumer targets
	Target() string

	// SetEventBus attaches an EventBus to the consumer
	SetEventBus(eventBus *EventBus)

	// ProcessMessage handles a single message from the source
	ProcessMessage(ctx context.Context, msg []byte) error

	// Start begins the message consumption process
	OnStart() error

	// Stop gracefully stops the message consumption process
	OnDestroy() error

	// Running returns the current state of the consumer
	Running() bool
}

// MessageTranslator converts raw messages into domain events
type MessageTranslator func(from []byte) (Event, error)

// BaseMessageConsumer provides basic functionality for message consumers
type baseMessageConsumer struct {
	target     string
	translator MessageTranslator
	running    atomic.Bool
	eventBus   *EventBus
	mutex      sync.RWMutex
}

// NewBaseMessageConsumer creates a new base message consumer
func NewBaseMessageConsumer(target string, translator MessageTranslator) *baseMessageConsumer {
	return &baseMessageConsumer{
		target:     target,
		translator: translator,
		running:    atomic.Bool{}, // Implicitly initialized to false
	}
}

// Target returns the name of the targeted message source
func (c *baseMessageConsumer) Target() string {
	return c.target
}

// SetEventBus attaches an EventBus to the consumer
func (c *baseMessageConsumer) SetEventBus(eventBus *EventBus) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.eventBus = eventBus
}

// OnStart begins the message consumption process
func (c *baseMessageConsumer) OnStart() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.running.Load() {
		return nil // Already running
	}

	if c.eventBus == nil {
		return errors.New("cannot start message consumer: event bus not set")
	}

	c.running.Store(true)
	return nil
}

// OnDestroy gracefully stops the message consumption process
func (c *baseMessageConsumer) OnDestroy() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.running.Load() {
		return nil // Already stopped
	}

	c.running.Store(false)
	return nil
}

// Running returns whether the consumer is currently active
func (c *baseMessageConsumer) Running() bool {
	return c.running.Load()
}

// ProcessMessage handles a single message
func (c *baseMessageConsumer) ProcessMessage(ctx context.Context, msg []byte) error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if !c.running.Load() {
		return errors.New("cannot process message: consumer not running")
	}

	if c.translator == nil {
		return errors.New("cannot process message: translator not set")
	}

	if c.eventBus == nil {
		return errors.New("cannot process message: event bus not set")
	}

	event, err := c.translator(msg)
	if err != nil {
		return err
	}

	// Enhance the context with message source information
	// Use a proper type instead of string as key
	// msgCtx := context.WithValue(ctx, MessageSourceKey{}, c.target)

	// Dispatch the event with context
	return c.eventBus.Dispatch(event)
}

//-------------------------------------------------------------
// InMemoryMessageConsumer for channel-based message transport
//-------------------------------------------------------------

// InMemoryMessageConsumer is a MessageConsumer that reads from an in-memory channel
type InMemoryMessageConsumer struct {
	*baseMessageConsumer
	log        *Logger
	channel    chan string
	processing chan string
	stopWait   sync.WaitGroup
	ctx        context.Context    // Internal context for goroutine management
	cancel     context.CancelFunc // Cancel function for cleanup
}

// NewInMemoryMessageConsumer creates a new consumer that reads from a string channel
func NewInMemoryMessageConsumer(target string, translator MessageTranslator, channel chan string) MessageConsumer {
	if channel == nil {
		panic(errors.New("channel cannot be nil"))
	}

	base := NewBaseMessageConsumer(target, translator)
	return &InMemoryMessageConsumer{
		log:                 NewLogger(),
		baseMessageConsumer: base,
		channel:             channel,
	}
}

// OnStart begins consuming messages from the channel
func (c *InMemoryMessageConsumer) OnStart() error {
	// Call the base implementation first
	if err := c.baseMessageConsumer.OnStart(); err != nil {
		return err
	}

	// Create internal context for goroutine management
	c.ctx, c.cancel = context.WithCancel(context.Background())
	c.processing = make(chan string)
	c.stopWait.Add(2) // Two goroutines will signal when done

	// First goroutine: read from input channel and send to processing channel
	go func() {
		defer c.stopWait.Done()
		for c.Running() {
			select {
			case <-c.ctx.Done():
				return
			case jsonString, ok := <-c.channel:
				if !ok {
					return // Channel closed
				}
				if c.Running() {
					select {
					case c.processing <- jsonString:
						// Message sent successfully
					case <-c.ctx.Done():
						return
					}
				}
			}
		}
	}()

	// Second goroutine: process messages from the processing channel
	go func() {
		defer c.stopWait.Done()
		for {
			select {
			case <-c.ctx.Done():
				return
			case jsonString, ok := <-c.processing:
				if !ok {
					return // Channel closed
				}

				// Create a derived context for each message
				msgCtx, cancel := context.WithCancel(c.ctx)
				err := c.ProcessMessage(msgCtx, []byte(jsonString))
				cancel() // Clean up the message context

				if err != nil {
					c.log.Error("Error processing message: %v", err)
				}
			}
		}
	}()

	c.log.Info("Started in-memory message consumer for target %s", c.Target())
	return nil
}

// OnDestroy gracefully stops consumption and waits for processing to complete
func (c *InMemoryMessageConsumer) OnDestroy() error {
	if !c.Running() {
		return nil // Already stopped
	}

	// Call base implementation to update running state
	if err := c.baseMessageConsumer.OnDestroy(); err != nil {
		return err
	}

	// Cancel the internal context to signal goroutines to stop
	if c.cancel != nil {
		c.cancel()
	}

	// Close the processing channel and wait for goroutines to finish
	if c.processing != nil {
		close(c.processing)

		// Wait for goroutines to finish with a timeout
		done := make(chan struct{})
		go func() {
			c.stopWait.Wait()
			close(done)
		}()

		// Create a timeout for cleanup
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		select {
		case <-ctx.Done():
			c.log.Warn("Timed out waiting for in-memory message consumer to stop")
		case <-done:
			// Successfully stopped
		}
	}

	c.log.Info("Stopped in-memory message consumer for target %s", c.Target())
	return nil
}

// GetMessageSource is a helper function to extract the message source from a context
func GetMessageSource(ctx context.Context) (string, bool) {
	source, ok := ctx.Value(MessageSourceKey{}).(string)
	return source, ok
}
