package adapter

import (
	"github.com/paulvitic/ddd-go"
	ddd_http "github.com/paulvitic/ddd-go/http"
	"net/http"
)

type roomEndpoint struct {
	ddd_http.Endpoint
}

func commandTranslator(from *http.Request) (go_ddd.Command, error) {
	return nil, nil
}

func queryTranslator(from *http.Request) (go_ddd.Query, error) {
	return nil, nil
}

func RoomEndpoint() ddd_http.Endpoint {
	return &roomEndpoint{
		Endpoint: ddd_http.NewEndpoint("/room").
			WithCommandTranslator(commandTranslator).
			WithQueryTranslator(queryTranslator),
	}
}
