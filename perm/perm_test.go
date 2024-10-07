package perm_test

import (
	"strings"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/tkrop/go-testing/mock"
	"github.com/tkrop/go-testing/perm"
	"github.com/tkrop/go-testing/test"
)

//go:generate mockgen -package=perm_test -destination=mock_iface_test.go -source=perm_test.go  IFace

type IFace interface {
	CallA(string)
}

func CallA(input string) mock.SetupFunc {
	return func(mocks *mock.Mocks) any {
		return mock.Get(mocks, NewMockIFace).EXPECT().
			CallA(input).Do(mocks.Do(IFace.CallA))
	}
}

func SetupPermTestABCDEF(mocks *mock.Mocks) *perm.Test {
	iface := mock.Get(mocks, NewMockIFace)
	return perm.NewTest(mocks,
		perm.TestMap{
			"a": func(test.Test) { iface.CallA("a") },
			"b": func(test.Test) { iface.CallA("b") },
			"c": func(test.Test) { iface.CallA("c") },
			"d": func(test.Test) { iface.CallA("d") },
			"e": func(test.Test) { iface.CallA("e") },
			"f": func(test.Test) { iface.CallA("f") },
		})
}

func MockSetup(t gomock.TestReporter, mockSetup mock.SetupFunc) *mock.Mocks {
	return mock.NewMocks(t).Expect(mockSetup)
}

var testPermTestParams = perm.ExpectMap{
	"b-a-c-d-e-f": test.Success,
	"a-b-c-d-e-f": test.Success,
	"a-c-b-d-e-f": test.Success,
	"a-c-d-b-e-f": test.Success,
	"a-c-d-e-b-f": test.Success,
	"a-c-d-e-f-b": test.Success,
}

func TestPermTest(t *testing.T) {
	test.Map(t, testPermTestParams.Remain(test.Failure)).
		Run(func(t test.Test, expect test.Expect) {
			// Given
			name := strings.Split(t.Name(), "/")[1]
			perm := strings.Split(name, "-")

			mockSetup := mock.Chain(
				CallA("a"),
				mock.Setup(
					CallA("b"),
				),
				CallA("c"),
				CallA("d"),
				CallA("e"),
				CallA("f"),
			)
			mock := MockSetup(t, mockSetup)

			// When
			test := SetupPermTestABCDEF(mock)

			// Then
			test.Test(t, perm, expect)
		})
}
