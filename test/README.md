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
    ...
    })(t)
}
```

But there are many other supported use case, you can discover reading the
below examples.


## Isolated parameterized parallel test runner

The `test` framework supports to run isolated, parameterized, parallel tests
using a lean test runner. The runner can be instantiated with a single test
parameter set (`test.Any`), a slice of test parameter sets (`test.Slice`), or a
map of test case name to test parameter sets (`test.Map` - preferred pattern).
The test is started by `Run` that accepts a simple test function as input,
using a `test.Test` interface, that is compatible with most tools, e.g.
[`gomock`][gomock].


```go
func TestUnit(t *testing.T) {
    test.Any|Slice|Map(t, testParams).
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
`test.Run|RunSeq(test.Success|Failure, func (t test.Test) {})`:

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


## Test result builder

Setting up tests and comparing test results is most efficient, when you
directly can set up and compare the actual objects. However, this is sometimes
prevented by the objects not being open for construction and having private
states.

To cope with this challenge the `test`-package supports helpers to access, i.e.
read and write, private fields of objects using reflection.

* `test.NewBuilder[...]()` allows constructing a new object from scratch.
* `test.NewGetter(...)` allows reading private fields of an object by name.
* `test.NewSetter(...)` allows writing private fields by name, and finally
* `test.NewAccessor(...)` allows reading and writing of private fields by name.

The following example shows a real world example of how the private properties
of a closed error can be set up using the `test.NewBuilder[...]()`.

```go
    err := test.NewBuilder[viper.ConfigFileNotFoundError]().
        Set("locations", fmt.Sprintf("%s", "...path...")).
        Set("name", "test").Build()
```

Similar we can set up input objects with private properties to minimize the
dependencies in the test setup, however, using this features exposes the test
to internal changes.


## Out-of-the-box test patterns

Currently, the package supports only one _out-of-the-box_ test pattern to test
the `main`-methods of commands.

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

The pattern executes the `main`-method in a separate process that setting up
the command line arguments (`Args`) and modifying the environment variables
(`Env`) and to capture and compare the exit code of the program execution.

**Note:** the general approach can be used to test any code calling `os.Exit`,
however, it is focused on testing the `main`-methods with and without parsing
command line arguments.


[gomock]: <https://github.com/golang/mock>
