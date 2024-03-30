package adapter

import (
	"github.com/paulvitic/ddd-go"
	dddhttp "github.com/paulvitic/ddd-go/http"
	"net/http"
)

type guestsEndpoint struct {
	dddhttp.Endpoint
}

func guestsQueryTranslator(from *http.Request) (go_ddd.Query, error) {
	return nil, nil
}

func GuestsEndpoint() dddhttp.Endpoint {
	return &guestsEndpoint{
		Endpoint: dddhttp.NewEndpoint("/guests").
			WithQueryTranslator(guestsQueryTranslator),
	}
}
