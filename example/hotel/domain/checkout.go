package domain

import (
	"errors"
	"github.com/paulvitic/ddd-go"
)

type checkout struct {
	ddd.Policy
}

func Checkout() ddd.Policy {
	return &checkout{
		ddd.NewPolicy([]interface{}{GuestLocationReceived{}}),
	}
}

func (c *checkout) When(event ddd.Event) (ddd.Command, error) {
	switch event.Type() {
	case ddd.EventType(GuestLocationReceived{}):
		payload := ddd.MapEventPayload(event, GuestLocationReceived{})
		if isCheckedOut(payload.Latitude, payload.Longitude) {
			return ddd.NewCommand(CheckoutGuest{}), nil
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
