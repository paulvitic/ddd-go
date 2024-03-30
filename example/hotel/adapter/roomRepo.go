package adapter

import (
	"errors"
	ddd "github.com/paulvitic/ddd-go"
	"github.com/paulvitic/ddd-go/example/hotel/domain"
)

type roomRepo struct {
	rooms map[ddd.ID]*domain.Room
}

func RoomRepo() ddd.Repository[domain.Room] {
	return &roomRepo{rooms: make(map[ddd.ID]*domain.Room)}
}

func (r *roomRepo) Save(aggregate *domain.Room) error {
	r.rooms[aggregate.ID()] = aggregate
	return nil
}

func (r *roomRepo) Load(id ddd.ID) (*domain.Room, error) {
	room, ok := r.rooms[id]
	if !ok {
		return &domain.Room{}, errors.New("room not found")
	}
	return room, nil
}

func (r *roomRepo) LoadAll() ([]*domain.Room, error) {
	rooms := make([]*domain.Room, 0, len(r.rooms))
	for _, room := range r.rooms {
		rooms = append(rooms, room)
	}
	return rooms, nil
}

func (r *roomRepo) Delete(id ddd.ID) error {
	delete(r.rooms, id)
	return nil
}

func (r *roomRepo) Update(aggregate *domain.Room) error {
	r.rooms[aggregate.ID()] = aggregate
	return nil
}
