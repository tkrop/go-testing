package test

import (
	"testing"
	"time"
)

// Benchmarker extends Test with sub-benchmark execution for the two-phase
// benchmark pattern.
type Benchmarker interface {
	Test
	// Benchmark runs a sub-benchmark with the given name and function.
	Benchmark(name string, call func(*testing.B)) bool
}

// benchmark is a private wrapper around *testing.B that satisfies test.Test.
type benchmark struct {
	*testing.B
}

// Benchmark wraps *testing.B as a Test for use with test.Map and Factory.
func Benchmark(b *testing.B) Test {
	return &benchmark{B: b}
}

// Parallel is a no-op for benchmarks.
func (*benchmark) Parallel() {}

// Deadline returns zero time and false (benchmarks have no deadline).
func (*benchmark) Deadline() (time.Time, bool) {
	return time.Time{}, false
}

// Run dispatches a sub-benchmark by name. The function argument is typed as
// `func(*benchmark)` so that reflect.Run[Test] can bridge the inner benchmark
// to the Test interface.
func (b *benchmark) Run(name string, call func(*benchmark)) bool {
	return b.B.Run(name, func(b *testing.B) {
		call(&benchmark{B: b})
	})
}

// Benchmark delegates to *testing.B.Benchmark for the two-phase benchmark loop.
func (b *benchmark) Benchmark(name string, call func(*testing.B)) bool {
	return b.B.Run(name, call)
}
