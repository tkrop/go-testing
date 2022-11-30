package mock_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tkrop/testing/mock"
	"github.com/tkrop/testing/perm"
	"github.com/tkrop/testing/test"
)

//go:generate mockgen -package=mock_test -destination=mock_iface_test.go -source=mock_test.go  IFace

type IFace interface {
	CallA(string)
	CallB(string) string
}

func CallA(input string) mock.SetupFunc {
	return func(mocks *mock.Mocks) any {
		return mock.Get(mocks, NewMockIFace).EXPECT().
			CallA(input).Times(mocks.Times(1)).
			Do(mocks.GetDone(1))
	}
}

func CallB(input string, output string) mock.SetupFunc {
	return func(mocks *mock.Mocks) any {
		return mock.Get(mocks, NewMockIFace).EXPECT().
			CallB(input).Return(output).
			Times(mocks.Times(1)).Do(mocks.GetDone(1))
	}
}

func NoCall() mock.SetupFunc {
	return func(mocks *mock.Mocks) any {
		return mock.Get(mocks, NewMockIFace).EXPECT()
	}
}

func MockSetup(t gomock.TestReporter, mockSetup mock.SetupFunc) *mock.Mocks {
	return mock.NewMock(t).Expect(mockSetup)
}

func MockValidate(
	t *test.TestingT, mocks *mock.Mocks,
	validate func(*test.TestingT, *mock.Mocks),
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
			"a": func(t *test.TestingT) { iface.CallA("a") },
			"b1": func(t *test.TestingT) {
				assert.Equal(t, "c", iface.CallB("b"))
			},
			"b2": func(t *test.TestingT) {
				assert.Equal(t, "d", iface.CallB("b"))
			},
			"c": func(t *test.TestingT) { iface.CallA("c") },
		})
}

func SetupPermTestABCD(mocks *mock.Mocks) *perm.Test {
	iface := mock.Get(mocks, NewMockIFace)
	return perm.NewTest(mocks,
		perm.TestMap{
			"a": func(t *test.TestingT) { iface.CallA("a") },
			"b": func(t *test.TestingT) { iface.CallA("b") },
			"c": func(t *test.TestingT) {
				assert.Equal(t, "d", iface.CallB("c"))
			},
			"d": func(t *test.TestingT) {
				assert.Equal(t, "e", iface.CallB("d"))
			},
		})
}

func SetupPermTestABCDEF(mocks *mock.Mocks) *perm.Test {
	iface := mock.Get(mocks, NewMockIFace)
	return perm.NewTest(mocks,
		perm.TestMap{
			"a": func(t *test.TestingT) { iface.CallA("a") },
			"b": func(t *test.TestingT) { iface.CallA("b") },
			"c": func(t *test.TestingT) {
				assert.Equal(t, "d", iface.CallB("c"))
			},
			"d": func(t *test.TestingT) {
				assert.Equal(t, "e", iface.CallB("d"))
			},
			"e": func(t *test.TestingT) { iface.CallA("e") },
			"f": func(t *test.TestingT) { iface.CallA("f") },
		})
}

var testSetupParams = perm.ExpectMap{
	"b2-b1-a-c": test.ExpectFailure,
	"b2-b1-c-a": test.ExpectFailure,
	"b2-c-b1-a": test.ExpectFailure,
	"b2-a-b1-c": test.ExpectFailure,
	"b2-c-a-b1": test.ExpectFailure,
	"b2-a-c-b1": test.ExpectFailure,
	"a-b2-b1-c": test.ExpectFailure,
	"c-b2-b1-a": test.ExpectFailure,
	"a-b2-c-b1": test.ExpectFailure,
	"c-b2-a-b1": test.ExpectFailure,
	"c-a-b2-b1": test.ExpectFailure,
	"a-c-b2-b1": test.ExpectFailure,
}

func TestSetup(t *testing.T) {
	t.Parallel()

	for message, expect := range testSetupParams.Remain(test.ExpectSuccess) {
		message, expect := message, expect
		t.Run(message, test.Run(expect, func(t *test.TestingT) {
			require.NotEmpty(t, message)

			// Given
			perm := strings.Split(message, "-")
			mockSetup := mock.Setup(
				CallA("a"),
				mock.Setup(
					CallB("b", "c"),
					CallB("b", "d"),
				),
				CallA("c"),
			)
			mock := MockSetup(t, mockSetup)

			// When
			test := SetupPermTestABC(mock)

			// Then
			test.Test(t, perm, expect)
		}, false))
	}
}

