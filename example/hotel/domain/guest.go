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
	ddd.View
	GuestInRoom(id int) *Guest
}

type GuestsBase struct {
	ddd.View
	add func(item *Guest)
}

func NewGuests(add func(item *Guest)) *GuestsBase {
	return &GuestsBase{
		View: ddd.NewView(CheckedOut{}),
		add:  add,
	}
}

func (v *GuestsBase) MutateWhen(event ddd.Event) error {
	switch event.Type() {
	case ddd.EventType(CheckedOut{}):
		payload := ddd.MapEventPayload(event, CheckedOut{})
		log.Printf("Guest %d checked out", payload)
		return nil
	default:
		return errors.New("unknown event type")
	}
}
