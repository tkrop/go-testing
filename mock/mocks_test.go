package mock_test

import (
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/tkrop/go-testing/internal/reflect"

	"github.com/tkrop/go-testing/mock"
	"github.com/tkrop/go-testing/perm"
	"github.com/tkrop/go-testing/test"
)

type IFace interface {
	CallA(input string)
	CallB(input string) string
}

func CallA(input string) mock.SetupFunc {
	return func(mocks *mock.Mocks) any {
		return mock.Get(mocks, NewMockIFace).EXPECT().
			CallA(mocks.Equal(input)).Do(mocks.Do(IFace.CallA))
	}
}

func CallB(input string, output string) mock.SetupFunc {
	return func(mocks *mock.Mocks) any {
		return mock.Get(mocks, NewMockIFace).EXPECT().
			CallB(mocks.Equal(input)).DoAndReturn(
			mocks.Call(IFace.CallB, func(...any) []any {
				return []any{output}
			}))
	}
}

// func CallBX(input string, output string) mock.SetupFunc {
// 	return mock.Mock(NewMockIFace, func(mock *MockIFace) *gomock.Call {
// 		return mock.EXPECT().CallB(input).Return(output)
// 	})
// }

func NoCall() mock.SetupFunc {
	return func(mocks *mock.Mocks) any {
		return mock.Get(mocks, NewMockIFace).EXPECT()
	}
}

type XFace interface {
	CallC(input any)
}

func CallC(input any) mock.SetupFunc {
	return func(mocks *mock.Mocks) any {
		return mock.Get(mocks, NewMockXFace).EXPECT().
			CallC(mocks.Equal(input)).Do(mocks.Do(XFace.CallC))
	}
}

// Generic source directory for caller path evaluation.
var SourceDir = test.Must(os.Getwd())

type MockParams struct {
	setup  mock.SetupFunc
	misses func(test.Test, *mock.Mocks) mock.SetupFunc
	call   func(test.Test, *mock.Mocks)
}

var mockTestCAses = map[string]MockParams{
	"single mock with single call": {
		setup: mock.Setup(
			CallA("ok"),
		),
		call: func(_ test.Test, mocks *mock.Mocks) {
			mock.Get(mocks, NewMockIFace).CallA("ok")
		},
	},
	"single mock with two calls": {
		setup: mock.Setup(
			CallA("ok"), CallA("okay"),
		),
		call: func(_ test.Test, mocks *mock.Mocks) {
			mock.Get(mocks, NewMockIFace).CallA("ok")
			mock.Get(mocks, NewMockIFace).CallA("okay")
		},
	},
	"single mock with missing calls": {
		setup: mock.Setup(
			CallA("ok"), CallA("okay"),
		),
		misses: test.MissingCalls(CallA("okay")),
		call: func(_ test.Test, mocks *mock.Mocks) {
			mock.Get(mocks, NewMockIFace).CallA("ok")
		},
	},
	"single mock with unexpected call": {
		misses: test.UnexpectedCall(NewMockIFace,
			"CallA", path.Join(SourceDir, "mocks_test.go:105"), "ok"),
		call: func(_ test.Test, mocks *mock.Mocks) {
			mock.Get(mocks, NewMockIFace).CallA("ok")
		},
	},
	"single mock with more than expected calls": {
		setup: mock.Setup(
			CallA("ok"),
		),
		misses: test.ConsumedCall(NewMockIFace,
			"CallA", path.Join(SourceDir, "mocks_test.go:117"),
			path.Join(SourceDir, "mocks_test.go:28"), "ok"),
		call: func(_ test.Test, mocks *mock.Mocks) {
			mock.Get(mocks, NewMockIFace).CallA("ok")
			mock.Get(mocks, NewMockIFace).CallA("ok")
		},
	},

	"single mock with many calls": {
		setup: mock.Setup(
			CallA("okay"),
			CallB("okay", "okay"),
		),
		call: func(_ test.Test, mocks *mock.Mocks) {
			mock.Get(mocks, NewMockIFace).CallA("okay")
			mock.Get(mocks, NewMockIFace).CallB("okay")
		},
	},
	"multiple mocks with many calls": {
		setup: mock.Setup(
			CallA("okay"),
			CallB("okay", "okay"),
			CallC("okay"),
		),
		call: func(_ test.Test, mocks *mock.Mocks) {
			mock.Get(mocks, NewMockIFace).CallA("okay")
			mock.Get(mocks, NewMockIFace).CallB("okay")
			mock.Get(mocks, NewMockXFace).CallC("okay")
		},
	},
}

