package http

import (
	"github.com/paulvitic/ddd-go"
	"log"
	"net/http"
)

type CommandEndpoint struct {
	*EndpointBase
	translator CommandTranslator
	commandBus ddd.CommandBus
}

func NewCommandEndpoint(path string, methods []string, translator CommandTranslator) *CommandEndpoint {
	return &CommandEndpoint{
		EndpointBase: NewEndpoint(path, methods),
		translator:   translator,
		commandBus:   nil,
	}
}

func (e *CommandEndpoint) RegisterCommandBus(bus ddd.CommandBus) {
	if e.commandBus != nil {
		log.Printf("Command bus already set for endpoint %s", e.path)
	}
	e.commandBus = bus
}

func (e *CommandEndpoint) Handler() func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		if e.translator == nil {
			http.Error(writer, "No translator for POST/PUT/DELETE methods", http.StatusNotAcceptable)
			return
		}

		command, err := e.translator(request)
		if err != nil {
			log.Printf("Error translating command: %v", err)
			http.Error(writer, "Bad request", http.StatusBadRequest)
			return
		}

		if command == nil {
			log.Printf("command translator returned nil command")
			http.Error(writer, "Internal server error", http.StatusInternalServerError)
			return
		}

		if command, ok := command.(ddd.Command); ok {
			err = e.commandBus.Dispatch(request.Context(), command)
			if err != nil {
				log.Printf("Error dispatching command: %v", err)
				http.Error(writer, "Internal server error", http.StatusInternalServerError)
				return
			}
			writer.WriteHeader(http.StatusAccepted)
		} else {
			log.Printf("No command bus to dispatch command %s", command.Type())
			http.Error(writer, "Not acceptable", http.StatusNotAcceptable)
		}
	}
}
