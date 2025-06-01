package http

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/paulvitic/ddd-go"
	"github.com/paulvitic/ddd-go/tests/test_server/test_context/domain/model"
)

// UsersEndpoint represents a test HTTP endpoint
type UsersEndpoint struct {
	ddd.Endpoint
	// You can inject other dependencies here if needed
	//Logger *ddd.Logger `resource:""`
}

// NewTestEndpoint is the constructor function for TestEndpoint
func NewUsersEndpoint(logger *ddd.Logger, router *mux.Router) *UsersEndpoint {
	paths := []string{"/users/{userId}", "/users"}
	return &UsersEndpoint{
		Endpoint: ddd.NewEndpoint(&UsersEndpoint{}, paths, logger, router),
	}
}

// Post handles POST requests - discovered by method name convention
func (t *UsersEndpoint) Post(w http.ResponseWriter, r *http.Request) {

	command, err := ToResigterUserCommand(r)
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
	}

	res, err := command.Execute()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
	}

	response := make(map[string]any, 0)
	response["message"] = res.(*model.User).ID().String()

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// Post handles POST requests - discovered by method name convention
func (t *UsersEndpoint) Get(w http.ResponseWriter, r *http.Request) {
	ctx := ddd.GetContext(r)

	ctx.Logger().Info("get method called")
	query, err := ToUserByIdQuery(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
	}

	res, err := query.Filter(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res.Items())
}

// Delete handles DELETE requests - discovered by method name convention
func (t *UsersEndpoint) Delete(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}
