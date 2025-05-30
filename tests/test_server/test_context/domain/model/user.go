package model

import "github.com/paulvitic/ddd-go"

type User struct {
	ddd.Aggregate
}

func (u *User) Register() {
	u.RaiseEvent(
		u.AggregateType(),
		u.ID(),
		UserRegistered{})
}