func TestMocks(t *testing.T) {
	test.Map(t, mockTestCAses).
		Run(func(t test.Test, param MockParams) {
			// Given
			mocks := mock.NewMocks(t)

			// When
			test.InRun(test.Success, func(tt test.Test) {
				// Given
				imocks := mock.NewMocks(tt)
				if param.misses != nil {
					mocks.Expect(param.misses(tt, imocks))
				}
				imocks.Expect(param.setup)

				// Connect the mock controller directly to the isolated parent
				// test environment to capture the mock controller failure.
				imocks.Ctrl.T = t

				// When
				param.call(tt, imocks)
			})(t)
		})
}

func TestMockArgs(t *testing.T) {
	// Given
	mocks := mock.NewMocks(t)

	// When
	mocks.SetArg("a", "a")

	// Than
	assert.Equal(t, mocks.GetArg("a"), "a")
	assert.Equal(t, mocks.GetArg("b"), nil)

	// When
	mocks.SetArgs(map[any]any{
		"a": "b",
		"b": "b",
	})

	// Than
	assert.Equal(t, mocks.GetArg("a"), "b")
	assert.Equal(t, mocks.GetArg("b"), "b")
}

func MockSetup(t gomock.TestReporter, setup mock.SetupFunc) *mock.Mocks {
	return mock.NewMocks(t).Expect(setup)
}

func MockValidate(
	t test.Test, mocks *mock.Mocks,
	validate func(test.Test, *mock.Mocks),
	failing bool,
) {
	if failing {
		// we need to execute failing test synchronous, since we setup full
		// permutations instead of stopping setup on first failing mock calls.
		validate(t, mocks)
	} else {
		// Test proper usage of `WaitGroup` on non-failing validation.
		validate(t, mocks)
		mocks.Wait()
	}
}

func SetupPermTestABC(mocks *mock.Mocks) *perm.Test {
	iface := mock.Get(mocks, NewMockIFace)
	return perm.NewTest(mocks,
		perm.TestMap{
			"a": func(test.Test) { iface.CallA("a") },
			"b1": func(t test.Test) {
				assert.Equal(t, "c", iface.CallB("b"))
			},
			"b2": func(t test.Test) {
				assert.Equal(t, "d", iface.CallB("b"))
			},
			"c": func(test.Test) { iface.CallA("c") },
		})
}

func SetupPermTestABCD(mocks *mock.Mocks) *perm.Test {
	iface := mock.Get(mocks, NewMockIFace)
	return perm.NewTest(mocks,
		perm.TestMap{
			"a": func(test.Test) { iface.CallA("a") },
			"b": func(test.Test) { iface.CallA("b") },
			"c": func(t test.Test) {
				assert.Equal(t, "d", iface.CallB("c"))
			},
			"d": func(t test.Test) {
				assert.Equal(t, "e", iface.CallB("d"))
			},
		})
}

func SetupPermTestABCDEF(mocks *mock.Mocks) *perm.Test {
	iface := mock.Get(mocks, NewMockIFace)
	return perm.NewTest(mocks,
		perm.TestMap{
			"a": func(test.Test) { iface.CallA("a") },
			"b": func(test.Test) { iface.CallA("b") },
			"c": func(t test.Test) {
				assert.Equal(t, "d", iface.CallB("c"))
			},
			"d": func(t test.Test) {
				assert.Equal(t, "e", iface.CallB("d"))
			},
			"e": func(test.Test) { iface.CallA("e") },
			"f": func(test.Test) { iface.CallA("f") },
		})
}

