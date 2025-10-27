//revive:disable-next-line:max-public-structs // structs for testing
package reflect_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tkrop/go-testing/internal/reflect"
	"github.com/tkrop/go-testing/test"
)

// CustomStringer is a test type that implements fmt.Stringer.
type CustomStringer struct {
	value string
}

func (c CustomStringer) String() string {
	return c.value
}

// isGomock is a marker struct for generated mocks.
type IsGomock struct{}

// MockStruct simulates a generated mock with isgomock field.
type MockStruct struct {
	isgomock IsGomock
	value    string
}

func (m MockStruct) String() string {
	return "mock-string-" + m.value
}

// StringArgsParam represents unified test parameters for all StringArgs related functions.
type StringArgsParam struct {
	args   []any
	expect []string
}

var stringArgsTestCases = map[string]StringArgsParam{
	// Testing single value conversion behavior.
	"nil-value": {
		args:   []any{nil},
		expect: []string{"<nil>"},
	},
	"string-value": {
		args:   []any{"hello"},
		expect: []string{"hello"},
	},
	"integer-value": {
		args:   []any{42},
		expect: []string{"42"},
	},
	"float-value": {
		args:   []any{3.14},
		expect: []string{"3.14"},
	},
	"boolean-true": {
		args:   []any{true},
		expect: []string{"true"},
	},
	"boolean-false": {
		args:   []any{false},
		expect: []string{"false"},
	},
	"slice-value": {
		args:   []any{[]int{1, 2}},
		expect: []string{"[1 2]"},
	},
	"empty-slice": {
		args:   []any{[]int{}},
		expect: []string{"[]"},
	},

	// Testing multi-value argument conversion.
	"empty-args": {
		args:   []any{},
		expect: []string{},
	},
	"string-args": {
		args:   []any{"hello", "world"},
		expect: []string{"hello", "world"},
	},
	"integer-args": {
		args:   []any{1, 42, -10},
		expect: []string{"1", "42", "-10"},
	},
	"float-args": {
		args:   []any{3.14, -2.5},
		expect: []string{"3.14", "-2.5"},
	},
	"boolean-args": {
		args:   []any{true, false},
		expect: []string{"true", "false"},
	},
	"mixed-types": {
		args:   []any{"test", 123, 3.14, true, nil},
		expect: []string{"test", "123", "3.14", "true", "<nil>"},
	},
	"slice-args": {
		args:   []any{[]int{1, 2, 3}, []string{"a", "b"}},
		expect: []string{"[1 2 3]", "[a b]"},
	},
	"map-args": {
		args:   []any{map[string]int{"key": 42}},
		expect: []string{"map[key:42]"},
	},
	"struct-args": {
		args:   []any{struct{ Name string }{Name: "test"}},
		expect: []string{"{test}"},
	},

	// Testing custom `String` method behavior.
	"custom-stringer": {
		args:   []any{CustomStringer{value: "custom-value"}},
		expect: []string{"custom-value"},
	},
	"custom-stringer-single": {
		args:   []any{CustomStringer{value: "custom"}},
		expect: []string{"custom"},
	},

	// Testing mock detection and type name return.
	"mock-struct": {
		args:   []any{MockStruct{value: "test", isgomock: IsGomock{}}},
		expect: []string{"reflect_test.MockStruct"},
	},
	"mock-struct-pointer": {
		args:   []any{&MockStruct{value: "test", isgomock: IsGomock{}}},
		expect: []string{"*reflect_test.MockStruct"},
	},

	// Testing normal structs to be not treated as mocks.
	"custom-struct": {
		args:   []any{CustomStringer{value: "test"}},
		expect: []string{"test"},
	},
	"custom-struct-pointer": {
		args:   []any{&CustomStringer{value: "test"}},
		expect: []string{"test"},
	},
	"interface-value": {
		args:   []any{fmt.Stringer(CustomStringer{value: "test"})},
		expect: []string{"test"},
	},
}

func TestStringArgs(t *testing.T) {
	test.Map(t, stringArgsTestCases).
		Run(func(t test.Test, param StringArgsParam) {
			// When
			result := reflect.StringArgs(param.args)

			// Then
			assert.Equal(t, param.expect, result)
		})
}
