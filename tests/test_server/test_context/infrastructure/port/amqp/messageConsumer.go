package amqp

import (
	"context"
	"log"
	"strconv"

	ddd "github.com/paulvitic/ddd-go"
	amqp "github.com/rabbitmq/amqp091-go"
)

type messageConsumer struct {
	ddd.MessageConsumer
	config  *AmqpConfig
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   *amqp.Queue
}

func MessageConsumer(base ddd.MessageConsumer, amqpConfig *AmqpConfig) ddd.MessageConsumer {
	return &messageConsumer{
		MessageConsumer: base,
		config:          amqpConfig,
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

	ctx := context.Background()

	go func() {
		for msg := range messages {
			log.Printf("Received a message: %s", msg.Body)
			err := p.MessageConsumer.ProcessMessage(ctx, msg.Body)
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

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func connectionUrl(settings *AmqpConfig) string {
	if settings.Username == "" {
		settings.Username = "guest"
	}
	if settings.Password == "" {
		settings.Password = "guest"
	}
	if settings.Host == "" {
		settings.Host = "localhost"
	}
	if settings.Port == 0 {
		settings.Port = 5672
	}
	return "amqp://" + settings.Username + ":" +
		settings.Password + "@" + settings.Host + ":" +
		strconv.Itoa(settings.Port) + "/" + settings.VirtualHost
}

func declareExchange(ch *amqp.Channel, name string) error {
	return ch.ExchangeDeclare(
		name,
		"fanout", // kind (fanout for distributing messages to all bound queues)
		true,     // durable (survive broker restarts)
		false,    // auto-delete (don't delete when no consumers are connected)
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
}

func declareQueue(ch *amqp.Channel, name string) (amqp.Queue, error) {
	return ch.QueueDeclare(
		name,
		true,  // durable (survive broker restarts)
		true,  // auto-delete (delete when no consumers are connected)
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
}

func bindQueue(ch *amqp.Channel, exchange, queue string) error {
	return ch.QueueBind(
		queue,    // queue name
		"",       // routing key (empty for fanout exchange)
		exchange, // exchange name
		false,    // no-wait
		nil,      // arguments
	)
}