var setupTestCases = perm.ExpectMap{
	"b2-b1-a-c": test.Failure,
	"b2-b1-c-a": test.Failure,
	"b2-c-b1-a": test.Failure,
	"b2-a-b1-c": test.Failure,
	"b2-c-a-b1": test.Failure,
	"b2-a-c-b1": test.Failure,
	"a-b2-b1-c": test.Failure,
	"c-b2-b1-a": test.Failure,
	"a-b2-c-b1": test.Failure,
	"c-b2-a-b1": test.Failure,
	"c-a-b2-b1": test.Failure,
	"a-c-b2-b1": test.Failure,
}

func TestSetup(t *testing.T) {
	perms := setupTestCases.Remain(test.Success)
	test.Map(t, perms).Run(func(t test.Test, expect test.Expect) {
		// Given
		name := strings.Split(t.Name(), "/")[1]
		perm := strings.Split(name, "-")
		setup := mock.Setup(
			CallA("a"),
			mock.Setup(
				CallB("b", "c"),
				CallB("b", "d"),
			),
			CallA("c"),
		)
		mock := MockSetup(t, setup)

		// When
		test := SetupPermTestABC(mock)

		// Then
		test.Test(t, perm, expect)
	})
}

var chainTestCases = perm.ExpectMap{
	"a-b1-b2-c": test.Success,
}

func TestChain(t *testing.T) {
	perms := chainTestCases.Remain(test.Failure)
	test.Map(t, perms).Run(func(t test.Test, expect test.Expect) {
		// Given
		name := strings.Split(t.Name(), "/")[1]
		perm := strings.Split(name, "-")
		setup := mock.Chain(
			CallA("a"),
			mock.Chain(
				CallB("b", "c"),
				CallB("b", "d"),
			),
			CallA("c"),
		)
		mock := MockSetup(t, setup)

		// When
		test := SetupPermTestABC(mock)

		// Then
		test.Test(t, perm, expect)
	})
}

var setupChainTestCases = perm.ExpectMap{
	"a-b-c-d": test.Success,
	"a-c-b-d": test.Success,
	"a-c-d-b": test.Success,
	"c-a-b-d": test.Success,
	"c-a-d-b": test.Success,
	"c-d-a-b": test.Success,
}

func TestSetupChain(t *testing.T) {
	perms := setupChainTestCases.Remain(test.Failure)
	test.Map(t, perms).Run(func(t test.Test, expect test.Expect) {
		// Given
		name := strings.Split(t.Name(), "/")[1]
		perm := strings.Split(name, "-")

		// Basic setup of two independent chains.
		setup := mock.Setup(
			mock.Chain(
				CallA("a"),
				CallA("b"),
			),
			mock.Chain(
				CallB("c", "d"),
				CallB("d", "e"),
			),
		)
		mock := MockSetup(t, setup)

		// When
		test := SetupPermTestABCD(mock)

		// Then
		test.Test(t, perm, expect)
	})
}

func TestChainSetup(t *testing.T) {
	perms := setupChainTestCases.Remain(test.Failure)
	test.Map(t, perms).Run(func(t test.Test, expect test.Expect) {
		// Given
		name := strings.Split(t.Name(), "/")[1]
		perm := strings.Split(name, "-")

		// Frail setup to detach a (sub-)chain.
		setup := mock.Chain(
			CallA("a"),
			CallA("b"),
			mock.Setup( // detaching (sub-)chain.
				mock.Chain(
					CallB("c", "d"),
					CallB("d", "e"),
				),
			),
		)
		mock := MockSetup(t, setup)

		// When
		test := SetupPermTestABCD(mock)

		// Then
		test.Test(t, perm, expect)
	})
}