var testChainParams = perm.ExpectMap{
	"a-b1-b2-c": test.ExpectSuccess,
}

func TestChain(t *testing.T) {
	t.Parallel()

	for message, expect := range testChainParams.Remain(test.ExpectFailure) {
		message, expect := message, expect
		t.Run(message, test.Run(expect, func(t *test.TestingT) {
			require.NotEmpty(t, message)

			// Given
			perm := strings.Split(message, "-")
			mockSetup := mock.Chain(
				CallA("a"),
				mock.Chain(
					CallB("b", "c"),
					CallB("b", "d"),
				),
				CallA("c"),
			)
			mock := MockSetup(t, mockSetup)

			// When
			test := SetupPermTestABC(mock)

			// Then
			test.Test(t, perm, expect)
		}, false))
	}
}

var testSetupChainParams = perm.ExpectMap{
	"a-b-c-d": test.ExpectSuccess,
	"a-c-b-d": test.ExpectSuccess,
	"a-c-d-b": test.ExpectSuccess,
	"c-a-b-d": test.ExpectSuccess,
	"c-a-d-b": test.ExpectSuccess,
	"c-d-a-b": test.ExpectSuccess,
}

func TestSetupChain(t *testing.T) {
	t.Parallel()

	for message, expect := range testSetupChainParams.Remain(test.ExpectFailure) {
		message, expect := message, expect
		t.Run(message, test.Run(expect, func(t *test.TestingT) {
			require.NotEmpty(t, message)

			// Given
			perm := strings.Split(message, "-")

			// Basic setup of two independent chains.
			mockSetup := mock.Setup(
				mock.Chain(
					CallA("a"),
					CallA("b"),
				),
				mock.Chain(
					CallB("c", "d"),
					CallB("d", "e"),
				),
			)
			mock := MockSetup(t, mockSetup)

			// When
			test := SetupPermTestABCD(mock)

			// Then
			test.Test(t, perm, expect)
		}, false))
	}
}

func TestChainSetup(t *testing.T) {
	t.Parallel()

	for message, expect := range testSetupChainParams.Remain(test.ExpectFailure) {
		message, expect := message, expect
		t.Run(message, test.Run(expect, func(t *test.TestingT) {
			require.NotEmpty(t, message)

			// Given
			perm := strings.Split(message, "-")

			// Frail setup to detach a (sub-)chain.
			mockSetup := mock.Chain(
				CallA("a"),
				CallA("b"),
				mock.Setup( // detaching (sub-)chain.
					mock.Chain(
						CallB("c", "d"),
						CallB("d", "e"),
					),
				),
			)
			mock := MockSetup(t, mockSetup)

			// When
			test := SetupPermTestABCD(mock)

			// Then
			test.Test(t, perm, expect)
		}, false))
	}
}

var testParallelChainParams = perm.ExpectMap{
	"a-b-c-d-e-f": test.ExpectSuccess,
	"a-b-c-e-d-f": test.ExpectSuccess,
	"a-b-e-c-d-f": test.ExpectSuccess,
	"a-c-b-d-e-f": test.ExpectSuccess,
	"a-c-d-b-e-f": test.ExpectSuccess,
	"a-c-d-e-b-f": test.ExpectSuccess,
	"a-c-b-e-d-f": test.ExpectSuccess,
	"a-c-e-d-b-f": test.ExpectSuccess,
	"a-c-e-b-d-f": test.ExpectSuccess,
	"a-e-b-c-d-f": test.ExpectSuccess,
	"a-e-c-b-d-f": test.ExpectSuccess,
	"a-e-c-d-b-f": test.ExpectSuccess,
}

