package mock

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

// IFace is an interface for testing.
type IFace interface {
	Call(value *Struct, args ...*reflect.Value) []any
}

// Struct is a non-interface for testing.
type Struct struct{}

// ErrAny is a special case of an none-named object.
var ErrAny = errors.New("argument failure")

func TestNothing(t *testing.T) {
	assert.Error(t, ErrAny)
}
