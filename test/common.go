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

// TODO: consider following convenience methods:
//
// // Result is a convenience method that returns the first argument ans swollows
// // all others assuming that the first argument contains the important result to
// // focus the test at.
// func Result[T any](result T, swollowed any) T {
// 	return result
// }

// // Check is a convenience method that returns the second argument and swollows
// // the first used to focus a test on the second.
// func Check[T any](swollowed any, check T) T {
// 	return check
// }

// // NoError is a convenience method to check whether the second error argument
// // is providing and actual error while extracting the first argument only. If
// // the error argument is an error, the method panics providing the error.
// func NoError[T any](result T, err error) T {
// 	if err != nil {
// 		panic(err)
// 	}
// 	return result
// }

// // Ok is a convenience method to check whether the second boolean argument is
// // `true` while returning the first argument. If the boolean argument is
// // `false`, the method panics.
// func Ok[T any](result T, ok bool) T {
// 	if !ok {
// 		panic("bool not okay")
// 	}
// 	return result
// }