func TestParallelChain(t *testing.T) {
	t.Parallel()

	for message, expect := range testParallelChainParams.Remain(test.ExpectFailure) {
		message, expect := message, expect
		t.Run(message, test.Run(expect, func(t *test.TestingT) {
			require.NotEmpty(t, message)

			// Given
			perm := strings.Split(message, "-")
			mockSetup := mock.Chain(
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
			mock := MockSetup(t, mockSetup)

			// When
			test := SetupPermTestABCDEF(mock)

			// Then
			test.Test(t, perm, expect)
		}, false))
	}
}

var testChainSubParams = perm.ExpectMap{
	"a-b-c-d-e-f": test.ExpectSuccess,
	"a-c-b-d-e-f": test.ExpectSuccess,
	"a-c-d-b-e-f": test.ExpectSuccess,
	"a-c-d-e-b-f": test.ExpectSuccess,
	"f-a-b-c-d-e": test.ExpectSuccess,
	"a-f-b-c-d-e": test.ExpectSuccess,
	"a-b-f-c-d-e": test.ExpectSuccess,
	"a-b-c-f-d-e": test.ExpectSuccess,
	"a-b-c-d-f-e": test.ExpectSuccess,
	"a-c-d-e-f-b": test.ExpectSuccess,

	"b-a-c-d-e-f": test.ExpectFailure,
	"c-a-b-d-e-f": test.ExpectFailure,
	"d-a-b-c-e-f": test.ExpectFailure,
	"a-b-c-e-d-f": test.ExpectFailure,
	"a-b-d-e-c-f": test.ExpectFailure,
}

func TestChainSub(t *testing.T) {
	t.Parallel()

	perms := testChainSubParams
	//	perms := PermRemain(testChainSubParams, test.ExpectFailure)
	for message, expect := range perms {
		message, expect := message, expect
		t.Run(message, test.Run(expect, func(t *test.TestingT) {
			// Given
			perm := strings.Split(message, "-")
			mockSetup := mock.Chain(
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
			mock := MockSetup(t, mockSetup)

			// When
			test := SetupPermTestABCDEF(mock)

			// Then
			test.Test(t, perm, expect)
		}, false))
	}
}

var testDetachParams = perm.ExpectMap{
	"a-b-c-d": test.ExpectSuccess,
	"a-b-d-c": test.ExpectSuccess,
	"a-d-b-c": test.ExpectSuccess,
	"b-a-c-d": test.ExpectSuccess,
	"b-a-d-c": test.ExpectSuccess,
	"b-d-a-c": test.ExpectSuccess,
	"d-a-b-c": test.ExpectSuccess,
	"d-b-a-c": test.ExpectSuccess,
}

func TestDetach(t *testing.T) {
	t.Parallel()

	for message, expect := range testDetachParams.Remain(test.ExpectFailure) {
		message, expect := message, expect
		t.Run(message, test.Run(expect, func(t *test.TestingT) {
			// Given
			perm := strings.Split(message, "-")
			mockSetup := mock.Chain(
				mock.Detach(mock.None, CallA("a")),
				mock.Detach(mock.Head, CallA("b")),
				mock.Detach(mock.Tail, CallB("c", "d")),
				mock.Detach(mock.Both, CallB("d", "e")),
			)
			mock := MockSetup(t, mockSetup)

			// When
			test := SetupPermTestABCD(mock)

			// Then
			test.Test(t, perm, expect)
		}, false))
	}
}

var testPanicParams = map[string]struct {
	setup       mock.SetupFunc
	expectError error
}{
	"setup": {
		setup:       mock.Setup(NoCall()),
		expectError: mock.ErrNoCall(NewMockIFace(nil).EXPECT()),
	},
	"chain": {
		setup:       mock.Chain(NoCall()),
		expectError: mock.ErrNoCall(NewMockIFace(nil).EXPECT()),
	},
	"parallel": {
		setup:       mock.Parallel(NoCall()),
		expectError: mock.ErrNoCall(NewMockIFace(nil).EXPECT()),
	},
	"detach": {
		setup:       mock.Detach(4, NoCall()),
		expectError: mock.ErrDetachMode(4),
	},
	"sub": {
		setup:       mock.Sub(0, 0, NoCall()),
		expectError: mock.ErrNoCall(NewMockIFace(nil).EXPECT()),
	},
	"sub-head": {
		setup:       mock.Sub(0, 0, mock.Detach(mock.Head, NoCall())),
		expectError: mock.ErrDetachNotAllowed(mock.Head),
	},
	"sub-tail": {
		setup:       mock.Sub(0, 0, mock.Detach(mock.Tail, NoCall())),
		expectError: mock.ErrDetachNotAllowed(mock.Tail),
	},
	"sub-both": {
		setup:       mock.Sub(0, 0, mock.Detach(mock.Both, NoCall())),
		expectError: mock.ErrDetachNotAllowed(mock.Both),
	},
}

func TestPanic(t *testing.T) {
	t.Parallel()

	for message, param := range testPanicParams {
		message, param := message, param
		t.Run(message, func(t *testing.T) {
			// Given
			defer func() {
				err := recover()
				assert.Equal(t, param.expectError, err)
			}()

			// When
			mock := MockSetup(t, param.setup)

			// Then
			require.Fail(t, "not paniced")
			mock.Wait()
		})
	}
}

var testGetSubSliceParams = map[string]struct {
	slice       []any
	from, to    int
	expectSlice any
}{
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
	t.Parallel()

	for message, param := range testGetSubSliceParams {
		message, param := message, param
		t.Run(message, func(t *testing.T) {
			// When
			slice := mock.GetSubSlice(param.from, param.to, param.slice)

			// Then
			assert.Equal(t, param.expectSlice, slice)
		})
	}
}

var testGetFuncParams = map[string]struct {
	numargs int
	exist   bool
}{
	"test 0 args":  {numargs: 0, exist: true},
	"test 1 args":  {numargs: 1, exist: true},
	"test 2 args":  {numargs: 2, exist: true},
	"test 3 args":  {numargs: 3, exist: true},
	"test 4 args":  {numargs: 4, exist: true},
	"test 5 args":  {numargs: 5, exist: true},
	"test 6 args":  {numargs: 6, exist: true},
	"test 7 args":  {numargs: 7, exist: true},
	"test 8 args":  {numargs: 8, exist: true},
	"test 9 args":  {numargs: 9, exist: true},
	"test 10 args": {numargs: 10},
	"test 11 args": {numargs: 11},
}

func TestGetDone(t *testing.T) {
	t.Parallel()

	for message, param := range testGetFuncParams {
		message, param := message, param
		t.Run(message, func(t *testing.T) {
			// Given
			mocks := MockSetup(t, nil)
			mocks.Times(1)
			if !param.exist {
				defer func() { recover() }()
			}

			// When
			fncall := mocks.GetDone(param.numargs)
			switch param.numargs {
			case 0:
				fncall.(func())()
			case 1:
				fncall.(func(any))(nil)
			case 2:
				fncall.(func(any, any))(nil, nil)
			case 3:
				fncall.(func(any, any, any))(nil, nil, nil)
			case 4:
				fncall.(func(any, any, any, any))(nil, nil, nil, nil)
			case 5:
				fncall.(func(
					any, any, any, any, any,
				))(nil, nil, nil, nil, nil)
			case 6:
				fncall.(func(
					any, any, any, any, any, any,
				))(nil, nil, nil, nil, nil, nil)
			case 7:
				fncall.(func(
					any, any, any, any, any, any, any,
				))(nil, nil, nil, nil, nil, nil, nil)
			case 8:
				fncall.(func(
					any, any, any, any, any, any, any, any,
				))(nil, nil, nil, nil, nil, nil, nil, nil)
			case 9:
				fncall.(func(
					any, any, any, any, any, any, any, any, any,
				))(nil, nil, nil, nil, nil, nil, nil, nil, nil)
			}

			// Then
			mocks.Wait()
			if !param.exist {
				assert.Fail(t, "not paniced on not supported argument number")
			}
		})
	}
}

func TestGetPanic(t *testing.T) {
	t.Parallel()

	for message, param := range testGetFuncParams {
		message, param := message, param
		t.Run(message, func(t *testing.T) {
			// Given
			mocks := MockSetup(t, nil)
			mocks.Times(1)
			defer func() {
				reason := recover()
				// Then
				if param.exist {
					require.Equal(t, "panic-test", reason)
					mocks.Wait()
				} else {
					assert.Equal(t, fmt.Sprintf(
						"argument number not supported: %d",
						param.numargs), reason)
				}
			}()

			// When
			fncall := mocks.GetPanic(param.numargs, "panic-test")
			switch param.numargs {
			case 0:
				fncall.(func())()
			case 1:
				fncall.(func(any))(nil)
			case 2:
				fncall.(func(any, any))(nil, nil)
			case 3:
				fncall.(func(any, any, any))(nil, nil, nil)
			case 4:
				fncall.(func(any, any, any, any))(nil, nil, nil, nil)
			case 5:
				fncall.(func(
					any, any, any, any, any,
				))(nil, nil, nil, nil, nil)
			case 6:
				fncall.(func(
					any, any, any, any, any, any,
				))(nil, nil, nil, nil, nil, nil)
			case 7:
				fncall.(func(
					any, any, any, any, any, any, any,
				))(nil, nil, nil, nil, nil, nil, nil)
			case 8:
				fncall.(func(
					any, any, any, any, any, any, any, any,
				))(nil, nil, nil, nil, nil, nil, nil, nil)
			case 9:
				fncall.(func(
					any, any, any, any, any, any, any, any, any,
				))(nil, nil, nil, nil, nil, nil, nil, nil, nil)
			}

			// Then
			assert.Fail(t, "not paniced on not supported argument number")
		})
	}
}
