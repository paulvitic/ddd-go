package http

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/paulvitic/ddd-go"
	"github.com/paulvitic/ddd-go/tests/test_server/test_context/application/command"
	"github.com/paulvitic/ddd-go/tests/test_server/test_context/application/query"
)

func ToResigterUserCommand(r *http.Request) (*command.RegisterUser, error) {
	type RequestData struct {
		UserId string
	}
	var data RequestData

	// Decode directly from request body
	decoder := json.NewDecoder(r.Body)
	// decoder.DisallowUnknownFields() // Optional: reject unknown fields

	if err := decoder.Decode(&data); err != nil {
		return nil, err
	}
	defer r.Body.Close()

	ctx := ddd.GetContext(r)
	return command.NewRegisterUser(ddd.NewID(data.UserId), ctx), nil
}

func ToUserByIdQuery(r *http.Request) (ddd.Query, error) {
	vars := mux.Vars(r)
	userID := vars["userId"]

	query := &query.UserById{
		UserId: userID,
	}
	return ddd.NewQuery(query.Filter), nil

}

func ToIdentityIdProviderEvent(r *http.Request) ddd.Event {

	event, err := ddd.EventFromJsonString("")
	if err != nil {
		panic("can not create event from json string")
	}
	return event
}
