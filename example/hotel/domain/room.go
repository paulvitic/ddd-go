package domain

import (
	ddd "github.com/paulvitic/ddd-go"
	"time"
)

type RoomType string

const (
	Single = RoomType("single")
	Double = RoomType("double")
)

func (s RoomType) String() string {
	return string(s)
}

type Room struct {
	ddd.Aggregate
	roomType      RoomType
	availableFrom time.Time
}

func NewRoom(number int, roomType RoomType) *Room {
	room := &Room{
		ddd.NewAggregate(ddd.NewID(number), Room{}),
		roomType,
		time.Now(),
	}
	room.RegisterEvent(
		room.AggregateType(),
		room.ID(),
		RoomCreated{RoomType: room.roomType})
	return room
}

func (r *Room) RoomType() RoomType {
	return r.roomType
}

func (r *Room) IsAvailable(from time.Time) bool {
	return r.availableFrom.Equal(from) || r.availableFrom.Before(from)
}

func (r *Room) Book(from time.Time, to time.Time) {
	if r.IsAvailable(from) {
		r.RegisterEvent(
			r.AggregateType(),
			r.ID(),
			RoomBooked{From: from, To: to})
	}
}
