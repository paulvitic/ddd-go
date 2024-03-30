package go_ddd

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type testCommand struct {
	Name string
}

func TestCommand_Type(t *testing.T) {
	cmd := NewCommand(testCommand{Name: "value"})
	assert.Equal(t, "github.com/paulvitic/ddd-go.testCommand", cmd.Type())
	assert.Equal(t, "value", cmd.Body().(testCommand).Name)
}
