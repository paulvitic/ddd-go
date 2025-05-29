package ddd_tests

import (
	"encoding/json"
	"net/http"

	"github.com/paulvitic/ddd-go"
)

type DatabaseConfig struct {
	ConnectionString string `json:"connectionString"`
}

func (c *DatabaseConfig) OnInit() {
	config, err := ddd.Configuration[DatabaseConfig]("configs/properties.json")
	if err != nil {
		panic(err)
	}
	*c = *config
}

// TestEndpoint represents a test HTTP endpoint
type TestEndpoint struct {
	// You can inject other dependencies here if needed
	Logger *ddd.Logger `resource:""`
}

// NewTestEndpoint is the constructor function for TestEndpoint
func NewTestEndpoint(commandBus *ddd.CommandBus) *TestEndpoint {
	return &TestEndpoint{}
}

// Path returns the endpoint's URL path - required by Endpoint interface
func (t *TestEndpoint) Path() string {
	return "/test"
}

// Get handles GET requests - discovered by method name convention
func (t *TestEndpoint) Get(w http.ResponseWriter, r *http.Request) {

	response := map[string]interface{}{
		"message": "GET request handled successfully",
		"path":    "/test",
		"method":  "GET",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Post handles POST requests - discovered by method name convention
func (t *TestEndpoint) Post(w http.ResponseWriter, r *http.Request) {
	t.Logger.Info("Test endpoint post method called")
	response := map[string]interface{}{
		"message": "POST request handled successfully",
		"path":    "/test",
		"method":  "POST",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// Put handles PUT requests - discovered by method name convention
func (t *TestEndpoint) Put(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"message": "PUT request handled successfully",
		"path":    "/test",
		"method":  "PUT",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Delete handles DELETE requests - discovered by method name convention
func (t *TestEndpoint) Delete(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

type Service interface {
	DoSomething() error
}

type Repository interface {
	FindById(id int) string
}

type DatabaseService interface {
	Connect() bool
}
