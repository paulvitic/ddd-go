package ddd

import (
	"reflect"
)

type Command interface {
	Type() string
	Body() any
}

type command struct {
	body any
}

func (c *command) Type() string {
	return reflect.TypeOf(c.Body()).PkgPath() + "." + reflect.TypeOf(c.Body()).Name()
}

func (c *command) Body() any {
	return c.body
}

func NewCommand(body any) Command {
	return &command{body}
}

func CommandType(body any) string {
	return reflect.TypeOf(body).PkgPath() + "." + reflect.TypeOf(body).Name()
}
