package perm

import (
	"strings"

	"github.com/stretchr/testify/require"

	"github.com/tkrop/testing/mock"
	"github.com/tkrop/testing/test"
	"github.com/tkrop/testing/utils/slices"
)

// ExpectMap defines a map of permutation tests that are expected to either
// fail or succeed to succeed depending on expressed expectation.
type ExpectMap map[string]test.Expect

// TestMap defines a map of test functions that is subject of the actual
// permutation.
type TestMap map[string]func(t *test.TestingT)

// Test permutation test.
type Test struct {
	mocks *mock.Mocks
	tests TestMap
}

// NewTest creates a new permutation test with given mock and given
// permutation test map.
func NewTest(mocks *mock.Mocks, tests TestMap) *Test {
	return &Test{
		mocks: mocks,
		tests: tests,
	}
}

// TestPerm tests a single permutation given by the string slice.
func (p *Test) TestPerm(t *test.TestingT, perm []string) {
	require.Equal(t, len(p.tests), len(perm),
		"permutation needs to cover all tests")
	for _, value := range perm {
		p.tests[value](t)
	}
}

// Test executes a permutation test with given permutation and expected result.
func (p *Test) Test(t *test.TestingT, perm []string, expect test.Expect) {
	switch expect {
	case test.ExpectSuccess:
		// Test proper usage of `WaitGroup` on non-failing validation.
		p.TestPerm(t, perm)
		p.mocks.Wait()
	case test.ExpectFailure:
		// we need to execute failing test synchronous, since we setup full
		// permutations instead of stopping setup on first failing mock calls.
		p.TestPerm(t, perm)
	}
}

// Remain calculate and add the missing permutations and add it with
// expected result to the given permmutation map.
func (perms ExpectMap) Remain(expect test.Expect) ExpectMap {
	cperms := ExpectMap{}
	for key, value := range perms {
		cperms[key] = value
	}

	// we only need to permutate the first key.
	for key := range cperms {
		slices.Permute(strings.Split(key, "-"),
			func(perm []string) {
				key := strings.Join(perm, "-")
				if _, ok := cperms[key]; !ok {
					cperms[key] = expect
				}
			}, 0)
		break
	}

	return cperms
}
