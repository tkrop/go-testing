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
		message, param := message, param
		t.Run(message, test.Run(param.expect, func(t *test.TestingT) {
			// Given

			// When

			// Then
		}, false))
	}
}
```


## More isolation pattners

Besides, `test.Run(Exepct, func(*test.TestingT), bool) func(*testing.T)`
optimized for parameterized test, the package provides two further methods for
isolation in simpler cases:

* `Failure(func(*test.TestingT) {}, bool) func(*testing.T)`, and
* `Success(func(*test.TestingT) {}, bool) func(*testing.T)`.

All methods provide a `bool`-flag to run test in parallel.

**Note:** there are some general requirements for running test in parallel:

1. The tests *must not modify* environment variables dynamically.
2. The tests *must not require* reserved service ports and open listeners.
3. The tests *must not use* [Gock][gock] for mocking on HTTP transport level
   (please use the internal [gock](../gock)-controller package),
4. The tests *must not use* [monkey patching][monkey] to modify commonly used
   functions, e.g. `time.Now()`, and finaly
5. The tests *must not share* any other resources, e.g. objects or database
   schemas, that need to be updated during the test execution.

If this conditions hold, the following pattern can bused to support parallel
test execution in an isolated test environment.


```go
func TestSetupChain(t *testing.T) {
	t.Parallel()
	for message, param := range testParams {
		message, param := message, param
		t.Run(message, test.Run(param.expect, func(t *test.TestingT) {
			// Given

			// When

			// Then
		}, true)) // <-- flag to setup `t.Parallel()` on the test.
	}
}
```

[gock]: https://github.com/h2non/gock "Gock"
[monkey]: https://github.com/bouk/monkey "Monkey Patching"
