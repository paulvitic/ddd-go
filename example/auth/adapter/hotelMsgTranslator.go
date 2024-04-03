package adapter

import ddd "github.com/paulvitic/ddd-go"

func hotelMsgTranslator(from []byte) (ddd.Event, error) {
	return nil, nil
}
func HotelMsgConsumer() ddd.MessageConsumer {
	return ddd.NewMessageConsumer("hotel", hotelMsgTranslator)
}
