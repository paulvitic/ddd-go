package amqp

import (
	"context"
	"github.com/paulvitic/ddd-go"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"time"
)

type eventPublisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   *amqp.Queue
}

func NewEventPublisher(config Configuration) go_ddd.EventPublisher {
	conn, err := amqp.Dial(connectionUrl(config))
	failOnError(err, "Failed to connect to RabbitMQ")
	publisher := &eventPublisher{
		conn: conn,
	}
	defer publisher.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	publisher.channel = ch

	err = declareExchange(ch, config.Exchange)
	failOnError(err, "Failed to declare exchange")

	err = declareExchange(ch, config.Exchange)
	failOnError(err, "Failed to declare exchange")

	queue, err := declareQueue(ch, config.Queue)
	failOnError(err, "Failed to declare queue")

	err = bindQueue(ch, config.Exchange, config.Queue)
	failOnError(err, "Failed to bind queue")

	publisher.queue = &queue

	return publisher
}

func (p *eventPublisher) Publish(event go_ddd.Event) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if body, err := event.ToJsonString(); err != nil {
		return err
	} else {
		if err = p.channel.PublishWithContext(ctx,
			"",           // exchange
			p.queue.Name, // routing key
			false,        // mandatory
			false,        // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(body),
			}); err != nil {
			return err
		}
		log.Printf(" [x] Sent %s\n", body)
		return nil
	}
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
	return p.queue
}

func (p *eventPublisher) Close() {
	p.queue = nil
	err := p.channel.Close()
	failOnError(err, "Failed to close channel")
	err = p.conn.Close()
	failOnError(err, "Failed to close connection")
}
