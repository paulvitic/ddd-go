package main

import (
	ddd "github.com/paulvitic/ddd-go/application"
	"github.com/paulvitic/ddd-go/example/auth"
	"github.com/paulvitic/ddd-go/example/hotel"
	"github.com/paulvitic/ddd-go/example/mobile"
	"github.com/paulvitic/ddd-go/example/payment"
)

func main() {
	if err := ddd.NewServer(nil).
		WithContext(hotel.Context()).
		WithContext(auth.Context()).
		WithContext(mobile.Context()).
		WithContext(payment.Context()).
		Start(); err != nil {
		panic(err)
	}
}
