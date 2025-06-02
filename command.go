package ddd

// TODO you can try to convert Command into request scoped context resource
type Command interface {
	Execute() (any, error)
}
