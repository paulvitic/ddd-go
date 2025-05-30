package file

import (
	"github.com/paulvitic/ddd-go/tests/test_server/test_context/domain/model"
)

type users struct {
}

func NewUsers() model.Users {
	return &users{}
}

func (u *users) ById(id string) (*model.UserView, error) {
	return &model.UserView{
		ID:       "1",
		Email:    "test_use@example.com",
		Name:     "Test User",
		Role:     "user",
		IsActive: true,
	}, nil
}