var parallelChainTestCases = perm.ExpectMap{
	"a-b-c-d-e-f": test.Success,
	"a-b-c-e-d-f": test.Success,
	"a-b-e-c-d-f": test.Success,
	"a-c-b-d-e-f": test.Success,
	"a-c-d-b-e-f": test.Success,
	"a-c-d-e-b-f": test.Success,
	"a-c-b-e-d-f": test.Success,
	"a-c-e-d-b-f": test.Success,
	"a-c-e-b-d-f": test.Success,
	"a-e-b-c-d-f": test.Success,
	"a-e-c-b-d-f": test.Success,
	"a-e-c-d-b-f": test.Success,
}

func TestParallelChain(t *testing.T) {
	perms := parallelChainTestCases.Remain(test.Failure)
	test.Map(t, perms).Run(func(t test.Test, expect test.Expect) {
		// Given
		name := strings.Split(t.Name(), "/")[1]
		perm := strings.Split(name, "-")
		setup := mock.Chain(
			CallA("a"),
			mock.Parallel(
				CallA("b"),
				mock.Chain(
					CallB("c", "d"),
					CallB("d", "e"),
				),
				mock.Parallel(
					CallA("e"),
				),
			),
			CallA("f"),
		)
		mock := MockSetup(t, setup)

		// When
		test := SetupPermTestABCDEF(mock)

		// Then
		test.Test(t, perm, expect)
	})
}

var chainSubTestCases = perm.ExpectMap{
	"a-b-c-d-e-f": test.Success,
	"a-c-b-d-e-f": test.Success,
	"a-c-d-b-e-f": test.Success,
	"a-c-d-e-b-f": test.Success,
	"f-a-b-c-d-e": test.Success,
	"a-f-b-c-d-e": test.Success,
	"a-b-f-c-d-e": test.Success,
	"a-b-c-f-d-e": test.Success,
	"a-b-c-d-f-e": test.Success,
	"a-c-d-e-f-b": test.Success,

	"b-a-c-d-e-f": test.Failure,
	"c-a-b-d-e-f": test.Failure,
	"d-a-b-c-e-f": test.Failure,
	"a-b-c-e-d-f": test.Failure,
	"a-b-d-e-c-f": test.Failure,
}

func TestChainSub(t *testing.T) {
	perms := chainSubTestCases
	// perms := chainSubTestCases.Remain(test.ExpectSuccess)
	test.Map(t, perms).Run(func(t test.Test, expect test.Expect) {
		// Given
		name := strings.Split(t.Name(), "/")[1]
		perm := strings.Split(name, "-")
		setup := mock.Chain(
			mock.Sub(0, 0, mock.Chain(
				CallA("a"),
				CallA("b"),
			)),
			mock.Sub(0, -1, mock.Parallel(
				CallB("c", "d"),
				CallB("d", "e"),
			)),
			mock.Sub(0, 0, CallA("e")),
			mock.Sub(2, 2, mock.Setup(CallA("f"))),
		)
		mock := MockSetup(t, setup)

		// When
		test := SetupPermTestABCDEF(mock)

		// Then
		test.Test(t, perm, expect)
	})
}

var detachTestCases = perm.ExpectMap{
	"a-b-c-d": test.Success,
	"a-b-d-c": test.Success,
	"a-d-b-c": test.Success,
	"b-a-c-d": test.Success,
	"b-a-d-c": test.Success,
	"b-d-a-c": test.Success,
	"d-a-b-c": test.Success,
	"d-b-a-c": test.Success,
}

func TestDetach(t *testing.T) {
	perms := detachTestCases.Remain(test.Failure)
	test.Map(t, perms).Run(func(t test.Test, expect test.Expect) {
		// Given
		name := strings.Split(t.Name(), "/")[1]
		perm := strings.Split(name, "-")
		setup := mock.Chain(
			mock.Detach(mock.None, CallA("a")),
			mock.Detach(mock.Head, CallA("b")),
			mock.Detach(mock.Tail, CallB("c", "d")),
			mock.Detach(mock.Both, CallB("d", "e")),
		)
		mock := MockSetup(t, setup)

		// When
		test := SetupPermTestABCD(mock)

		// Then
		test.Test(t, perm, expect)
	})
}

