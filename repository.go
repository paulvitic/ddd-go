package ddd

type Repository[T any] interface {
	// Save persists an aggregate
	Save(aggregate *T) error
	// Load retrieves an aggregate by its ID
	Load(id ID) (*T, error)
	// LoadAll retrieves all aggregates of the provided type
	LoadAll() ([]*T, error)
	// Delete removes an aggregate from the repository
	Delete(id ID) error
	// Update persists the changes made to an aggregate
	Update(*T) error
}
