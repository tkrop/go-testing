# Package testing/test

The [`test`][test] package provides a small but sophisticated framework to
isolate the test execution and safely check whether a test succeeds or fails as
expected. In combination with the [`mock`](../mock) package it ensures, that a
test finishes reliably and reports its failure, even if a system under test is
spawning go-routines or panics.

To accomplish this, the [`test`][test] package supports a lean common test
[`Factory`][factory] for parameterized test creating the common, isolating test
[`Context`][context] running in parallel. Besides functions for setting up
timeouts and cleaning up test resources, the test [`Factory`][factory] also
provides a [`Filter`][filter] to simplify the selection of test cases for a
specific test scenario.

The [`Factory`] can be instantiated by global functions with a single test
parameter set ([`test.Param`][param]), a slice of test parameter sets
([`test.Slice`][slice]), or a map of test case name to test parameter sets
([`test.Map`][map] - idiomatic pattern). The tests are started by calling the
[`Run`][factory], [`RunSeq`][factory], or [`Benchmark`][factory] methods that
with the exception of the last accept a simple test function as input, using a
[`test.Test`][itest] interface compatible with most other extensions, e.g.
[`gomock`][gomock].


[test]: <https://pkg.go.dev/github.com/tkrop/go-testing/test>
[param]: <https://pkg.go.dev/github.com/tkrop/go-testing/test#Param>
[slice]: <https://pkg.go.dev/github.com/tkrop/go-testing/test#Slice>
[map]: <https://pkg.go.dev/github.com/tkrop/go-testing/test#Map>
[itest]: <https://pkg.go.dev/github.com/tkrop/go-testing/test#Test>
[factory]: <https://pkg.go.dev/github.com/tkrop/go-testing/test#Factory>
[context]: <https://pkg.go.dev/github.com/tkrop/go-testing/test#Context>
[filter]: <https://pkg.go.dev/github.com/tkrop/go-testing/test#FilterFunc>


## Example usage

Use the following example to intercept and validate a panic using the isolated
test environment.

```go
func TestUnit(t *testing.T) {
    test.New(t).Run(func(t test.Test){
        // Given
        mock.NewMocks(t).Expect(
            test.Panic("fail"),
        )

        // When
        panic("fail")
    ...
    })
}
```

But there are many other supported use case, you can discover reading the
below examples.


## Isolated parameterized test setup

The most common usage to run an isolated, parameterized, parallel (or
sequential) tests using the lean test [`Factory`][factory] as follows:

