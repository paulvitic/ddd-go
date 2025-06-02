package test_server

import (
	"github.com/paulvitic/ddd-go"
	"github.com/paulvitic/ddd-go/tests/test_server/test_context"
)

func TestServer() *ddd.Server {
	return ddd.NewServer(ddd.NewServerConfig()).
		WithContexts(test_context.TestContext)
}
