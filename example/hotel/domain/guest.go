package domain

import (
	"errors"
	"github.com/paulvitic/ddd-go"
	"log"
)

type Guest struct {
	ID   int
	Name string
}
type Guests interface {
	go_ddd.View
	GuestInRoom(id int) *Guest
}

type GuestsBase struct {
	go_ddd.View
	add func(item *Guest)
}

func NewGuests(add func(item *Guest)) *GuestsBase {
	return &GuestsBase{
		View: go_ddd.NewView(CheckedOut{}),
		add:  add,
	}
}

func (v *GuestsBase) MutateWhen(event go_ddd.Event) error {
	switch event.Type() {
	case go_ddd.EventType(CheckedOut{}):
		payload := go_ddd.MapEventPayload(event, CheckedOut{})
		log.Printf("Guest %d checked out", payload)
		return nil
	default:
		return errors.New("unknown event type")
	}
}
