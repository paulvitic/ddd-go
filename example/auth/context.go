package auth

import (
	"github.com/paulvitic/ddd-go/context"
	"github.com/paulvitic/ddd-go/example/auth/adapter"
)

func Context() *context.Context {
	return context.NewContext(context.Configuration{Name: "auth"}).
		RegisterMessageConsumer(adapter.HotelMsgConsumer()).
		RegisterMessageConsumer(adapter.ExtMsgConsumer())

}
