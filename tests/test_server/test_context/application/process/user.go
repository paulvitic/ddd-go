package process

import (
	"github.com/paulvitic/ddd-go"
	"github.com/paulvitic/ddd-go/tests/test_server/test_context/domain/model"
	"github.com/paulvitic/ddd-go/tests/test_server/test_context/domain/repository"
)

type userProcessor struct {
	repo repository.UserRepository
}

func UserProcessor(ctx *ddd.Context) ddd.EventHandler {
	if repo, err := ddd.Resolve[repository.UserRepository](ctx); err != nil {
		panic("repo nor found")
	} else {
		return &userProcessor{
			repo: repo,
		}
	}
}

func (u *userProcessor) SubscribedTo() map[string]ddd.HandleEvent {
	subscriptions := make(map[string]ddd.HandleEvent)
	subscriptions[ddd.EventType(model.UserRegistered{})] = u.onRegistered
	return subscriptions
}

func (u *userProcessor) onRegistered(event ddd.Event) error {
	user, err := u.repo.Load(event.AggregateID())
	if err != nil {
		return err
	}
	user.Approve()
	u.repo.Update(user)
	return nil

}
