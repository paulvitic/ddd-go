package ddd

type Command interface {
	Execute(ctx *Context) (any, error)
}
