# Usage patterns of testing/test

This package provides a small framework to simply isolate the test execution
and safely check whether a test fails as expected. This is primarily very handy
to validate a test framework as provided by the [mock](../mock) package but may
be handy in other cases too.


## Isolated parameterized parallel test pattern

The `test` framework allows to isolate a test by calling the system under test
via `test.Run(test.ExpectFailure, func(t test.Test) {})` to create a wrapper.
The wrapper ensures that errors are intercepted and handled as defined by the
test expections, i.e. either `ExpectFailure` or `ExpectSuccess`.

To enable this we have defined a test interface `test.Test` that is compatible
with the requirements of most tools, e.g. [Gomock][gomock].

The main pattern for parameterized unit test looks as follows:

```go
func TestUnitCall(t *testing.T) {
	t.Parallel()

	for message, param := range testParams {
		message, param := message, param
		t.Run(message, test.Run(param.expect, func(t test.Test) {
			t.Parallel()

			// Given

			// When

			// Then
		}))
	}
}
```

Besides, `test.Run(Exepct, func(test.Test)) func(*testing.T)` optimized for
parameterized test, the package provides two further methods for isolation in
simpler test cases:

* `test.Failure(func(test.Test) {}) func(*testing.T)`, and
* `test.Success(func(test.Test) {}) func(*testing.T)`.

As well as similar set of methods for sub-isolation test cases, i.e.

* `test.Run(Exepct, func(test.Test)) func(test.Test)`,
* `test.Failure(func(test.Test) {}) func(test.Test)`, and
* `test.Success(func(test.Test) {}) func(test.Test)`.

**Note:** See [Parallel tests requirements](..#parallel-tests-requirements)
for more information on requirements in parallel parameterized tests.

[gomock]: https://github.com/golang/mock "GoMock"
