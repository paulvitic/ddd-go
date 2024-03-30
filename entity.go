package go_ddd

type Entity interface {
	ID() ID
	Equals(other any) bool
}

// entity represents a base entity with an ID and properties.
type entity struct {
	id ID
}

// NewEntity creates a new Entity instance.
func NewEntity(id ID) Entity {
	return &entity{
		id: id,
	}
}

// ID returns the entity's ID.
func (e *entity) ID() ID {
	return e.id
}

// Equals checks if two entities are equal.
func (e *entity) Equals(other any) bool {
	if other == nil {
		return false
	}

	_, ok := other.(Entity)
	if ok {
		return e.ID().Equals(other.(Entity).ID())
	}
	return false
}