type PanicParams struct {
	setup       mock.SetupFunc
	expectError error
}

var panicTestCases = map[string]PanicParams{
	"setup": {
		setup:       mock.Setup(NoCall()),
		expectError: mock.NewErrNoCall(NewMockIFace(nil).EXPECT()),
	},
	"chain": {
		setup:       mock.Chain(NoCall()),
		expectError: mock.NewErrNoCall(NewMockIFace(nil).EXPECT()),
	},
	"parallel": {
		setup:       mock.Parallel(NoCall()),
		expectError: mock.NewErrNoCall(NewMockIFace(nil).EXPECT()),
	},
	"detach": {
		setup:       mock.Detach(4, NoCall()),
		expectError: mock.NewErrDetachMode(4),
	},
	"sub": {
		setup:       mock.Sub(0, 0, NoCall()),
		expectError: mock.NewErrNoCall(NewMockIFace(nil).EXPECT()),
	},
	"sub-head": {
		setup:       mock.Sub(0, 0, mock.Detach(mock.Head, NoCall())),
		expectError: mock.NewErrDetachNotAllowed(mock.Head),
	},
	"sub-tail": {
		setup:       mock.Sub(0, 0, mock.Detach(mock.Tail, NoCall())),
		expectError: mock.NewErrDetachNotAllowed(mock.Tail),
	},
	"sub-both": {
		setup:       mock.Sub(0, 0, mock.Detach(mock.Both, NoCall())),
		expectError: mock.NewErrDetachNotAllowed(mock.Both),
	},
}

func TestPanic(t *testing.T) {
	test.Map(t, panicTestCases).Run(func(t test.Test, param PanicParams) {
		// Given
		defer func() {
			err := recover()
			assert.Equal(t, param.expectError, err)
		}()

		// When
		MockSetup(t, param.setup)

		// Then
		require.Fail(t, "not paniced")
	})
}

type GetSubSliceParams struct {
	slice       []any
	from, to    int
	expectSlice any
}

var getSubSliceTestCases = map[string]GetSubSliceParams{
	"first": {
		slice: []any{"a", "b", "c", "d", "e"},
		from:  0, to: 0,
		expectSlice: "a",
	},
	"last": {
		slice: []any{"a", "b", "c", "d", "e"},
		from:  -1, to: -1,
		expectSlice: "e",
	},
	"middle": {
		slice: []any{"a", "b", "c", "d", "e"},
		from:  2, to: 2,
		expectSlice: "c",
	},
	"begin": {
		slice: []any{"a", "b", "c", "d", "e"},
		from:  0, to: 2,
		expectSlice: []any{"a", "b", "c"},
	},
	"end": {
		slice: []any{"a", "b", "c", "d", "e"},
		from:  -3, to: -1,
		expectSlice: []any{"c", "d", "e"},
	},
	"all": {
		slice: []any{"a", "b", "c", "d", "e"},
		from:  0, to: -1,
		expectSlice: []any{"a", "b", "c", "d", "e"},
	},
	"sub": {
		slice: []any{"a", "b", "c", "d", "e"},
		from:  -2, to: 1,
		expectSlice: []any{"b", "c", "d"},
	},
	"out-of-bound": {
		slice: []any{"a", "b", "c", "d", "e"},
		from:  -7, to: 7,
		expectSlice: []any{"a", "b", "c", "d", "e"},
	},
}

func TestGetSubSlice(t *testing.T) {
	test.Map(t, getSubSliceTestCases).
		Run(func(t test.Test, param GetSubSliceParams) {
			// When
			slice := mock.GetSubSlice(param.from, param.to, param.slice)

			// Then
			assert.Equal(t, param.expectSlice, slice)
		})
}