```go
func TestUnit(t *testing.T) {
    // Set up the test using various test case definitions. No need to set up
    // parallel execution manually here, since this is done by default in the
    // by the test factory - when possible.
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

This creates and starts a lean test [`Context`][context] using the common
[`Test`][itest] interface, that isolates test execution and intercepts all
failures (including panics), to either forward or suppress them.

**Note:** See [Parallel tests requirements](..#parallel-tests-requirements)
for more information on requirements in parallel parameterized tests. If
parallel parameterized test are undesired, [`RunSeq`][runseq] can be used
to enforce a sequential test execution.


### Parameters and expectations

The test [`Context`] can be controlled by providing a test parameter of type
[`test.Expect`][expect] (idiomatic parameter name `expect`) that supports
[`test.Failure`][failure] (false) and [`test.Success`][success]
(true - default).

Similar a test case name can be provided using a `string` type with parameter
name `name` (default value `unknown-%d`) or as key using a test case name to
parameter set mapping.

[expect]: <https://pkg.go.dev/github.com/tkrop/go-testing/test#Expect>
[failure]: <https://pkg.go.dev/github.com/tkrop/go-testing/test#Expect>
[success]: <https://pkg.go.dev/github.com/tkrop/go-testing/test#Expect>


### Factory setup

The test [`Factory`][factory] setup allows to define a test specific
[`Timeout`][timeout] and a grace period to [`StopEarly`][stop-early] giving the
[`Cleanup`][cleanup]-functions sufficient time to free resources. In addition,
it is possible to (de-)select a subset of tests for a specific test execution
by setting up a highly customizable [`Filter`][filter] function - see
[Filter functions](#filter-functions) for more information.

Last the [`Factory`][factory] setup offers a method to systematically
[`Prefix`][prefix] test case names to structure tests into logic groups and
label them for statistical tools, e.g. [`benchstat`][benchstat].

[prefix]: <https://pkg.go.dev/github.com/tkrop/go-testing/test#Factory>
[benchstat]: <https://pkg.go.dev/golang.org/x/perf/cmd/benchstat>


### Filter functions

The [`test`][test] package supports the following default filter functions:

* [`test.All()`][all] — filter that filters all test cases.
* [`test.None()`][none] — filter that filters no test case at all.
* [`test.Not(filter)`][not] — filter for a logical `not` on a single wrapped
  filter.
* [`test.And(filter...)`][and] — filter for a logical `and` on a variable list
  of wrapped filters.
* [`test.Or(filter...)`][or] — filter for a logical `or` on a variable list of
  wrapped filters.
* [`Xor(filter...)`][xor] — filter for a logical `xor` (exclusive `or`) on a
  variable list of wrapped filters.
* `test.Implies(filter...)` — a convenience filter for a logical implication.
  It can also be expressed by `Or(Not(filter),And(filter...))`.
* `test.Pattern(regex)` — filter for selecting test cases by normalized test
  case names using a regular expression.
* `test.OS(name)` — filter for selecting operating system specific test cases
  by system name.
* `test.Arch(name)` — filter for selecting processor architecture specific test
  cases by architecture name.

[timeout]: <https://pkg.go.dev/github.com/tkrop/go-testing/test#Factory>
[stop-early]: <https://pkg.go.dev/github.com/tkrop/go-testing/test#Factory>
[cleanup]: <https://pkg.go.dev/github.com/tkrop/go-testing/test#Factory>


## Manual isolated test setup

You have two options to set up a manual isolated test environment. The first
option is to create standard test functions for the default test runner using
the [`test.Run`][run] or [`test.RunSeq`][runseq] as follows:

```go
func TestUnit(t *testing.T) {
    // Need to set up parallel execution manually here.
    t.Parallel()

    for name, param := range unitTestCases {
        t.Run(name, test.Run|test.RunSeq(func(t test.Tester) {
            // Set up sequential test execution (default is test.Parallel).
            t.Mode(test.Parallel|test.Sequential)
            // Set up the test expectation (default is test.Success).
            t.Expect(test.Success|test.Failure)
            // Define a test specific timeout (default: none).
            t.Timeout(50*time.Millisecond)
            // Define a safety margin for cleaning up (default: none).
            t.StopEarly(5*time.Millisecond)

            // Given

            // When

            // Then
        }))
    }
}
```

To allow for more flexibility the test function is provided with an extended
[`test.Tester`][tester] interface, that allows for additional test control,
e.g. setting up the test expectation and execution mode, defining timeouts and
safety margins for cleaning up.


## Manual isolated test context setup

If this pattern is insufficient, and you need more control about the test
execution, you can also create your own customized, parallel, isolated test
wrapper based on the common [`test.Context`][context]. It extends the basic
[`test.Test`][test] interface abstraction and can be utilized as follows:

```go
func TestUnit(t *testing.T) {
    t.Parallel()

    test.New(t).
        // Set up sequential test execution (default is test.Parallel).
        Mode(test.Parallel|test.Sequential).
        // Set up the test expectation (default is test.Success).
        Expect(test.Success|test.Failure).
        // Define a test specific timeout (default: none).
        Timeout(50*time.Millisecond).
        // Define a safety margin for cleaning up (default: none).
        StopEarly(5*time.Millisecond).
        // Run the test function.
        Run("test-name", func(t test.Test){
            // Given

            // When

            // Then
        })
}
```

[run]: <https://pkg.go.dev/github.com/tkrop/go-testing/test#Run>
[runseq]: <https://pkg.go.dev/github.com/tkrop/go-testing/test#RunSeq>
[tester]: <https://pkg.go.dev/github.com/tkrop/go-testing/test#Tester>


## Isolated in-test environment setup

Using the [before pattern](#manual-isolated-test-context-setup) it is also
possible to isolate only a single test step by setting up a small test
function that is executed in isolation:

```go
func TestUnit(t *testing.T) {
    test.Param|Slice|Map|Any(t, unitTestCases).
        ...
        // Run the test in parallel or sequential.
        Run|RunSeq(func(t test.Test, param UnitParams){
            // Given

            // When
            test.New(t).Run(func(t test.Test) {
                ...
            })

            // Then
        })
}
```


## Isolated failure/panic validation

Besides just capturing the failure in the isolated test environment, it is also
possible to easily validate failures/panics using the self installing validator
that is tightly integrated with the [`mock`](../mock) framework.

```go
func TestUnit(t *testing.T) {
    test.New(t).Run(func(t test.Test){
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

1. [`test.Main(func())`](#main-method-test-pattern) - allows to test main
   methods by calling the main method with arguments in a well controlled test
   environment.
2. [`test.Recover(Test,any)`](#recover-test-pattern) - allows to check the
   panic result in simple test scenarios where [`test.Panic(any)`][panic] is
   not applicable.


### Main method test pattern

The [`test.Main(func())`][main] pattern executes the `main` method in a
separate test process to protect the test execution against `os.Exit` calls
while allowing to capture and check the exit code against the expectation. The
following example demonstrates how to use the pattern to test a `main` method:

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

[main]: <https://pkg.go.dev/github.com/tkrop/go-testing/test#Main>


### Recover test pattern

The [`test.Recover(Test,any)`][recover] test pattern is a very specific, rare
pattern that can be applied as follows:

```go
func TestPanic(t *testing.T) {
    // Given
    defer test.Recover(t, "gock not supported by test setup")

    // When
    gock.NewGock(gomock.NewController(struct{ gomock.TestReporter }{}))

    // Then
    assert.Fail(t, "did not panic")
}
```


## Parameterized benchmark setup

The [`test`][test] package also supports a consistent pattern for setting up
parameterized benchmarks with two minor changes:

1. Since `*testing.B` is missing a small number of functions of the
   [`test.Test`][itest] interface abstraction, it must be wrapped using
   [`test.Benchmark`][benchmark].
2. Since running of benchmarks is slightly different, the benchmark is executed
   using [`Benchmark`][bench] supporting the special two-phase parameterized
   [`BenchmarkFunc`][bench-func] function needed in this case. The first phase
   is used for setup, while the second is used to run the benchmark loop.

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
[`Prefix`][prefix]-pattern may become very handy for you:

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

[benchmark]: <https://pkg.go.dev/github.com/tkrop/go-testing/test#Benchmark>
[bench-func]: <https://pkg.go.dev/github.com/tkrop/go-testing@/test#BenchmarkFunc>


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
