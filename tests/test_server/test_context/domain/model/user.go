package model

import "github.com/paulvitic/ddd-go"

type User struct {
	ddd.Aggregate
	Email    string
	Name     string
	Role     string
	IsActive bool
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
