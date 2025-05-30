package http

import (
	"encoding/json"
	"net/http"

	"github.com/paulvitic/ddd-go"
)

// TestEndpoint represents a test HTTP endpoint
type TestEndpoint struct {
	// You can inject other dependencies here if needed
	Logger *ddd.Logger `resource:""`
}

// NewTestEndpoint is the constructor function for TestEndpoint
func NewTestEndpoint() *TestEndpoint {
	return &TestEndpoint{}
}

// Path returns the endpoint's URL path - required by Endpoint interface
func (t *TestEndpoint) Path() string {
	return "/test"
}

// Post handles POST requests - discovered by method name convention
func (t *TestEndpoint) Post(w http.ResponseWriter, r *http.Request) {
	ctx := ddd.GetContext(r)
	ctx.Logger().Info("Test endpoint post method called")

	command, err := ToResigterUserCommand(r)
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
	}

	res, err := command.Execute(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

// Post handles POST requests - discovered by method name convention
func (t *TestEndpoint) GET(w http.ResponseWriter, r *http.Request) {
	ctx := ddd.GetContext(r)
	ctx.Logger().Info("Test endpoint get method called")

	query, err := ToUserByIdQuery(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
	}

	res, err := query.Respond(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res.Items())
}

// Delete handles DELETE requests - discovered by method name convention
func (t *TestEndpoint) Delete(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}
