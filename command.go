package ddd

import (
	"reflect"
)

type Command interface {
	Type() string
	Body() interface{}
}

type command struct {
	body interface{}
}

func (c *command) Type() string {
	return reflect.TypeOf(c.Body()).PkgPath() + "." + reflect.TypeOf(c.Body()).Name()
}

func (c *command) Body() any {
	return c.body
}

func NewCommand(body interface{}) Command {
	return &command{body}
}

func CommandType(body interface{}) string {
	return reflect.TypeOf(body).PkgPath() + "." + reflect.TypeOf(body).Name()
}
