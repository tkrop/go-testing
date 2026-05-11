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


## Isolated parameterized test setup

The `test` framework supports to run isolated, parameterized, parallel tests
using a lean test runner. The runner can be instantiated with a single test
parameter set (`test.Param`), a slice of test parameter sets (`test.Slice`), or
a map of test case name to test parameter sets (`test.Map` - preferred pattern).
The test is started by `Run` or `RunSeq` that accepts a simple test function as
input, using a `test.Test` interface, that is compatible with most tools, e.g.
[`gomock`][gomock].


```go
func TestUnit(t *testing.T) {
    // Set up the test using various test case definitions.
    test.Param|Slice|Map|Any(t, unitTestCases).
        // Exclude of test cases temporary or permanent.
        Filter(test.Not(test.Pattern[T]("^test-case-prefix"))).
        // Include of test cases temporary or permanent.
        Filter(test.Pattern[T]("^test-case-name$")).
        // Define a test specific timeout.
        Timeout(50*time.Millisecond).
        // Define a safety margin for cleaning up.
        StopEarly(5*time.Millisecond).
        // Run the test in parallel (or sequential).
        Run|RunSeq(func(t test.Test, param UnitParams){
            // Given

            // When

            // Then
        }).Cleanup(func(){
            // clean test resources
        })
}
```

<!-- TODO: create sub-section with extended explanation on variables -->

This creates and starts a lean test wrapper using a common interface, that
isolates test execution and intercepts all failures (including panics), to
either forward or suppress them. The result is controlled by providing a test
parameter of type `test.Expect` (name `expect`) that supports `test.Failure`
(false) and `test.Success` (true - default).

Similar a test case name can be provided using type `test.Name` (name `name` -
default value `unknown-%d`) or as key using a test case name to parameter set
mapping.

**Note:** See [Parallel tests requirements](..#parallel-tests-requirements)
for more information on requirements in parallel parameterized tests. If
parallel parameterized test are undesired, `RunSeq` can be used to enforce a
sequential test execution.

<!-- TODO: create sub-section with extended explanation on filter functions -->

The setup allows to define a test specific `Timeout` and a grace period to
`StopEarly` giving the `Cleanup`-functions sufficient time to free resources.
In addition, it is possible to (de-)select a subset of tests for execution by
setting up a highly customizable `Filter` function that provides the following
default implementations:

<!-- TODO: create sub-section with extended explanation on filter functions -->

* `test.None|All` — convenience filters that filter nothing/all.
* `test.Not(filter)` — for negating logic operation of filter function.
* `test.And|Or|Xor(filter...)` — for conducting logical operations on filter
  functions.
* `test.Implies` — a convenience filter for a logical implication. It can also
  be expressed by `Or(Not(filter),And(filter...))`.
* `test.Pattern(name)` — for selecting test cases by normalized names using a
  regular expression.
* `test.OS(name)` — for selecting operating system specific test by system name.
* `test.Arch(name)` — for selecting processor architecture specific test cases
  by architecture name.


## Isolated in-test environment setup

It is also possible to isolate only a single test step by setting up a small
test function that is run in isolation.

```go
func TestUnit(t *testing.T) {
    test.Param|Slice|Map|Any(t, unitTestCases).
        ...
        // Run the test in parallel or sequential.
        Run|RunSeq(func(t test.Test, param UnitParams){
            // Given

            // When
            test.InRun(test.Success|Failure, func(t test.Test) {
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

    test.New(t, test.Success|Failure).
        // Define a test specific timeout.
        Timeout(50*time.Millisecond).
        // Define a safety margin for cleaning up.
        StopEarly(5*time.Millisecond).
        // Run the test function.
        Run("test", func(t test.Test){
            // Given

            // When

            // Then
        })
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


## Parameterized benchmark setup

The `test` framework also supports a consistent pattern for setting up
parameterized benchmarks with two minor changes:

1. Since `*testing.B` is missing a small number of functions of the `test.Test`
   interface abstraction, it must be wrapped using `test.Benchmark(b)`.
2. Since running of benchmarks is slightly different, the benchmark is executed
   using `Benchmark(BenchmarkFunc[P])` supporting the two-phase parameterized
   benchmark function. The first phase is used for setup, while the second is
   used to run the benchmark loop.

The full parameterized benchmark setup example looks as follows, and can make
use of the same features as the regular parameterized test setup:

```go
func BenchmarkUnit(b *testing.B) {
    test.Map(test.Benchmark(b), unitTestCases).
        // Exclude of test cases temporary or permanent.
        Filter(test.Not(test.Pattern[T]("^test-case-prefix"))).
        // Include of test cases temporary or permanent.
        Filter(test.Pattern[T]("^test-case-name$")).
        // Execute benchmark setup and loop phases.
        Benchmark(func(b *testing.B, param UnitParams) func(b *testing.B) {
            // Setup
            unit := NewUnit(param.input*...)

            // Define processed bytes.
            b.SetBytes(len(param.input*))

            // Loop
            return func(b *testing.B) {
                result, err := unit.call(param.input*...)

                // Prevent optimization.
                runtime.KeepAlive(result)
                runtime.KeepAlive(err)
            }
        })
}
```

**Note:** in a benchmark you need to ensure that you reserve sufficient memory
for the `unit`-under-test in the setup phase to avoid additional memory allocs
in the loop. While you also should prevent return values from being optimized
away in the loop using `runtime.KeepAlive`, you should not do this for
multi-byte results, since these also creates additional memory allocations due
the the copy nature of the `runtime.KeepAlive`.

If you want to compare and analyse the performance of functions with the same
signature and same parameter set using [benchstat][benchstat], the following
`Prefix`-pattern may become very handy for you:

```go
func benchmarkUnit(
    b *testing.B, string name,
    call func(*UnitService, <input>...) (<output>...),
) {
    test.Map(test.Benchmark(b), unitTestCases).
        // Add a label prefix to the benchmark name.
        Prefix("method="+name + "/test=")
        // Execute benchmark setup and loop phases.
        Benchmark(func(b *testing.B, param UnitParams) func(b *testing.B) {
            // Setup
            unit := NewUnit(param.input*...)

            // Define processed bytes.
            b.SetBytes(len(param.input*))

            // Loop
            return func(b *testing.B) {
                result, err := unit.call(param.input*...)

                // Prevent optimization.
                runtime.KeepAlive(result)
                runtime.KeepAlive(err)
            }
        })
}

func BenchmarkUnit(b *testing.B) {
    benchmarkUnit(b, "call-a", (*Unit).callA))
    benchmarkUnit(b, "call-b", (*Unit).callB))
}
```

It allows you to analyse the performance of your alternative functions using
the following [benchstat][benchstat] command line:

```bash
benchstat -row /test -col /method file.bench
```

[benchstat]: <https://pkg.go.dev/golang.org/x/perf/cmd/benchstat>


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
