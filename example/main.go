package main

import (
	"github.com/paulvitic/ddd-go/config"
	"github.com/paulvitic/ddd-go/example/auth"
	"github.com/paulvitic/ddd-go/example/hotel"
	"github.com/paulvitic/ddd-go/example/mobile"
	"github.com/paulvitic/ddd-go/example/payment"
	"github.com/paulvitic/ddd-go/server"
)

func main() {
	if err := server.NewServer(config.Properties[server.Configuration]()).
		WithContext(hotel.Context()).
		WithContext(auth.Context()).
		WithContext(mobile.Context()).
		WithContext(payment.Context()).
		Start(); err != nil {
		panic(err)
	}
}
