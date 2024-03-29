package sync_test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tkrop/go-testing/internal/sync"
)

func TestWaitGroup(t *testing.T) {
	t.Parallel()

	// Given
	defer func() { _ = recover() }()
	wg := sync.NewWaitGroup()

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
	wg := sync.NewLenientWaitGroup()

	// When
	wg.Add(3)
	wg.Done()
	wg.Add(math.MinInt32)
	wg.Done()

	// Then
	wg.Wait()
}
