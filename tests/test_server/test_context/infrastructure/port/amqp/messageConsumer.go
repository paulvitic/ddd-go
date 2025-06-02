package amqp

import (
	"log"

	ddd "github.com/paulvitic/ddd-go"
	amqp "github.com/rabbitmq/amqp091-go"
)

type messageConsumer struct {
	ddd.MessageConsumer
	config  ddd.Configuration
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   *amqp.Queue
}

func MessageConsumer(base ddd.MessageConsumer, config Configuration) ddd.MessageConsumer {
	return &messageConsumer{
		MessageConsumer: base,
		config:          config,
	}
}

func (p *messageConsumer) Start() error {
	conn, err := amqp.Dial(connectionUrl(p.config))
	failOnError(err, "Failed to connect to RabbitMQ")
	p.conn = conn

	defer p.Stop()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	p.channel = ch

	queue, err := declareQueue(ch, p.config.Queue)
	failOnError(err, "Failed to declare queue")
	p.queue = &queue

	messages, err := p.channel.Consume(
		p.queue.Name, // queue
		"",           // consumer
		true,         // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
	if err != nil {
		return err
	}

	var forever chan struct{}

	go func() {
		for msg := range messages {
			log.Printf("Received a message: %s", msg.Body)
			err := p.MessageConsumer.ProcessMessage(msg.Body)
			if err != nil {
				log.Println(err)
			}
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
	return nil
}

func (p *messageConsumer) Stop() {
	err := p.channel.Close()
	failOnError(err, "Failed to close channel")
	err = p.conn.Close()
	failOnError(err, "Failed to close connection")
}
