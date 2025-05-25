package ddd_tests

import (
	"encoding/json"
	"net/http"

	"github.com/paulvitic/ddd-go"
)

type SomeStruct struct{}
type SomeDependencyInterface interface{}
type SomeDependencyStruct struct{}
type SomeStructRepo struct {
	logger                  ddd.Logger              `resource:""`
	someDependencyInterface SomeDependencyInterface `resource:""`
	someDependencyStruct    SomeDependencyStruct    `resource:"customDepenedencyName"`
	someDependencyPointer   *SomeDependencyStruct   `resource:""`
	nonResourceDependency   string
}

func (s SomeStructRepo) Save(aggregate *SomeStruct) error    { return nil }
func (s SomeStructRepo) Load(id ddd.ID) (*SomeStruct, error) { return nil, nil }
func (s SomeStructRepo) Delete(id ddd.ID) error              { return nil }
func (s SomeStructRepo) Update(aggregate *SomeStruct) error  { return nil }

type DatabaseConfig struct {
	ConnectionString string `json:"connectionString,omitempty"`
}

func (s *DatabaseConfig) OnInit() {
	config, err := ddd.Configuration[DatabaseConfig]("configs/properties.json")
	if err != nil {
		panic(err)
	}
	*s = *config
}

// TestEndpoint represents a test HTTP endpoint
type TestEndpoint struct {
	// You can inject other dependencies here if needed
	// Logger *ddd.Logger `resource:""`
}

// NewTestEndpoint is the constructor function for TestEndpoint
func NewTestEndpoint() *TestEndpoint {
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
