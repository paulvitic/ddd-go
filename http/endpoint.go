package http

import (
	"encoding/json"
	ddd "github.com/paulvitic/ddd-go"
	"log"
	"net/http"
)

type Endpoint interface {
	Path() string
	WithCommandTranslator(CommandTranslator) Endpoint
	WithQueryTranslator(QueryTranslator) Endpoint
	RegisterCommandBus(ddd.CommandBus)
	RegisterQueryBus(ddd.QueryBus)
	Methods() []string
	Handler() func(http.ResponseWriter, *http.Request)
}

type endpoint struct {
	path              string
	methods           []string
	commandBus        ddd.CommandBus
	queryBus          ddd.QueryBus
	commandTranslator CommandTranslator
	queryTranslator   QueryTranslator
}

func NewEndpoint(path string) Endpoint {
	return &endpoint{
		path:              path,
		methods:           make([]string, 0),
		commandTranslator: nil,
		commandBus:        nil,
		queryTranslator:   nil,
		queryBus:          nil,
	}
}

func (e *endpoint) Path() string {
	return e.path
}

func (e *endpoint) RegisterCommandBus(bus ddd.CommandBus) {
	if e.commandBus != nil {
		log.Printf("Command bus already set for endpoint %s", e.path)
	}
	e.commandBus = bus
}

func (e *endpoint) RegisterQueryBus(bus ddd.QueryBus) {
	if e.queryBus != nil {
		log.Printf("Query bus already set for endpoint %s", e.path)
	}
	e.queryBus = bus
}
func (e *endpoint) WithCommandTranslator(translator CommandTranslator) Endpoint {
	if e.commandTranslator != nil {
		log.Printf("Command translator already set for endpoint %s", e.path)
	}
	e.commandTranslator = translator
	e.methods = append(e.methods, http.MethodPost)
	e.methods = append(e.methods, http.MethodPut)
	e.methods = append(e.methods, http.MethodDelete)
	return e
}

func (e *endpoint) WithQueryTranslator(translator QueryTranslator) Endpoint {
	if e.queryTranslator != nil {
		log.Printf("Query translator already set for endpoint %s", e.path)
	}
	e.queryTranslator = translator
	e.methods = append(e.methods, http.MethodGet)
	return e
}

func (e *endpoint) Methods() []string {
	return e.methods
}

func (e *endpoint) Handler() func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodGet:
			e.handleGet(writer, request)
		case http.MethodPost, http.MethodPut, http.MethodDelete:
			e.handleCommand(writer, request)
		default:
			http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func (e *endpoint) handleGet(writer http.ResponseWriter, request *http.Request) {
	if e.queryTranslator == nil {
		http.Error(writer, "No translator for GET method", http.StatusNotAcceptable)
		return
	}

	query, err := e.queryTranslator(request)
	if err != nil {
		log.Printf("Error translating query: %v", err)
		http.Error(writer, "Bad request", http.StatusBadRequest)
		return
	}

	if query == nil {
		log.Printf("query translator returned nil query")
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}

	if query, ok := query.(ddd.Query); ok && e.queryBus != nil {
		res, err := e.queryBus.Dispatch(request.Context(), query)
		if err != nil {
			log.Printf("Error dispatching query: %v", err)
			http.Error(writer, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Write response
		writer.WriteHeader(http.StatusOK)
		err = json.NewEncoder(writer).Encode(res)
		if err != nil {
			log.Printf("Error encoding response: %v", err)
			http.Error(writer, "Internal server error", http.StatusInternalServerError)
			return
		}
	} else {
		log.Printf("No query bus to dispatch query %s", query.Type())
		http.Error(writer, "Not acceptable", http.StatusNotAcceptable)
	}
}

func (e *endpoint) handleCommand(writer http.ResponseWriter, request *http.Request) {
	if e.commandTranslator == nil {
		http.Error(writer, "No translator for POST/PUT/DELETE methods", http.StatusNotAcceptable)
		return
	}

	command, err := e.commandTranslator(request)
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
