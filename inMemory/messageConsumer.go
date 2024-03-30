package inMemory

import (
	"errors"
	go_ddd "github.com/paulvitic/ddd-go"
	"log"
	"sync/atomic"
)

type messageConsumer struct {
	go_ddd.MessageConsumer
	channel    *chan string
	processing chan string
	running    atomic.Bool
}

func MessageConsumer(base go_ddd.MessageConsumer, channel *chan string) go_ddd.MessageConsumer {
	if channel == nil {
		panic(errors.New("queue cannot be nil"))
	}

	return &messageConsumer{
		MessageConsumer: base,
		channel:         channel,
		running:         atomic.Bool{}, // Implicitly set to false
		processing:      nil,
	}
}

func (p *messageConsumer) Start() error {
	p.processing = make(chan string)
	p.running.Store(true)

	go func() {
		for p.running.Load() {
			select {
			case jsonString := <-*p.channel:
				p.processing <- jsonString
			}
		}
	}()

	go func() {
		for jsonString := range p.processing {
			err := p.MessageConsumer.ProcessMessage([]byte(jsonString))
			if err != nil {
				log.Println(err)
			}
		}
	}()
	return nil
}

func (p *messageConsumer) Stop() {
	close(p.processing)
	p.running.Store(false)
}
