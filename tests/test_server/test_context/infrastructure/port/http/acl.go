package http

import (
	"encoding/json"
	"net/http"

	"github.com/paulvitic/ddd-go"
	"github.com/paulvitic/ddd-go/tests/test_server/test_context/application/command"
	"github.com/paulvitic/ddd-go/tests/test_server/test_context/application/query"
)

func ToResigterUserCommand(r *http.Request) (*command.RegisterUser, error) {
	type RequestData struct {
		userId string
	}
	var data RequestData

	// Decode directly from request body
	decoder := json.NewDecoder(r.Body)
	// decoder.DisallowUnknownFields() // Optional: reject unknown fields

	if err := decoder.Decode(&data); err != nil {
		return nil, err
	}
	defer r.Body.Close()
	return command.NewRegisterUser(ddd.NewID(data.userId)), nil
}

func ToUserByIdQuery(r *http.Request) (ddd.Query, error) {
	type RequestData struct {
		userId string
	}
	var data RequestData
	// Decode directly from request body
	decoder := json.NewDecoder(r.Body)
	// decoder.DisallowUnknownFields() // Optional: reject unknown fields

	if err := decoder.Decode(&data); err != nil {
		return nil, err
	}
	defer r.Body.Close()

	filter := &query.UserById{
		UserId: data.userId,
	}
	return ddd.NewQuery(filter), nil

}
