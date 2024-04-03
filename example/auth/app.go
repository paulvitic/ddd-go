package auth

import (
	app "github.com/paulvitic/ddd-go/application"
	"github.com/paulvitic/ddd-go/example/auth/adapter"
)

func Context() *app.Context {
	return app.NewContext("auth").
		RegisterMessageConsumer(adapter.HotelMsgConsumer()).
		RegisterMessageConsumer(adapter.ExtMsgConsumer())

}
