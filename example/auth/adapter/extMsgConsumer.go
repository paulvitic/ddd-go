package adapter

import (
	ddd "github.com/paulvitic/ddd-go"
	"github.com/paulvitic/ddd-go/inMemory"
)

func ExtMsgTranslator(from []byte) (ddd.Event, error) {
	return nil, nil
}

func ExtMsgConsumer() ddd.MessageConsumer {
	//return amqp.MessageConsumer(
	//	ddd.NewMessageConsumer("external", ExtMsgTranslator),
	//	amqp.Configuration{
	//		Host:        "cow.rmq2.cloudamqp.com",
	//		Port:        5672,
	//		Username:    "domkxzfl",
	//		Password:    "pAXyK5EOSI2RX-nJftcmUFjdrRbMa_5f",
	//		VirtualHost: "domkxzfl",
	//		Exchange:    "external",
	//		Queue:       "events",
	//	})
	queue := make(chan string, 1000)
	return inMemory.MessageConsumer(
		ddd.NewMessageConsumer("external", ExtMsgTranslator),
		&queue)
}
