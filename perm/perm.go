package perm

import (
	"strings"

	"github.com/stretchr/testify/require"

	"github.com/tkrop/testing/mock"
	"github.com/tkrop/testing/test"
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
	for key := range perms {
		Slice(strings.Split(key, "-"), func(perm []string) {
			key := strings.Join(perm, "-")
			if _, ok := perms[key]; !ok {
				perms[key] = expect
			}
		}, 0)
		break // we only need to permutate the first key.
	}
	return perms
}

// Slice permutates the given slice starting at the position given by and
// call the `do` function on each permutation to collect the result. For a full
// permutation the `index` must start with `0`.
func Slice[T any](slice []T, do func([]T), index int) {
	if index <= len(slice) {
		Slice(slice, do, index+1)
		for offset := index + 1; offset < len(slice); offset++ {
			slice[index], slice[offset] = slice[offset], slice[index]
			Slice(slice, do, index+1)
			slice[index], slice[offset] = slice[offset], slice[index]
		}
	} else {
		do(slice)
	}
}
