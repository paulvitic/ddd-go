package command

import (
	"github.com/paulvitic/ddd-go"
	"github.com/paulvitic/ddd-go/tests/test_server/test_context/domain/repository"
)

type RegisterUser struct {
	userId ddd.ID
	repo   repository.UserRepository
}

func NewRegisterUser(userId ddd.ID, ctx *ddd.Context) *RegisterUser {
	repo, err := ddd.Resolve[repository.UserRepository](ctx)
	if err != nil {
		panic("repo not found")
	}
	return &RegisterUser{
		userId: userId,
		repo:   repo,
	}
}

func (c *RegisterUser) Execute() (any, error) {
	user, err := c.repo.Load(c.userId)
	if err != nil {
		panic("can not find user")
	}
	user.Register()
	c.repo.Update(user)
	return user, nil
}
