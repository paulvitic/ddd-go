package process

import "github.com/paulvitic/ddd-go"

type userProcessor struct {
}

func UserProcessor() ddd.EventHandler {
	return &userProcessor{}
}

func (u *userProcessor) SubscribedTo() map[string]ddd.HandleEvent {
	subscriptions := make()
}

func (u *userProcessor) onRegister(ctx *ddd.Context, event ddd.Event) error {

}
