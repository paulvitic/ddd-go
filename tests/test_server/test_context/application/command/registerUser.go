package command

import (
	"github.com/paulvitic/ddd-go"
	"github.com/paulvitic/ddd-go/tests/test_server/test_context/domain/model"
)

type RegisterUser struct {
	userId ddd.ID
}

func NewRegisterUser(userId ddd.ID) *RegisterUser {
	return &RegisterUser{
		userId: userId,
	}
}

func (c *RegisterUser) Execute(ctx *ddd.Context) (any, error) {
	repo, err := ddd.Resolve[ddd.Repository[model.User]](ctx)
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
