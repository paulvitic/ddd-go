package ddd

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValueObject_Equals(t *testing.T) {
	deepValueObject1 := NewValue(map[string]interface{}{
		"key": "value",
		"key2": map[string]interface{}{
			"key3": "value3",
		},
	})
	deepValueObject2 := NewValue(map[string]interface{}{
		"key": "value",
		"key2": map[string]interface{}{
			"key3": "value3",
		},
	})
	assert.True(t, deepValueObject1.Equals(deepValueObject2))
}

func TestValueObject_NotEquals(t *testing.T) {
	deepValueObject1 := NewValue(map[string]interface{}{
		"key": "value",
		"key2": map[string]interface{}{
			"key3": "value3",
		},
	})
	deepValueObject2 := NewValue(map[string]interface{}{
		"key": "value",
		"key2": map[string]interface{}{
			"key3": "value4",
		},
	})
	assert.False(t, deepValueObject1.Equals(deepValueObject2))
}

func TestValueObject_ToString(t *testing.T) {
	deepValueObject := NewValue(map[string]interface{}{
		"key": "value",
		"key2": map[string]interface{}{
			"key3": "value3",
		},
	})
	assert.Equal(t, "map[key:value key2:map[key3:value3]]", deepValueObject.String())
}
