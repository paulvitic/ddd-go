package ddd

type Command interface {
	Execute() (any, error)
}
