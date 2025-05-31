package test_context

import (
	"context"

	"github.com/paulvitic/ddd-go"
	"github.com/paulvitic/ddd-go/tests/test_server/test_context/application/process"
	"github.com/paulvitic/ddd-go/tests/test_server/test_context/infrastructure/adapter/file"
	"github.com/paulvitic/ddd-go/tests/test_server/test_context/infrastructure/port/http"
)

func TestContext(ctx context.Context) *ddd.Context {
	return ddd.NewContext(ctx, "test").
		WithResources(
			ddd.Resource(http.NewUsersEndpoint),
			ddd.Resource(file.NewFilePersitenceConfig),
			ddd.Resource(file.NewUsers),
			ddd.Resource(file.NewUserRepository),
			ddd.Resource(process.UserProcessor, "userProcessor"),
		)
}
