package inMemory

import (
	"context"
	"github.com/paulvitic/ddd-go"
)

type EventPublisherConfiguration struct {
	ChannelSettings struct {
		BufferSize int `json:"bufferSize"`
	} `json:"channelSettings"`
	ProducerSettings struct {
		MessageFormat string `json:"messageFormat"`
		MessageKey    string `json:"messageKey"`
	} `json:"producerSettings"`
}

type eventPublisher struct {
	queue chan string
}

func NewEventPublisher() go_ddd.EventPublisher {
	return &eventPublisher{
		queue: make(chan string),
	}
}

func (p *eventPublisher) Publish(event go_ddd.Event) error {
	jsonString, err := event.ToJsonString()
	if err != nil {
		return err
	}
	go func() { p.queue <- jsonString }()
	return nil
}

func (p *eventPublisher) Close() {
	close(p.queue)
}

func (p *eventPublisher) Middleware() go_ddd.MiddlewareFunc {
	return func(next go_ddd.HandlerFunc) go_ddd.HandlerFunc {
		return func(ctx context.Context, msg go_ddd.Payload) (interface{}, error) {
			publishEvent, ok := ctx.Value("publishEvent").(bool)
			if !ok || (ok && publishEvent) {
				err := p.Publish(msg.(go_ddd.Event))
				if err != nil {
					return nil, err
				}
			}
			return next(ctx, msg)
		}
	}
}

func (p *eventPublisher) Queue() interface{} {
	return &p.queue
}
