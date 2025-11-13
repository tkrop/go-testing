# Package testing/test

The goal of this package is to provide a small framework to isolate the test
execution and safely check whether a test succeeds or fails as expected. In
combination with the [`mock`](../mock) package it ensures that a test finishes
reliably and reports its failure even if a system under test is spawning
go-routines.


## Example usage

Use the following example to intercept and validate a panic using the isolated
test environment.

```go
func TestUnit(t *testing.T) {
    test.Run(test.Success, func(t test.Test){
        // Given
        mock.NewMocks(t).Expect(
            test.Panic("fail"),
        )

        // When
        panic("fail")
    ...
    })(t)
}
```

But there are many other supported use case, you can discover reading the
below examples.


## Isolated parameterized parallel test runner

The `test` framework supports to run isolated, parameterized, parallel tests
using a lean test runner. The runner can be instantiated with a single test
parameter set (`test.Param`), a slice of test parameter sets (`test.Slice`), or
a map of test case name to test parameter sets (`test.Map` - preferred pattern).
The test is started by `Run` that accepts a simple test function as input,
using a `test.Test` interface, that is compatible with most tools, e.g.
[`gomock`][gomock].


```go
func TestUnit(t *testing.T) {
    test.Param|Slice|Map|Any(t, unitTestCases).
        Filter("test-case-name", false|true).
        Timeout(5*time.Millisecond).
        StopEarly(time.Millisecond).
        Run|RunSeq(func(t test.Test, param UnitParams){
            // Given

            // When

            // Then
        }).Cleanup(func(){
            // clean test resources
        })
}
```

This creates and starts a lean test wrapper using a common interface, that
isolates test execution and intercepts all failures (including panics), to
either forward or suppress them. The result is controlled by providing a test
parameter of type `test.Expect` (name `expect`) that supports `test.Failure`
(false) and `Success` (true - default).

Similar a test case name can be provided using type `test.Name` (name `name` -
default value `unknown-%d`) or as key using a test case name to parameter set
mapping.

**Note:** See [Parallel tests requirements](..#parallel-tests-requirements)
for more information on requirements in parallel parameterized tests. If
parallel parameterized test are undesired, `RunSeq` can be used to enforce a
sequential test execution.

It is also possible to select a subset of tests for execution by setting up a
`Filter` using a regular expression to match or filter by the normalized test
name, or to set up a `Timeout` as well as a grace period to `StopEarly` for
giving the `Cleanup`-functions sufficient time to free resources.


## Isolated in-test environment setup

It is also possible to isolate only a single test step by setting up a small
test function that is run in isolation.

```go
func TestUnit(t *testing.T) {
    test.Map(t, unitTestCases).
        Run|RunSeq(func(t test.Test, param UnitParams){
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
`test.Run|RunSeq(test.Success|Failure, func (t test.Test) {})`:

```go
func TestUnit(t *testing.T) {
    t.Parallel()

    for name, param := range unitTestCases {
        t.Run(name, test.Run(param.expect, func(t test.Test) {
            t.Parallel()

            // Given

            // When

            // Then
        }))
    }
}
```

Or finally, use even more directly the flexible `test.Context` that is
providing the features on top of the underlying `test.Test` interface
abstraction, if you need more control about the test execution:

```go
func TestUnit(t *testing.T) {
    t.Parallel()

    test.New(t, test.Success).
        Timeout(5*time.Millisecond).
        StopEarly(time.Millisecond).
        Run(func(t test.Test){
            // Given

            // When

            // Then
        })(t)
}
```


## Isolated failure/panic validation

Besides just capturing the failure in the isolated test environment, it is also
very simple possible to validate the failures/panics using the self installing
validator that is tightly integrated with the [`mock`](../mock) framework.

```go
func TestUnit(t *testing.T) {
    test.Run(func(t test.Test){
        // Given
        mock.NewMocks(t).Expect(mock.Setup(
            test.Errorf("fail"),
            test.Fatalf("fail"),
            test.FailNow(),
            test.Panic("fail"),
        ))

        // When
        t.Errorf("fail")
        ...
        // And one of the terminal calls.
        t.Fatalf("fail")
        t.FailNow()
        panic("fail")

        // Then
    })(t)
}
```

**Note:** To enable panic testing, the isolated test environment is recovering
from all panics by default and converting them in fatal error messages. This is
often most usable and sufficient to fix the issue. If you need to discover the
source of the panic, you need to spawn a new unrecovered go-routine.

**Hint:** [`gomock`][gomock] uses very complicated reporting patterns that are
hard to recreate. Do not try it.


## Out-of-the-box test patterns

Currently, the package supports two _out-of-the-box_ test patterns:

1. `test.Main(func())` - allows to test main methods by calling the main
   method with arguments in a well controlled test environment.
2. `test.Recover(Test,any)` - allows to check the panic result in simple test
   scenarios where `test.Panic(any)` is not applicable.


### Main method tests pattern

The `test.Main(func())` pattern executes the `main` method in a separate test
process to protect the test execution against `os.Exit` calls while allowing to
capture and check the exit code against the expectation. The following example
demonstrates how to use the pattern to test a `main` method:

```go
mainTestCases := map[string]test.MainParams{
    "no mocks": {
        Args: []string{"mock", "arg1", "arg2"},
        Env: []string{"VAR=value"},
        ExitCode: 0,
    },
}

func TestMain(t *testing.T) {
    test.Map(t, mainTestCases).Run(test.TestMain(main))
}
```

If the test process is expected to run longer than the default test timeout, a
context with timeout can be provided to interrupt the test process in time,
e.g. as follows:

```go
    Ctx: test.First(context.WithTimeout(context.Bachground(), time.Second))
```

**Note:** the general approach can be used to test any code calling `os.Exit`,
however, it is focused on testing the `main` methods with and without parsing
command line arguments.

**Note:** In certain situations, `test.Main(func())` currently fails to obtain
the coverage metrics for the test execution, since `go test` is using the
standard output to collect results. We are investigating how we can separate
these in the test execution from expected test output.


## Convenience functions

The test package contains a number of convenience functions to simplify the
test setup and apply certain test patterns. Currently, the following functions
currently supported:

* `test.Must[T](T, error) T` - a convenience method for fluent test case setup
  that converts an error into a panic.
* `test.Cast[T](T) T` - a convenience method for fluent test case setup that
  converts an casting error into a panic compliant with linting requirements.
* `test.Ptr[T](T) *T` - a convenience method for fluent test case setup that
  converts a literal value into a pointer.
* `test.First[T](T, ...any)` - a convenience method for fluent test case setup
  that extracts the first value of a response ignoring the others.

Please also have a look at the convenience functions provided by the
[reflect](../reflect) package, that allows you to fluently access non-exported
fields for setting up and checking.


[gomock]: <https://go.uber.org/mock>
