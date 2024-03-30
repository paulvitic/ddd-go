package domain

import "time"

type CreateRoom struct {
	Number   int
	RoomType RoomType
}

type BookRoom struct {
	Number int
	From   time.Time
	To     time.Time
}

type CheckoutGuest struct {
	RoomNumber int
}
