package test

type (
	// Expect the expectation whether a test will succeed or fail.
	Expect bool
	// Name represents a test case name.
	Name string
)

// Constants to express test expectations.
const (
	// Success used to express that a test is supposed to succeed.
	Success Expect = true
	// Failure used to express that a test is supposed to fail.
	Failure Expect = false

	// unknown default unknown test case name.
	unknown Name = "unknown"

	// Parallel is the global flag to switch test runs to be executed in
	// parallel instead of sequentially.
	Parallel = true
)
