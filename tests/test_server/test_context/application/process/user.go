package process

import (
	"github.com/paulvitic/ddd-go"
	"github.com/paulvitic/ddd-go/tests/test_server/test_context/domain/model"
)

type userProcessor struct {
}

func UserProcessor() ddd.EventHandler {
	return &userProcessor{}
}

func (u *userProcessor) SubscribedTo() map[string]ddd.HandleEvent {
	subscriptions := make(map[string]ddd.HandleEvent)
	subscriptions[ddd.EventType(model.UserRegistered{})] = u.onRegistered
	return subscriptions
}

func (u *userProcessor) onRegistered(ctx *ddd.Context, event ddd.Event) error {
	if repo, err := ddd.Resolve[ddd.Repository[model.User]](ctx); err != nil {
		return err
	} else {
		if user, err := repo.Load(event.AggregateID()); err != nil {
			return err
		} else {
			user.Approve()
			repo.Update(user)
			return nil
		}
	}
}
