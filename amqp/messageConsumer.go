package amqp

import (
	"github.com/paulvitic/ddd-go"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

type messageConsumer struct {
	go_ddd.MessageConsumer
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   *amqp.Queue
}

func MessageConsumer(base go_ddd.MessageConsumer, config Configuration) (go_ddd.MessageConsumer, error) {
	conn, err := amqp.Dial(connectionUrl(config))
	failOnError(err, "Failed to connect to RabbitMQ")

	consumer := &messageConsumer{
		MessageConsumer: base,
		conn:            conn,
	}
	defer consumer.Stop()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	consumer.channel = ch

	queue, err := declareQueue(ch, config.Queue)
	failOnError(err, "Failed to declare queue")
	consumer.queue = &queue

	return consumer, nil
}

func (p *messageConsumer) Start() error {
	msgs, err := p.channel.Consume(
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
		for msg := range msgs {
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
