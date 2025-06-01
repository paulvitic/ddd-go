package test_context

import (
	"context"

	"github.com/gorilla/mux"
	"github.com/paulvitic/ddd-go"
	"github.com/paulvitic/ddd-go/tests/test_server/test_context/application/process"
	"github.com/paulvitic/ddd-go/tests/test_server/test_context/infrastructure/adapter/file"
	"github.com/paulvitic/ddd-go/tests/test_server/test_context/infrastructure/port/http"
)

func TestContext(ctx context.Context, router *mux.Router) *ddd.Context {
	return ddd.NewContext(ctx, router, "test").
		WithResources(
			ddd.Resource(http.NewUsersEndpoint),
			ddd.Resource(file.NewFilePersitenceConfig),
			ddd.Resource(file.NewUsersView),
			ddd.Resource(file.NewUserRepository),
			ddd.Resource(process.UserProcessor, "userProcessor"),
		)
}
