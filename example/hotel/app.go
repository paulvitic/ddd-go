package hotel

import (
	app "github.com/paulvitic/ddd-go/application"
	"github.com/paulvitic/ddd-go/example/hotel/adapter"
	"github.com/paulvitic/ddd-go/example/hotel/application"
	"github.com/paulvitic/ddd-go/example/hotel/domain"
)

func Context() *app.Context {
	return app.NewContext("hotel").
		RegisterPolicy(domain.Checkout()).
		RegisterView(adapter.Guests()).
		RegisterCommandService(application.RoomService(adapter.RoomRepo())).
		RegisterQueryService(application.GuestsService(adapter.Guests())).
		RegisterHttpEndpoint(adapter.RoomEndpoint()).
		RegisterHttpEndpoint(adapter.GuestsEndpoint()).
		RegisterMessageConsumer(adapter.AuthMsgConsumer())
}
