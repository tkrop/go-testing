# Package testing/test

The goal of this package is to provide a small framework to isolate the test
execution and safely check whether a test succeeds or fails as expected. In
combination with the [mock](../mock) package it ensures that a test finishs
reliably and reports its failure even if a system under test is spawning
go-routines.


## Example usage

TODO


## Isolated parameterized parallel test runner

The `test` framework supports to run isolated, parameterized, parallel tests
using a lean test runner. The runner can be instantiated with a single test
parameter set (`New`), a slice of test parameter sets (`Slice`), or a map of
test case name to test parameter sets (`Map` - preferred pattern). The test is
started by `Run` that accepts a simple test function as input, using a
`test.Test` interface, that is compatible with most tools, e.g.
[Gomock][gomock].


```go
func TestUnit(t *testing.T) {
    test.New|Slice|Map(t, testParams).
        Run(func(t test.Test, param UnitParams){
            // Given

            // When

            // Then
        })
}
```

This creates and starts a lean test wrapper using a common interface, that
isolates test execution and intercepts all failures (including panics), to
either forward or suppress them. The result is controlled by providing a test
parameter of type `test.Expect` (name `expect`) that supports `Failure` (false)
and `Success` (true - default).

Similar a test case name can be provided using type `test.Name` (name `name` -
default value `unknown-%d`) or as key using a test case name to parameter set
mapping.

**Note:** See [Parallel tests requirements](..#parallel-tests-requirements)
for more information on requirements in parallel parameterized tests.


## Isolated in-test environment setup

It is also possible to isolate only a single test step by setting up a small
test function that is run in isolation.

```go
func TestUnit(t *testing.T) {
    test.Map(t, testParams).
        Run(func(t test.Test, param UnitParams){
            // Given

            // When
            test.InRun(test.Failure, func(t test.Test) {
                ...
            })(t)

            // Then
        })
}
```


## Manual isolated test environment setup

If the above pattern is not sufficient, you can create your own customized
parameterized, parallel, isolated test wrapper using the basic abstraction
`test.Run(test.Success|Failure, func (t test.Test) {})`:

```go
func TestUnit(t *testing.T) {
    t.Parallel()

    for name, param := range testParams {
        name, param := name, param
        t.Run(name, test.Run(param.expect, func(t test.Test) {
            t.Parallel()

            // Given

            // When

            // Then
        }))
    }
}
```

Or the interface of the underlying `test.Tester`:

```go
func TestUnit(t *testing.T) {
    t.Parallel()

    test.Tester(t, test.Success).Run(func(t test.Test){
        // Given

        // When

        // Then
    })(t)
}
```

But this should usually be unnecessary.


## Isolated failure validation

Asides just capturing the failure in the isolated test environment, it is also
possible to validate the test failures using a self installing validator, that
is integrated with the [mock](../mock) framework.

```go
func TestUnit(t *testing.T) {
    test.Run(func(t test.Test){
        // Given
        mocks := mock.NewMock(t).Expect(test.Fatalf("fail"))

        // When
        t.Fatalf("fail")

        // Then
    })(t)
}
```

**Note:** Test failures are often use very complicated reporting patterns,
e.g. in [GoMock][gomock].


[gomock]: https://github.com/golang/mock "GoMock"
