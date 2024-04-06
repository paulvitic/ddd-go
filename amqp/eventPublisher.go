package amqp

import (
	"context"
	"github.com/paulvitic/ddd-go"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"time"
)

type eventPublisher struct {
	config  Configuration
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   *amqp.Queue
}

func NewEventPublisher(config Configuration) ddd.EventPublisher {
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

	queue, err := declareQueue(ch, config.Queue)
	failOnError(err, "Failed to declare queue")

	err = bindQueue(ch, config.Exchange, config.Queue)
	failOnError(err, "Failed to bind queue")

	publisher.queue = &queue

	return publisher
}

func (p *eventPublisher) Publish(event ddd.Event) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if body, err := event.ToJsonString(); err != nil {
		return err
	} else {
		if err = p.channel.PublishWithContext(ctx,
			p.config.Exchange, // exchange
			p.queue.Name,      // routing key
			false,             // mandatory
			false,             // immediate
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

func (p *eventPublisher) Middleware() ddd.MiddlewareFunc {
	return func(next ddd.HandlerFunc) ddd.HandlerFunc {
		return func(ctx context.Context, msg ddd.Payload) (interface{}, error) {
			publishEvent, ok := ctx.Value("publishEvent").(bool)
			if !ok || (ok && publishEvent) {
				err := p.Publish(msg.(ddd.Event))
				if err != nil {
					return nil, err
				}
			}
			return next(ctx, msg)
		}
	}
}

func (p *eventPublisher) Queue() interface{} {
	return p.config
}

func (p *eventPublisher) Close() {
	p.queue = nil
	err := p.channel.Close()
	failOnError(err, "Failed to close channel")
	err = p.conn.Close()
	failOnError(err, "Failed to close connection")
}
