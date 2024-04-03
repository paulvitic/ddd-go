package go_ddd

import (
	"context"
	"errors"
	"log"
)

type MessageConsumer interface {
	SetEventBus(eventBus EventBus)
	Target() string
	ProcessMessage(msg []byte) error
	Start() error
	Stop()
}

type messageConsumer struct {
	target     string
	eventBus   EventBus
	translator MessageTranslator
}

func NewMessageConsumer(target string, translator MessageTranslator) MessageConsumer {
	return &messageConsumer{
		target:     target,
		translator: translator,
		eventBus:   nil,
	}
}

func (p *messageConsumer) SetEventBus(eventBus EventBus) {
	p.eventBus = eventBus
}

func (p *messageConsumer) Target() string {
	return p.target
}

func (p *messageConsumer) Start() error {
	return errors.New("BaseMessageConsumer: need a concrete implementation")
}

func (p *messageConsumer) Stop() {
	log.Printf("[Error] Stop not implemented")
}

func (p *messageConsumer) ProcessMessage(msg []byte) error {
	if p.translator == nil {
		return errors.New("MessageTranslator not set")
	}
	if p.eventBus == nil {
		return errors.New("EventBus not set")
	}
	event, err := p.translator(msg)
	if err != nil {
		return err
	}
	ctx := context.WithValue(context.Background(), "publishEvent", false)
	return p.eventBus.Dispatch(ctx, event)
}
