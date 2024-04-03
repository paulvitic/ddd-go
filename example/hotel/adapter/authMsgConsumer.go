package adapter

import ddd "github.com/paulvitic/ddd-go"

func AuthMsgTranslator(from []byte) (ddd.Event, error) {
	return nil, nil
}
func AuthMsgConsumer() ddd.MessageConsumer {
	return ddd.NewMessageConsumer("auth", AuthMsgTranslator)
}
