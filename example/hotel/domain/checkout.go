package domain

import (
	"errors"
	"github.com/paulvitic/ddd-go"
)

type checkout struct {
	go_ddd.Policy
}

func Checkout() go_ddd.Policy {
	return &checkout{
		go_ddd.NewPolicy([]interface{}{GuestLocationReceived{}}),
	}
}

func (c *checkout) When(event go_ddd.Event) (go_ddd.Command, error) {
	switch event.Type() {
	case go_ddd.EventType(GuestLocationReceived{}):
		payload := go_ddd.MapEventPayload(event, GuestLocationReceived{})
		if isCheckedOut(payload.Latitude, payload.Longitude) {
			return go_ddd.NewCommand(CheckoutGuest{}), nil
		}
		return nil, nil
	default:
		return nil, errors.New("unknown event type")
	}
}

func isCheckedOut(lat float64, long float64) bool {
	if lat < 0 && long < 0 {
		return true
	}
	return false
}
