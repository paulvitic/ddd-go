package http

import (
	"encoding/json"
	"github.com/paulvitic/ddd-go"
	"log"
	"net/http"
)

type Endpoint interface {
	Path() string
	WithCommandTranslator(CommandTranslator) Endpoint
	WithQueryTranslator(QueryTranslator) Endpoint
	RegisterCommandBus(go_ddd.CommandBus)
	RegisterQueryBus(go_ddd.QueryBus)
	Methods() []string
	Handler() func(http.ResponseWriter, *http.Request)
}

type endpoint struct {
	path              string
	methods           []string
	commandBus        go_ddd.CommandBus
	queryBus          go_ddd.QueryBus
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

func (e *endpoint) RegisterCommandBus(bus go_ddd.CommandBus) {
	if e.commandBus != nil {
		log.Printf("Command bus already set for endpoint %s", e.path)
	}
	e.commandBus = bus
}

func (e *endpoint) RegisterQueryBus(bus go_ddd.QueryBus) {
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
		if request.Method == http.MethodGet && e.queryTranslator != nil {
			query, err := e.queryTranslator(request)
			if err != nil {
				log.Printf("Error translating query: %v", err)
				writer.WriteHeader(http.StatusBadRequest)
				return
			}
			res, err := e.queryBus.Dispatch(request.Context(), query)
			if err != nil {
				log.Printf("Error dispatching query: %v", err)
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}
			// Write response
			writer.WriteHeader(http.StatusOK)
			err = json.NewEncoder(writer).Encode(res)
			if err != nil {
				log.Printf("Error encoding response: %v", err)
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else if e.commandTranslator != nil {
			command, err := e.commandTranslator(request)
			if err != nil {
				log.Printf("Error translating command: %v", err)
				writer.WriteHeader(http.StatusBadRequest)
				return
			}
			err = e.commandBus.Dispatch(request.Context(), command)
			if err != nil {
				log.Printf("Error dispatching command: %v", err)
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}
			writer.WriteHeader(http.StatusAccepted)
		} else {
			log.Printf("No translator for method %s on endpoint %s", request.Method, e.path)
			writer.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}
