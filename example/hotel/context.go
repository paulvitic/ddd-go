package hotel

import (
	"github.com/paulvitic/ddd-go/context"
	"github.com/paulvitic/ddd-go/example/hotel/adapter"
	"github.com/paulvitic/ddd-go/example/hotel/application"
	"github.com/paulvitic/ddd-go/example/hotel/domain"
)

func Context() *context.Context {
	return context.NewContext(context.Configuration{Name: "hotel"}).
		RegisterPolicy(domain.Checkout()).
		RegisterView(adapter.Guests()).
		RegisterCommandEndpoint(adapter.RoomEndpoint()).
		RegisterCommandService(application.RoomService(adapter.RoomRepo())).
		RegisterQueryEndpoint(adapter.GuestsEndpoint()).
		RegisterQueryService(application.GuestsService(adapter.Guests())).
		RegisterMessageConsumer(adapter.AuthMsgConsumer())
}
