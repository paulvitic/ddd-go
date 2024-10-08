package amqp

import (
	"context"
	"github.com/paulvitic/ddd-go"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type MockMessageConsumer struct {
	ddd.MessageConsumer
	ReceivedMessage []byte
}

func (m *MockMessageConsumer) ProcessMessage(msg []byte) error {
	m.ReceivedMessage = msg
	return nil
}

func TestMessageConsumer_Start(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	// Setup
	config := Configuration{
		Host:        "cow.rmq2.cloudamqp.com",
		Port:        5672,
		Username:    "domkxzfl",
		Password:    "xxxxxxxxxxx",
		VirtualHost: "domkxzfl",
		Exchange:    "test",
		Queue:       "test",
	}
	mockConsumer := &MockMessageConsumer{}
	consumer := MessageConsumer(mockConsumer, config)

	// Start the consumer
	go func() {
		err := consumer.Start()
		if err != nil {
			t.Error(err)
			return
		}
	}()

	// Give the consumer some time to start
	time.Sleep(2 * time.Second)

	conn, err := amqp.Dial(connectionUrl(config))
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	err = declareExchange(ch, config.Exchange)
	failOnError(err, "Failed to declare exchange")

	queue, err := declareQueue(ch, config.Queue)
	failOnError(err, "Failed to declare queue")

	err = bindQueue(ch, config.Exchange, config.Queue)
	failOnError(err, "Failed to bind queue")

	// Send a message to the queue
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = ch.PublishWithContext(
		ctx,             // context
		config.Exchange, // exchange
		queue.Name,      // routing key
		false,           // mandatory
		false,           // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte("Hello, world!"),
		})
	if err != nil {
		t.Fatal(err)
	}

	// Give the consumer some time to process the message
	time.Sleep(2 * time.Second)

	// Check if the message was received and processed
	assert.Equal(t, []byte("Hello, world!"), mockConsumer.ReceivedMessage)

	// Cleanup
	consumer.Stop()
}
