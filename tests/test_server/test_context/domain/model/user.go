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

	user.RaiseEvent(
		user.AggregateType(),
		user.ID(),
		UserRegistered{})

	return user
}

func (u *User) Register() {
	u.RaiseEvent(
		u.AggregateType(),
		u.ID(),
		UserRegistered{})
}

func (u *User) Approve() {
	u.RaiseEvent(
		u.AggregateType(),
		u.ID(),
		UserApproved{})
}

func (u *User) Reject() {
	u.RaiseEvent(
		u.AggregateType(),
		u.ID(),
		UserRejected{})
}
