package hotel

import (
	app "github.com/paulvitic/ddd-go/application"
	"github.com/paulvitic/ddd-go/example/hotel/adapter"
	"github.com/paulvitic/ddd-go/example/hotel/application"
	"github.com/paulvitic/ddd-go/example/hotel/domain"
)

func NewHotel() *app.Context {
	return app.NewContext("hotel").
		RegisterPolicy(domain.Checkout()).
		RegisterView(adapter.Guests()).
		RegisterHttpEndpoint(adapter.RoomEndpoint()).
		RegisterHttpEndpoint(adapter.GuestsEndpoint()).
		RegisterCommandService(application.RoomService(adapter.RoomRepo())).
		RegisterQueryService(application.GuestsService(adapter.Guests()))
}
