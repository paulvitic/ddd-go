package repository

import (
	"github.com/paulvitic/ddd-go"
	"github.com/paulvitic/ddd-go/tests/test_server/test_context/domain/model"
)

type UserRepository interface {
	ddd.Repository[model.User]
}
