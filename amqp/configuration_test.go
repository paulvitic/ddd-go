package amqp

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConnectionUrl(t *testing.T) {
	config := Configuration{
		Host:        "localhost",
		Port:        5672,
		Username:    "guest",
		Password:    "guest",
		VirtualHost: "vhost",
	}

	expected := "amqp://guest:guest@localhost:5672/vhost"
	assert.Equal(t, expected, connectionUrl(config))
}
