package adapter

import (
	"github.com/paulvitic/ddd-go"
	dddhttp "github.com/paulvitic/ddd-go/http"
	"net/http"
)

func commandTranslator(from *http.Request) (ddd.Command, error) {
	return nil, nil
}

func RoomEndpoint() *dddhttp.CommandEndpoint {
	return dddhttp.NewCommandEndpoint("room", []string{}, commandTranslator)
}
