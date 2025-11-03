package reflect

import (
	"math/rand"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewPrimitiveDefault tests the defensive default case in newPrimitive.
// This case should never be reached in normal operation, but we test it for coverage.
func TestNewPrimitiveDefault(t *testing.T) {
	// Given
	r := &random{
		rand:   rand.New(rand.NewSource(42)),
		size:   5,
		length: 20,
	}

	// When - call with an unsupported kind (not a primitive)
	result := r.newPrimitive(reflect.Array)

	// Then - should return nil for unsupported kinds
	assert.Nil(t, result)
}
