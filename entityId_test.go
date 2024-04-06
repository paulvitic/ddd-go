package ddd

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewID(t *testing.T) {
	// Test with valid string value
	stringID := NewID("test-string")
	assert.IsType(t, &entityId[string]{}, stringID, "NewID with string value did not create an entityId[string]")

	// Test with valid int value
	intID := NewID(123)
	assert.IsType(t, &entityId[int]{}, intID, "NewID with int value did not create an entityId[int]")

	identifier := NewValue(123)
	assert.Equal(t, 123, identifier.Value(), "Expected value to be 123")
	assert.Equal(t, "123", identifier.String(), "Expected string value to be 123")

	// Test with invalid value
	func() {
		defer func() {
			if r := recover(); r == nil {
				assert.Fail(t, "NewID with invalid value did not panic")
			}
		}()
		NewID(struct{}{})
	}()
}

func TestIDEquals(t *testing.T) {
	// Test with equal string IDs
	id1 := NewID("test-string")
	id2 := NewID("test-string")
	assert.True(t, id1.Equals(id2), "Equal string IDs did not return true for Equals")

	// Test with equal int IDs
	id3 := NewID(123)
	id4 := NewID(123)
	assert.True(t, id3.Equals(id4), "Equal int IDs did not return true for Equals")

	// Test with different string IDs
	id5 := NewID("different-string")
	assert.False(t, id1.Equals(id5), "Different string IDs returned true for Equals")

	// Test with different int IDs
	id6 := NewID(456)
	assert.False(t, id3.Equals(id6), "Different int IDs returned true for Equals")

	// Test with nil
	assert.False(t, id1.Equals(nil), "Equals returned true for nil value")

	// Test with non-ID type
	assert.False(t, id1.Equals(struct{}{}), "Equals returned true for non-ID type")
}
