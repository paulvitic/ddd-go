package adapter

import (
	"github.com/paulvitic/ddd-go"
	"github.com/paulvitic/ddd-go/example/hotel/domain"
)

type guests struct {
	go_ddd.View
	// In-memory store
	store map[int]*domain.Guest
}

func Guests() domain.Guests {
	store := make(map[int]*domain.Guest)

	add := func(item *domain.Guest) {
		store[item.ID] = item
	}

	return &guests{
		View:  domain.NewGuests(add),
		store: store,
	}
}

func (g *guests) GuestInRoom(id int) *domain.Guest {
	return g.store[id]
}

func (g *guests) MutateWhen(event go_ddd.Event) error {
	return g.View.MutateWhen(event)
}
