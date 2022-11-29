package sync

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWaitGroup(t *testing.T) {
	t.Parallel()

	// Given
	defer func() { recover() }()
	wg := NewWaitGroup()

	// When
	wg.Add(3)
	wg.Done()
	wg.Add(math.MinInt32)
	wg.Done()

	// Then
	assert.Fail(t, "not recovered from panic")
}

func TestLenientWaitGroup(t *testing.T) {
	t.Parallel()

	// Given
	wg := NewLenientWaitGroup()

	// When
	wg.Add(3)
	wg.Done()
	wg.Add(math.MinInt32)
	wg.Done()

	// Then
	wg.Wait()
}
