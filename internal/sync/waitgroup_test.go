package sync_test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tkrop/go-testing/internal/sync"
	"github.com/tkrop/go-testing/test"
)

func TestWaitGroup(t *testing.T) {
	t.Parallel()

	// Given
	defer test.Recover(t, "sync: negative WaitGroup counter")
	wg := sync.NewWaitGroup()

	// When
	wg.Add(3)
	wg.Done()
	wg.Add(math.MinInt32)
	wg.Done()

	// Then
	assert.Fail(t, "did not panic")
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
