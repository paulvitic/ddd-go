package payment

import (
	"github.com/paulvitic/ddd-go/context"
)

func Context() *context.Context {
	return context.NewContext(context.Configuration{Name: "payment"})
}
