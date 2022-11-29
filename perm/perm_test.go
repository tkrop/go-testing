package perm_test

import (
	"strings"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/tkrop/testing/mock"
	"github.com/tkrop/testing/perm"
	"github.com/tkrop/testing/test"
)

//go:generate mockgen -package=perm_test -destination=mock_iface_test.go -source=perm_test.go  IFace

type IFace interface {
	CallA(string)
}

func CallA(input string) mock.SetupFunc {
	return func(mocks *mock.Mocks) any {
		return mock.Get(mocks, NewMockIFace).EXPECT().
			CallA(input).Times(mocks.Times(1)).
			Do(mocks.GetDone(1))
	}
}

func SetupPermTestABC(mocks *mock.Mocks) *perm.Test {
	iface := mock.Get(mocks, NewMockIFace)
	return perm.NewTest(mocks,
		perm.TestMap{
			"a": func(t *test.TestingT) { iface.CallA("a") },
			"b": func(t *test.TestingT) { iface.CallA("b") },
			"c": func(t *test.TestingT) { iface.CallA("c") },
		})
}

func MockSetup(t gomock.TestReporter, mockSetup mock.SetupFunc) *mock.Mocks {
	return mock.NewMock(t).Expect(mockSetup)
}

var testTestParams = perm.ExpectMap{
	"b-a-c": test.ExpectSuccess,
	"a-b-c": test.ExpectSuccess,
	"a-c-b": test.ExpectSuccess,
}

func TestTest(t *testing.T) {
	t.Parallel()

	for message, expect := range testTestParams.Remain(test.ExpectFailure) {
		message, expect := message, expect
		t.Run(message, test.Run(expect, func(t *test.TestingT) {
			// Given
			perm := strings.Split(message, "-")
			mockSetup := mock.Chain(
				CallA("a"),
				mock.Setup(
					CallA("b"),
				),
				CallA("c"),
			)
			mock := MockSetup(t, mockSetup)

			// When
			test := SetupPermTestABC(mock)

			// Then
			test.Test(t, perm, expect)
		}, true))
	}
}
