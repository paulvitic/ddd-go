package process

import (
	"fmt"

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
	subscriptions[ddd.EventType(model.UserRegistered{})] = u.onUserRegistered
	return subscriptions
}

func (u *userProcessor) onUserRegistered(ctx *ddd.Context, event ddd.Event) error {
	fmt.Println("onUserRegistered", event)

	repo, err := ddd.Resolve[ddd.Repository[model.User]](ctx)
	if err != nil {
		return err
	}

	user, err := repo.Load(event.AggregateID())
	if err != nil {
		return err
	}

	user.Approve()
	repo.Update(user)
	return nil
}
