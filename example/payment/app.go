package payment

import (
	ddd "github.com/paulvitic/ddd-go/application"
)

func Context() *ddd.Context {
	return ddd.NewContext("payment")
}
