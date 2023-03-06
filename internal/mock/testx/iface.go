// Package test contains an interface and types for testing the mock package
// loading and mock file generating from template. The interface and structs
// contain the minimal amount of features to cover all test aspects.
package test_test

import (
	"errors"
	"reflect"

	"github.com/tkrop/go-testing/internal/mock"
	"github.com/tkrop/go-testing/test"
)

// IFace is an interface for testing.
type IFace interface {
	CallA(value *Struct, args ...*reflect.Value) ([]any, error)
	//revive:disable-next-line // used for testing.
	CallB() (fn func([]*mock.File) []interface{}, err error)
	CallC(test test.Tester)
}

// Struct is a non-interface for testing.
type Struct struct{}

// ErrAny is a special case of an none-named object.
var ErrAny = errors.New("argument failure")
