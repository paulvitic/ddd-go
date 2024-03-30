package domain

import "time"

type RoomCreated struct {
	RoomType RoomType
}

type RoomBooked struct {
	From time.Time
	To   time.Time
}

type GuestLocationReceived struct {
	MobileAppId string
	Latitude    float64
	Longitude   float64
}

type CheckedOut struct {
}
