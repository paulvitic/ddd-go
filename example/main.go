package main

import (
	ddd "github.com/paulvitic/ddd-go/application"
	"github.com/paulvitic/ddd-go/example/hotel"
)

func main() {
	srv, _ := ddd.NewServer(nil)
	srv.WithContext(hotel.NewHotel())
	err := srv.Start()
	if err != nil {
		panic(err)
	}
}
