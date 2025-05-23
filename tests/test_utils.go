package ddd_tests

import (
	"fmt"
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
	PathValue   string
	HandlersMap map[ddd.HttpMethod]http.HandlerFunc
}

func (e *TestEndpoint) Path() string {
	return e.PathValue
}

func (e *TestEndpoint) Handlers() map[ddd.HttpMethod]http.HandlerFunc {
	return e.HandlersMap
}

// Test interfaces
// type Logger interface {
// 	Log(message string)
// }

type Service interface {
	DoSomething() error
}

type Repository interface {
	FindById(id int) string
}

type DatabaseService interface {
	Connect() bool
}

type UserController interface {
	GetUser(id int) string
}

type Config interface {
	GetValue(key string) string
}

// Test implementations
type SimpleConfig struct {
	values map[string]string
}

func (c *SimpleConfig) GetValue(key string) string {
	return c.values[key]
}

// TestLogger has flags to check if hooks were called
type TestLogger struct {
	LogLevel      string `resource:"logLevel"`
	InitCalled    bool
	StartCalled   bool
	DestroyCalled bool
}

func (l *TestLogger) Log(message string) {
	// No-op
}

func (l *TestLogger) OnInit() error {
	l.InitCalled = true
	return nil
}

func (l *TestLogger) OnStart() error {
	l.StartCalled = true
	return nil
}

func (l *TestLogger) OnDestroy() error {
	l.DestroyCalled = true
	return nil
}

// ErrorLogger returns errors from lifecycle hooks
type ErrorLogger struct {
	LogLevel string `resource:"logLevel"`
}

func (l *ErrorLogger) Log(message string) {
	// No-op
}

func (l *ErrorLogger) OnInit() error {
	return fmt.Errorf("init error")
}

// NoHooksStruct doesn't implement any lifecycle hooks
type NoHooksStruct struct{}

func (s *NoHooksStruct) DoSomething() error { return nil }

type SimpleDatabaseService struct {
	ConnectionString string `resource:"connectionString"`
	Connected        bool
}

func (s *SimpleDatabaseService) Connect() bool {
	// Simulate connecting to a database
	s.Connected = true
	return true
}

func (s *SimpleDatabaseService) OnInit() error {
	// Validate connection string
	if s.ConnectionString == "" {
		return nil // Not an error for testing
	}
	return nil
}

type UserService struct {
	Logger     ddd.Logger `resource:"logger"`
	Repository Repository `resource:"userRepo"`
	ApiKey     string     `resource:"apiKey"`
}

func (s *UserService) DoSomething() error {
	s.Logger.Info("Doing something with repository")
	s.Repository.FindById(1)
	return nil
}

func (s *UserService) OnInit() error {
	if s.ApiKey == "" {
		return fmt.Errorf("ApiKey cannot be empty")
	}
	return nil
}

func (s *UserService) OnStart() error {
	return nil
}

type UserRepository struct{}

func (r *UserRepository) FindById(id int) string {
	return fmt.Sprintf("User with ID %d", id)
}

func (r *UserRepository) OnInit() error {
	return nil
}

type SimpleUserController struct {
	Logger        ddd.Logger      `resource:"logger"`
	DbService     DatabaseService `resource:"dbService"`
	Repository    Repository      `resource:"userRepo"`
	InitCalled    bool
	StartCalled   bool
	DestroyCalled bool
}

func (c *SimpleUserController) GetUser(id int) string {
	c.Logger.Info("Getting user " + c.Repository.FindById(id))
	return "User " + fmt.Sprint(id)
}

func (c *SimpleUserController) OnInit() error {
	c.InitCalled = true
	return nil
}

func (c *SimpleUserController) OnStart() error {
	c.StartCalled = true
	c.DbService.Connect()
	return nil
}

func (c *SimpleUserController) OnDestroy() error {
	c.DestroyCalled = true
	return nil
}
