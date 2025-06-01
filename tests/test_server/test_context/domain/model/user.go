package model

import "github.com/paulvitic/ddd-go"

type User struct {
	ddd.Aggregate
	Email    string
	Name     string
	Role     string
	IsActive bool
}

func LoadUser(id ddd.ID) *User {
	user := &User{
		ddd.NewAggregate(id, User{}),
		"",
		"",
		"",
		true,
	}

	user.RaiseEvent(UserRegistered{})

	return user
}

func (u *User) Register() {
	u.RaiseEvent(UserRegistered{})
}

func (u *User) Approve() {
	u.RaiseEvent(UserApproved{})
}

func (u *User) Reject() {
	u.RaiseEvent(UserRejected{})
}
