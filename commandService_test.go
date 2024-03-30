package go_ddd

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCommandService(t *testing.T) {
	handler := func(context.Context, Command) error {
		return nil
	}
	service := NewCommandService(handler, testCommand{})
	assert.Equal(t, "github.com/paulvitic/ddd-go.testCommand", service.SubscribedTo()[0])

	ctx := context.Background()
	err := handler(ctx, NewCommand(testCommand{Name: "value"}))
	assert.NoError(t, err)
}
