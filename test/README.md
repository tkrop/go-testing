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
        mocks := mock.NewMocks(t).Expect(
            test.Panic("fail"),
        )

        // When
        panic("fail")
    })(t)

    // Then
    ...
}
```


## Isolated parameterized parallel test runner

The `test` framework supports to run isolated, parameterized, parallel tests
using a lean test runner. The runner can be instantiated with a single test
parameter set (`New`), a slice of test parameter sets (`Slice`), or a map of
test case name to test parameter sets (`Map` - preferred pattern). The test is
started by `Run` that accepts a simple test function as input, using a
`test.Test` interface, that is compatible with most tools, e.g.
[`gomock`][gomock].


```go
func TestUnit(t *testing.T) {
    test.New|Slice|Map(t, testParams).
        /* Filter("test-case-name", false|true). */
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
parameter of type `test.Expect` (name `expect`) that supports `Failure` (false)
and `Success` (true - default).

Similar a test case name can be provided using type `test.Name` (name `name` -
default value `unknown-%d`) or as key using a test case name to parameter set
mapping.

**Note:** See [Parallel tests requirements](..#parallel-tests-requirements)
for more information on requirements in parallel parameterized tests. If
parallel parameterized test are undesired, `RunSeq` can be used to enforce a
sequential test execution.

It is also possible to select a subset of tests for execution by setting up a
`Filter` using a regular expression to match or filter by the normalized test
name.


## Isolated in-test environment setup

It is also possible to isolate only a single test step by setting up a small
test function that is run in isolation.

```go
func TestUnit(t *testing.T) {
    test.Map(t, testParams).
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

    test.NewTester(t, test.Success).Run(func(t test.Test){
        // Given

        // When

        // Then
    })(t)
}
```

But this should usually be unnecessary.


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
        // And one of the following
        t.Fatalf("fail")
        t.FailNow()
        panic("fail")

        // Then
    })(t)
}
```

**Note:** To enable panic testing, the isolated test environment is recovering
from all panics by default and converting it in a fatal error message. This is
often most usable and sufficient to fix the issue. If you need to discover the
source of the panic, you need to spawn a new unrecovered go-routine.

**Hint:** [`gomock`][gomock] uses very complicated reporting patterns that are
hard to recreate. Do not try it.


## Out-of-the-box test patterns

Currently, the package supports the following _out-of-the-box_ test pattern for
testing of `main`-methods of commands.

```go
testMainParams := map[string]test.MainParams{
    "no mocks": {
        Args:     []string{"mock"},
        Env:      []string{},
        ExitCode: 0,
    },
}

func TestMain(t *testing.T) {
    test.Map(t, testMainParams).Run(test.TestMain(main))
}
```

The pattern executes a the `main`-method in a separate process that allows to
setup the command line arguments (`Args`) as well as to modify the environment
variables (`Env`) and to capture and compare the exit code.

**Note:** the general approach can be used to test any code calling `os.Exit`,
however, it is focused on testing methods without arguments parsing command
line arguments, i.e. in particular `func main() { ... }`.


[gomock]: <https://github.com/golang/mock>
