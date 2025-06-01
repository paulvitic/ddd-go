package model

import (
	"fmt"
	"time"

	"github.com/paulvitic/ddd-go"
)

type User struct {
	ddd.Aggregate
	logger   *ddd.Logger
	Email    string
	Name     string
	Role     string
	IsActive bool
}

func LoadUser(id ddd.ID) *User {
	user := &User{
		ddd.NewAggregate(id, User{}),
		ddd.NewLogger(),
		"",
		"",
		"",
		true,
	}

	return user
}

func (u *User) Register() {
	u.RaiseEvent(UserRegistered{
		ProcessingID: fmt.Sprintf("proc-%d", time.Now().UnixNano()),
	})
	u.logger.Info("registered user %s", u.ID().String())
}

func (u *User) Approve() {
	u.RaiseEvent(UserApproved{})
	u.logger.Info("approved user %s", u.ID().String())
}

func (u *User) Reject() {
	u.RaiseEvent(UserRejected{})
}
