# Usage patterns of testing/test

This package provides a small framework to simply isolate the test execution
and safely check whether a test fails as expected. This is primarily very handy
to validate a test framework as provided by the [mock](../mock) package but may
be handy in other cases too.


## Isolation pattern for parameterized test

The `test` framework simply allows to isolate a test by calling the system
under test via `test.Run(test.ExpectFailure, func(t *test.TestingT) {})` to
create the test wrapper. Unfortunately, this alters the signature of the test
function slightly, since the `Golang` test environment is not porivding the
necessary abstractions to directly use it, however, the introduced abstraction
is compatible with major tooling, e.g. [Gomock](https://github.com/golang/mock).
The main pattern for parameterized unit test looks as follows:

```go
func TestSetupChain(t *testing.T) {
	for message, param := range testParams {
		t.Run(message, test.Run(param.expect, func(t *test.TestingT) {
			require.NotEmpty(t, message)

			// Given

			// When

			// Then
		}))
	}
}
```


## Isolation pattner for simple tests

Besides, providing `Run(Exepct, func(*test.TestingT)) func(*testing.T)` for
parameterized test, the package contains two further methods for isolation in
simpler test situations:

* `Failure(*testing.T, func(*test.TestingT) {})`, and
* `Success(*testing.T, func(*test.TestingT) {})`.

In more generic situations you can also use the base framework:

```go
  NewTestingT(*testing.T, test.Expect).test(func(*testing.T){})`
```
