package go_ddd

import (
	"context"
)

type CommandExecutor func(context.Context, Command) error

type CommandService interface {
	SubscribedTo() []string
	Executor() CommandExecutor
	WithEventBus(bus EventBus) CommandService
	DispatchFrom(ctx context.Context, producer EventProducer) error
}

type commandService struct {
	subscribedTo []string
	executor     CommandExecutor
	eventBus     EventBus
}

func NewCommandService(handler CommandExecutor, subscribedTo ...interface{}) CommandService {
	var cmdTypes []string
	for _, c := range subscribedTo {
		cmdTypes = append(cmdTypes, NewCommand(c).Type())
	}
	return &commandService{
		subscribedTo: cmdTypes,
		executor:     handler,
	}
}

func (p *commandService) WithEventBus(bus EventBus) CommandService {
	p.eventBus = bus
	return p
}

func (p *commandService) SubscribedTo() []string {
	return p.subscribedTo
}

func (p *commandService) Executor() CommandExecutor {
	return p.executor
}

func (p *commandService) DispatchFrom(ctx context.Context, producer EventProducer) error {
	return p.eventBus.DispatchFrom(ctx, producer)
}