type FuncParams struct {
	setup  mock.SetupFunc
	call   any
	result []any
	expect []any
}

var funcTestCases = map[string]FuncParams{
	"in-0-out-0": {
		call: func(any) {},
	},
	"in-1-out-0": {
		call: func(any, any) {},
	},
	"in-1-var-out-0": {
		call: func(any, ...any) {},
	},
	"in-2-out-0": {
		call: func(any, any, any) {},
	},
	"in-2-var-out-0": {
		call: func(any, any, ...any) {},
	},
	"in-4-out-0": {
		call: func(any, any, any, any, any) {},
	},
	"in-4-var-out-0": {
		call: func(any, any, any, any, ...any) {},
	},

	"in-0-out-1": {
		call:   func(any) any { return nil },
		result: []any{"string"},
		expect: []any{"string"},
	},
	"in-0-out-2": {
		call:   func(any) (any, any) { return nil, nil },
		result: []any{"string", 1},
		expect: []any{"string", 1},
	},
	"in-0-out-3": {
		call: func(any) (any, any, any) {
			return nil, nil, nil
		},
		result: []any{"string", 1, true},
		expect: []any{"string", 1, true},
	},
	"in-0-out-4": {
		call: func(any) (any, any, any, any) {
			return nil, nil, nil, nil
		},
		result: []any{"string", 1, true, assert.AnError},
		expect: []any{"string", 1, true, assert.AnError},
	},
}

var funcDoNoReturnTestCases = map[string]FuncParams{
	"in-0-no-out-1": {
		call:   func(any) string { return "okay" },
		result: []any{},
		expect: []any{""},
	},
	"in-0-no-out-2": {
		call:   func(any) (string, int) { return "okay", 1 },
		result: []any{},
		expect: []any{"", 0},
	},
	"in-0-no-out-3": {
		call: func(any) (string, int, bool) {
			return "okay", 1, true
		},
		result: []any{},
		expect: []any{"", 0, false},
	},
	"in-0-no-out-4": {
		call: func(any) (string, int, bool, any) {
			return "oaky", 1, true, nil
		},
		result: []any{},
		expect: []any{"", 0, false, nil},
	},
}

func TestFuncDo(t *testing.T) {
	test.Map(t, funcTestCases, funcDoNoReturnTestCases).
		Run(func(t test.Test, param FuncParams) {
			// Given
			mocks := MockSetup(t, param.setup)
			ctype := reflect.TypeOf(param.call)

			// When
			call := mocks.Do(param.call, param.result...)

			// Then
			ftype := reflect.TypeOf(call)
			assert.Equal(t, ctype.NumIn()-1, ftype.NumIn())
			assert.Equal(t, ctype.NumOut(), ftype.NumOut())
			if len(param.result) > 0 {
				assert.Equal(t, len(param.result), ftype.NumOut())
			}

			// When
			result := reflect.ArgsOf(reflect.ValueOf(call).Call(
				reflect.ValuesIn(ftype, make([]any, ftype.NumIn())...),
			)...)

			// Then
			assert.Equal(t, param.expect, result)
			mocks.Wait()
		})
}

var funcReturnNoneTestCases = map[string]FuncParams{
	"in-0-no-out-1": {
		setup:  test.Panic("not enough arguments"),
		call:   func(any) string { return "okay" },
		result: []any{},
		expect: []any{""},
	},
	"in-0-no-out-2": {
		setup:  test.Panic("not enough arguments"),
		call:   func(any) (string, int) { return "okay", 1 },
		result: []any{},
		expect: []any{"", 0},
	},
	"in-0-no-out-3": {
		setup: test.Panic("not enough arguments"),
		call: func(any) (string, int, bool) {
			return "okay", 1, true
		},
		result: []any{},
		expect: []any{"", 0, false},
	},
	"in-0-no-out-4": {
		setup: test.Panic("not enough arguments"),
		call: func(any) (string, int, bool, any) {
			return "oaky", 1, true, nil
		},
		result: []any{},
		expect: []any{"", 0, false, nil},
	},
}

