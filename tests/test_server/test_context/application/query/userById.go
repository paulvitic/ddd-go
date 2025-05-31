package query

import (
	"github.com/paulvitic/ddd-go"
	"github.com/paulvitic/ddd-go/tests/test_server/test_context/domain/model"
)

type UserById struct {
	UserId string
}

func (u *UserById) Filter(ctx *ddd.Context) (ddd.QueryResponse, error) {
	users, err := ddd.Resolve[model.Users](ctx)
	if err != nil {
		return nil, err
	}
	res, err := users.ById(u.UserId)
	if err != nil {
		return nil, err
	}
	return ddd.NewQueryResponse(res), nil
}
