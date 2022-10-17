package mock_test

import (
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tkrop/testing/mock"
	"github.com/tkrop/testing/perm"
	"github.com/tkrop/testing/test"
)

//go:generate mockgen -package=mock -destination=mock_iface_test.go -source=mock_test.go  IFace

type IFace interface {
	CallA(string)
	CallB(string) string
}

func CallA(input string) mock.SetupFunc {
	return func(mocks *mock.Mocks) any {
		mocks.WaitGroup().Add(1)
		return mock.Get(mocks, mock.NewMockIFace).EXPECT().
			CallA(input).Times(1).
			Do(func(arg any) {
				defer mocks.WaitGroup().Done()
			})
	}
}

func CallB(input string, output string) mock.SetupFunc {
	return func(mocks *mock.Mocks) any {
		mocks.WaitGroup().Add(1)
		return mock.Get(mocks, mock.NewMockIFace).EXPECT().
			CallB(input).Return(output).Times(1).
			Do(func(arg any) {
				defer mocks.WaitGroup().Done()
			})
	}
}

func NoCall() mock.SetupFunc {
	return func(mocks *mock.Mocks) any {
		return mock.Get(mocks, mock.NewMockIFace).EXPECT()
	}
}

func MockSetup(t gomock.TestReporter, mockSetup mock.SetupFunc) *mock.Mocks {
	return mock.NewMock(t).Expect(mockSetup)
}

func MockValidate(
	t *test.TestingT, mocks *mock.Mocks, validate func(*test.TestingT, *mock.Mocks), failing bool,
) {
	if failing {
		// we need to execute failing test synchronous, since we setup full
		// permutations instead of stopping setup on first failing mock calls.
		validate(t, mocks)
	} else {
		// Test proper usage of `WaitGroup` on non-failing validation.
		validate(t, mocks)
		mocks.WaitGroup().Wait()
	}
}

func SetupPermTestABC(mocks *mock.Mocks) *perm.Test {
	iface := mock.Get(mocks, mock.NewMockIFace)
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
	iface := mock.Get(mocks, mock.NewMockIFace)
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
	iface := mock.Get(mocks, mock.NewMockIFace)
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
	for message, expect := range testSetupParams.Remain(test.ExpectSuccess) {
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
		}))
	}
}

var testChainParams = perm.ExpectMap{
	"a-b1-b2-c": test.ExpectSuccess,
}

func TestChain(t *testing.T) {
	for message, expect := range testChainParams.Remain(test.ExpectFailure) {
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
		}))
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
	for message, expect := range testSetupChainParams.Remain(test.ExpectFailure) {
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
		}))
	}
}

func TestChainSetup(t *testing.T) {
	for message, expect := range testSetupChainParams.Remain(test.ExpectFailure) {
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
		}))
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
	for message, expect := range testParallelChainParams.Remain(test.ExpectFailure) {
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
		}))
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
	perms := testChainSubParams
	//	perms := PermRemain(testChainSubParams, test.ExpectFailure)
	for message, expect := range perms {
		t.Run(message, test.Run(expect, func(t *test.TestingT) {
			require.NotEmpty(t, message)

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
		}))
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
	for message, expect := range testDetachParams.Remain(test.ExpectFailure) {
		t.Run(message, test.Run(expect, func(t *test.TestingT) {
			require.NotEmpty(t, message)

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
		}))
	}
}

var testPanicParams = map[string]struct {
	setup       mock.SetupFunc
	expectError error
}{
	"setup": {
		setup:       mock.Setup(NoCall()),
		expectError: mock.ErrNoCall(mock.NewMockIFace(nil).EXPECT()),
	},
	"chain": {
		setup:       mock.Chain(NoCall()),
		expectError: mock.ErrNoCall(mock.NewMockIFace(nil).EXPECT()),
	},
	"parallel": {
		setup:       mock.Parallel(NoCall()),
		expectError: mock.ErrNoCall(mock.NewMockIFace(nil).EXPECT()),
	},
	"detach": {
		setup:       mock.Detach(4, NoCall()),
		expectError: mock.ErrDetachMode(4),
	},
	"sub": {
		setup:       mock.Sub(0, 0, NoCall()),
		expectError: mock.ErrNoCall(mock.NewMockIFace(nil).EXPECT()),
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
	for message, param := range testPanicParams {
		t.Run(message, func(t *testing.T) {
			require.NotEmpty(t, message)

			// Given
			defer func() {
				err := recover()
				assert.Equal(t, param.expectError, err)
			}()

			// When
			mock := MockSetup(t, param.setup)

			// Then
			require.Fail(t, "not paniced")
			mock.WaitGroup().Wait()
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
	for message, param := range testGetSubSliceParams {
		t.Run(message, func(t *testing.T) {
			require.NotEmpty(t, message)

			// Given

			// When
			slice := mock.GetSubSlice(param.from, param.to, param.slice)

			// Then
			assert.Equal(t, param.expectSlice, slice)
		})
	}
}
