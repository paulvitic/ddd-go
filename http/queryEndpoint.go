package http

import (
	"encoding/json"
	"github.com/paulvitic/ddd-go"
	"log"
	"net/http"
)

type QueryEndpoint struct {
	*EndpointBase
	translator QueryTranslator
	queryBus   ddd.QueryBus
}

func NewQueryEndpoint(path string, methods []string, translator QueryTranslator) *QueryEndpoint {
	return &QueryEndpoint{
		EndpointBase: NewEndpoint(path, methods),
		translator:   translator,
		queryBus:     nil,
	}
}

func (e *QueryEndpoint) RegisterQueryBus(bus ddd.QueryBus) {
	if e.queryBus != nil {
		log.Printf("Query bus already set for endpoint %s", e.path)
	}
	e.queryBus = bus
}

func (e *QueryEndpoint) Handler() func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		if e.translator == nil {
			http.Error(writer, "No translator for GET method", http.StatusNotAcceptable)
			return
		}

		query, err := e.translator(request)
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
}
