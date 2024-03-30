package amqp

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"strconv"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

type Configuration struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	VirtualHost string `json:"virtualHost"`
	Exchange    string `json:"exchange"`
	Queue       string `json:"queue"`
}

func connectionUrl(settings Configuration) string {
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
		strconv.Itoa(settings.Port) + "/"
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
