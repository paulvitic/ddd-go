package application

import (
	"context"
	ddd "github.com/paulvitic/ddd-go"
	"github.com/paulvitic/ddd-go/example/hotel/domain"
)

type roomService struct {
	ddd.CommandService
}

func (s *roomService) Create(repo ddd.Repository[domain.Room], cmd domain.CreateRoom) (*domain.Room, error) {
	room := domain.NewRoom(cmd.Number, cmd.RoomType)
	err := repo.Save(room)
	return room, err
}

func (s *roomService) Book(repo ddd.Repository[domain.Room], cmd domain.BookRoom) (*domain.Room, error) {
	room, err := repo.Load(ddd.NewID(cmd.Number))
	if err != nil {
		return nil, err
	}
	room.Book(cmd.From, cmd.To)
	err = repo.Update(room)
	return room, err
}

func RoomCommandExecutor(service *roomService, repo ddd.Repository[domain.Room]) ddd.CommandExecutor {
	return func(ctx context.Context, cmd ddd.Command) error {
		var room *domain.Room
		var err error

		switch cmd.Type() {
		case ddd.EventType(domain.CreateRoom{}):
			room, err = service.Create(repo, cmd.Body().(domain.CreateRoom))
		case ddd.EventType(domain.BookRoom{}):
			room, err = service.Book(repo, cmd.Body().(domain.BookRoom))
		}

		if err != nil {
			return err
		}
		return service.DispatchFrom(ctx, room)
	}
}

func RoomService(repo ddd.Repository[domain.Room]) ddd.CommandService {
	return &roomService{
		ddd.NewCommandService(
			RoomCommandExecutor(&roomService{}, repo), domain.CreateRoom{}, domain.BookRoom{}),
	}
}