func TestFuncReturn(t *testing.T) {
	test.Map(t, funcTestCases, funcReturnNoneTestCases).
		Run(func(t test.Test, param FuncParams) {
			// Given
			mocks := MockSetup(t, param.setup)
			ctype := reflect.TypeOf(param.call)

			// When
			call := mocks.Return(param.call, param.result...)

			// Then
			ftype := reflect.TypeOf(call)
			assert.Equal(t, ctype.NumIn()-1, ftype.NumIn())
			assert.Equal(t, ctype.NumOut(), ftype.NumOut())
			if len(param.result) > 0 {
				assert.Equal(t, len(param.result), ftype.NumOut())
			}

			// When
			result := reflect.ArgsOf(reflect.ValueOf(call).Call(
				reflect.ValuesIn(ftype, make([]any, ftype.NumIn())...),
			)...)

			// Then
			assert.Equal(t, param.expect, result)
			mocks.Wait()
		})
}

func TestFuncPanic(t *testing.T) {
	test.Map(t, funcTestCases, funcDoNoReturnTestCases).
		Run(func(t test.Test, param FuncParams) {
			// Given
			mocks := MockSetup(t, param.setup)
			ctype := reflect.TypeOf(param.call)
			defer func() {
				require.Equal(t, "panic-test", recover())
				mocks.Wait()
			}()

			// When
			call := mocks.Panic(param.call, "panic-test")

			// Then
			ftype := reflect.TypeOf(call)
			assert.Equal(t, ctype.NumIn()-1, ftype.NumIn())
			assert.Equal(t, ctype.NumOut(), ftype.NumOut())
			if len(param.result) > 0 {
				assert.Equal(t, len(param.result), ftype.NumOut())
			}

			// When
			reflect.ValueOf(call).Call(
				reflect.ValuesIn(ftype, make([]any, ftype.NumIn())...),
			)

			// Then
			assert.Fail(t, "not paniced")
		})
}

type FailureParams struct {
	expect test.Expect
	test   test.Func
}

var failureTestCases = map[string]FailureParams{
	"success": {
		test:   func(test.Test) {},
		expect: test.Success,
	},

	"errorf": {
		test:   func(t test.Test) { t.Errorf("%s", "fail") },
		expect: test.Failure,
	},

	"fatalf": {
		test:   func(t test.Test) { t.Fatalf("%s", "fail") },
		expect: test.Failure,
	},

	"failnow": {
		test:   func(t test.Test) { t.FailNow() },
		expect: test.Failure,
	},

	"panic": {
		test:   func(test.Test) { panic("panic") },
		expect: test.Failure,
	},
}

func TestFailures(t *testing.T) {
	test.Map(t, failureTestCases).
		Run(func(t test.Test, param FailureParams) {
			// Given
			mocks := mock.NewMocks(t).Expect(CallA("a"))
			defer func() {
				if err := recover(); err != nil && err != "panic" {
					// Test thread will not wait on failures.
					mocks.Wait()
				}
			}()

			// When
			param.test(t)

			// Then
			mock.Get(mocks, NewMockIFace).CallA("a")
			mocks.Wait()
		})
}

// TODO: add more adequate testing for waiting.
type WaitParams struct {
	expect test.Expect
}

var waitTestCases = map[string]WaitParams{
	"simple wait": {
		expect: test.Success,
	},
}

func TestFuncWait(t *testing.T) {
	test.Map(t, waitTestCases).
		Run(func(t test.Test, _ WaitParams) {
			// Given
			mocks := mock.NewMocks(t)

			// When
			mocks.Add(5)
			mocks.Times(-5)
			mocks.Done()

			// Then
			mocks.Wait()
		})
}
