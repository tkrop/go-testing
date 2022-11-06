package sync

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWaitGroup(t *testing.T) {
	defer func() { recover() }()
	// Given
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
