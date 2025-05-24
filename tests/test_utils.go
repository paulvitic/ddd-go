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

// Test endpoint implementation for Context tests
type TestEndpoint struct {
	path     string
	handlers map[ddd.HttpMethod]http.HandlerFunc
}

func (e *TestEndpoint) OnInit() {
	e.path = "test"
	e.handlers = map[ddd.HttpMethod]http.HandlerFunc{
		ddd.GET: func(w http.ResponseWriter, r *http.Request) {
			response := map[string]interface{}{
				"message": "Test endpoint GET request successful",
				"path":    e.path,
				"method":  "GET",
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
		},
	}
}

func (e *TestEndpoint) Path() string {
	return e.path
}

func (e *TestEndpoint) Handlers() map[ddd.HttpMethod]http.HandlerFunc {
	return e.handlers
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
