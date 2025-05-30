package ddd_tests

import (
	"encoding/json"
	"net/http"

	"github.com/paulvitic/ddd-go"
)

// ======================================
// Configuration
// ======================================
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

// ==================================
// Aggregate
// ==================================
type User struct {
	ddd.Aggregate
}

func (u *User) Register() {
	return
}

// ====================================
// Command
// ====================================
type registerUser struct {
	userId ddd.ID
}

func RegisterUser(userId ddd.ID) *registerUser {
	return &registerUser{
		userId: userId,
	}
}

func (c *registerUser) Execute(ctx *ddd.Context) (any, error) {
	repo, err := ddd.Resolve[ddd.Repository[User]](ctx)
	if err != nil {
		panic("repo not found")
	}
	user, err := repo.Load(c.userId)
	if err != nil {
		panic("can not find user")
	}
	user.Register()
	repo.Update(user)
	return user, nil
}

// ===================================
// Endpoint
// ===================================
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

// acl
func ToResigterUserComand(r *http.Request) (*registerUser, error) {
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
	return RegisterUser(ddd.NewID(data.userId)), nil
}

// Post handles POST requests - discovered by method name convention
func (t *TestEndpoint) Post(w http.ResponseWriter, r *http.Request) {
	ctx := ddd.GetContext(r)
	ctx.Logger().Info("Test endpoint post method called")

	command, err := ToResigterUserComand(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
	}

	res, err := command.Execute(ctx)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

// Delete handles DELETE requests - discovered by method name convention
func (t *TestEndpoint) Delete(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

type Service interface {
	DoSomething() error
}

type DatabaseService interface {
	Connect() bool
}
