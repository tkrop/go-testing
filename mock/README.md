# Package testing/mock

Goal of this package is to provide a small extension library that provides a
common mock controller interface for [`gomock`][gomock] and [`gock`][gock] that
enables a unified, highly reusable integration pattern.

Unfortunately, we had to sacrifice a bit of type-safety to allow for chaining
mock calls in an arbitrary way during setup. Anyhow, the offered in runtime
validation is a sufficient strategy to cover for the missing type-safety.


## Example usage

The `mock`-framework provides a simple [`gomock`][gomock] handler extension to
creates singleton mock controllers on demand by accepting mock constructors in
its method calls. In addition, it provides a mock setup abstraction to simply
setup complex mock request/response chains.

```go
func TestUnit(t *testing.T) {
    // Given
    mocks := mock.NewMocks(t)

    mockSetup := mock.Get(mocks, NewServiceMock).EXPECT()...
    mocks.Expect(mockSetup)

    service := NewUnitService(
        mock.Get(mocks, NewServiceMock))

    // When
    ...
}
```

Using the features of the `mock`-framework we can design more advanced usage
patterns as described in the following.


## Generic mock controller setup

Usually, a new system under test must be create for each test run. Therefore,
the following generic pattern to setup the mock controller with an arbitrary
system under test is very useful.

```go
func SetupUnit(
    t test.Test,
    mockSetup mock.SetupFunc,
) (*Unit, *Mocks) {
    mocks := mock.NewMocks(t).Expect(mockSetup)

    unit := NewUnitService(
        mock.Get(mocks, NewServiceMock)
    ).(*Unit)

    return unit, mocks
}
```

**Note:** The `mock.Get(mocks, NewServiceMock)` is the standard pattern to
request a new or existing mock instance from the mock controller. As input, any
test interface or entity compatible with the [`gomock.TestReporter`][gomock-rep]
can be used.


## Generic mock call setup

Now we need to define the mock service inline or better via a function calls
following the below common coding and naming pattern, that we may support by
code generation in the future.

```go
func Call(input..., output..., error) mock.SetupFunc {
    return func(mocks *mock.Mocks) any {
        return mock.Get(mocks, NewServiceMock).EXPECT().Call(input...).
            { DoAndReturn(mocks.Do(Service.Call, output..., error))
            | Return(output..., error).Do(mocks.Do(Service.Call)) }
        ]
    }
}
```

The pattern combines regular as well as error behavior and is out-of-the-box
prepared to handle tests with detached *goroutines*, i.e. functions that are
spawned by the system-under-test without waiting for their result.

The mock handler therefore provides a `WaitGroup` and automatically registers
a single mock call on each request using `mocks.Do(...)` and notifies the
completion via `Do|DoAndReturn()`. For test with detached *goroutines* the
test can wait via `mocks.Wait()`, before finishing and checking whether the
mock calls are completely consumed.

**Note:** Since waiting for mock calls can take literally for ever in case of
test failures, it is advised to use an isolated [test environment](../test)
that unlocks the waiting test in case of failures and fatal errors, e.g.
by using:

```go
test.Run(test.Success, func(t *TestingT) {
    // Given
    ...

    // When
    ...
    mocks.Wait()

    // Then
})
```

A static series of mock service calls can now simply expressed by chaining the
mock service calls as follows using `mock.Chain` and while defining a new mock
call setup function:

```go
func CallChain(input..., output..., error) mock.SetupFunc {
    return func(mocks *Mocks) any {
        return mock.Chain(
            CallA(input...),
            CallB(input...),
            ...
    }
}
```

**Note:** As a special test case it is possible to `panic` as mock a result by
using `Do(mocks.GetPanic(<#input-args>,<reason>))`.


## Generic mock ordering patterns

With the above preparations for mocking service calls we can now define the
*mock setup* easily  using the following ordering methods:

* `Chain` allows to create an ordered chain of mock calls that can be combined
  with other setup methods that determine the predecessors and successor mock
  calls.

* `Parallel` allows to creates an unordered set of mock calls that can be
  combined with other setup methods that determine the predecessor and
  successor mock calls.

* `Setup` allows to create an unordered detached set of mock calls that creates
  no relation to predecessors and successors it was defined with.

Beside this simple (un-)ordering methods there are two further methods for
completeness, that allow to control how predecessors and successors are used
to setup ordering conditions:

* `Sub` allows to define a sub-set or sub-chain of elements in `Parallel` and
  `Chain` as predecessor and successor context for further combination.

* `Detach` allows to detach an element from the predecessor context (`Head`),
  from the successor context (`Tail`), or from both which is used in `Setup`.

The application of these two functions may be a bit more complex but still
follows the intuition.


## Generic parameterized test pattern

The ordering methods and the mock service call setups can now be used to define
the mock call expectations, in a parameter setup as follows to show the most
common use cases:

```go
var testUnitCallParams = map[string]struct {
    mockSetup    mock.SetupFunc
    input*...    *model.*
    expect       test.Expect
    expect*...   *model.*
    expectError  error
}{
    "single mock setup": {
        mockSetup: Call(...),
    }
    "chain mock setup": {
        mockSetup: mock.Chain(
            CallA(...),
            CallB(...),
            ...
        )
    }
    "nested chain mock setup": {
        mockSetup: mock.Chain(
            CallA(...),
            mock.Chain(
                CallA(...),
                CallB(...),
                ...
            ),
            CallB(...),
            ...
        )
    }
    "parallel chain mock setup": {
        mockSetup: mock.Parallel(
            CallA(...),
            mock.Chain(
                CallB(...),
                CallC(...),
                ...
            ),
            mock.Chain(
                CallD(...),
                CallE(...),
                ...
            ),
            ...
        )
    }
    ...
}
```

This test parameter setup can now be use for all parameterized unit test using
the following common parallel pattern, that includes `mocks.Wait()` to handle
detached *goroutines* as well as the isolated [test environment](../test) to
unlocks the waiting group in case of failures:

```go
func TestUnitCall(t *testing.T) {
    t.Parallel()

for message, param := range testUnitCallParams {
        message, param := message, param
        t.Run(message, test.Run(param.expect, func(t test.Test) {
            t.Parallel()

            // Given
            unit, mocks := SetupTestUnit(t, param.mockSetup)

            // When
            result, err := unit.UnitCall(param.input*, ...)

            mocks.Wait()

            // Then
            if param.expectError != nil {
                assert.Equal(t, param.expectError, err)
            } else {
                require.NoError(t, err)
            }
            assert.Equal(t, param.expect*, result)
            ...
        }))
    }
}
```

**Note:** See [Parallel tests requirements](..#parallel-tests-requirements)
for more information on requirements in parallel parameterized tests.


[gomock]: <https://github.com/golang/mock>
[gomock-rep]: <https://github.com/golang/mock/blob/v1.6.0/gomock/controller.go#L65>
[gock]: <https://github.com/h2non/gock>
