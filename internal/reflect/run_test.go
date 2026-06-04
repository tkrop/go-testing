package reflect_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tkrop/go-testing/internal/reflect"
	"github.com/tkrop/go-testing/mock"
	"github.com/tkrop/go-testing/test"
)

// runner is a test double whose Run method delivers itself to the callback.
type runner struct {
	called bool
}

func (r *runner) Run(_ string, call func(*runner)) bool {
	call(r)

	return true
}

// failer is a test double that does not implement the expected type.
type failer struct{}

func (failer) Run(_ string, call func(failer)) bool {
	call(failer{})

	return true
}

type RunParams struct {
	setup      mock.SetupFunc
	target     any
	expectCall bool
}

var runTestCases = map[string]RunParams{
	"nil-target": {
		setup: test.Panic("call: target must not be nil"),
	},
	"no-run-method": {
		target: struct{}{},
		setup: test.Panic("call: target does not implement method " +
			"[struct {} => Run]"),
	},
	"not-type": {
		target: failer{},
		setup: test.Panic("call: target does not implement type " +
			"[reflect_test.failer => *reflect_test.runner]"),
	},

	"run": {
		target:     &runner{},
		expectCall: true,
	},
}

func TestRun(t *testing.T) {
	test.Map(t, runTestCases).
		Run(func(t test.Test, param RunParams) {
			// Given
			mock.NewMocks(t).Expect(param.setup)

			// When
			reflect.Run(param.target, "call", func(inner *runner) {
				inner.called = true
			})

			// Then
			assert.Equal(t, param.expectCall, param.target != nil &&
				test.Cast[*runner](param.target).called)
		})
}
