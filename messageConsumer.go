package go_ddd

import (
	"context"
	"errors"
	"log"
)

type MessageConsumer interface {
	WithEventBus(eventBus EventBus) MessageConsumer
	WithEventTranslator(translator EventTranslator) MessageConsumer
	Target() string
	ProcessMessage(msg []byte) error
	Start() error
	Stop()
}

type messageConsumer struct {
	target     string
	eventBus   EventBus
	translator EventTranslator
}

func NewMessageConsumer(target string) MessageConsumer {
	return &messageConsumer{
		target:     target,
		eventBus:   nil,
		translator: nil,
	}
}

func (p *messageConsumer) WithEventBus(eventBus EventBus) MessageConsumer {
	p.eventBus = eventBus
	return p
}

func (p *messageConsumer) WithEventTranslator(translator EventTranslator) MessageConsumer {
	p.translator = translator
	return p
}

func (p *messageConsumer) Target() string {
	return p.target
}

func (p *messageConsumer) Start() error {
	return errors.New("need a concrete implementation of message consumer")
}

func (p *messageConsumer) Stop() {
	log.Printf("[Error] Stop not implemented")
}

func (p *messageConsumer) ProcessMessage(msg []byte) error {
	if p.translator == nil {
		return errors.New("EventTranslator not set")
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
