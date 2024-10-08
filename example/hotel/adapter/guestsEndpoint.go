package adapter

import (
	"github.com/paulvitic/ddd-go"
	dddhttp "github.com/paulvitic/ddd-go/http"
	"net/http"
)

func guestsQueryTranslator(from *http.Request) (ddd.Query, error) {
	return nil, nil
}

func GuestsEndpoint() *dddhttp.QueryEndpoint {
	return dddhttp.NewQueryEndpoint("guests", []string{}, guestsQueryTranslator)
}
