package http

import (
	"encoding/json"
	"log"
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
	return command.NewRegisterUser(ddd.NewID(data.UserId)), nil
}

func ToUserByIdQuery(r *http.Request) (ddd.Query, error) {
	// type RequestData struct {
	// 	UserId string
	// }
	// var data RequestData
	// // Decode directly from request body
	// decoder := json.NewDecoder(r.Body)
	// // decoder.DisallowUnknownFields() // Optional: reject unknown fields

	// if err := decoder.Decode(&data); err != nil {
	// 	return nil, err
	// }
	// defer r.Body.Close()

	vars := mux.Vars(r)

	response := map[string]interface{}{
		"url":  r.URL.String(),
		"path": r.URL.Path,
		"vars": vars,
	}

	log.Print(response)

	userID := vars["userId"]
	// userID := r.URL.Query().Get("userId")

	query := &query.UserById{
		UserId: userID,
	}
	return ddd.NewQuery(query.Filter), nil

}
