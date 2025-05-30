package model

// User represents a user in the system from the admin context perspective
type UserView struct {
	ID       string
	Email    string
	Name     string
	Role     string
	IsActive bool
}

// Users defines the contract for querying users
type Users interface {
	// ById retrieves a user by their ID
	ById(id string) (*UserView, error)
}
