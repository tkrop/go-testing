# Usage patterns of go-base/mock

This package contains a small extension library to handle mock call setup in a
standardized way. Unfortunately, we had to sacrifice type-safety to allow for
chaining mock calls in an arbitrary way during setup. Anyhow, in tests runtime
validation is likely a sufficient strategy.


## Generic mock setup pattern

To setup a generic mock handler for any number of mocks, one can simply use the
following template to setup an arbitrary system under test.

```go
func SetupTestUnit(
    t *gomock.TestReporter,
    mockSetup mock.SetupFunc,
    ...
) (*Unit, *Mocks) {
    mocks := mock.NewMock(t).Expect(mockSetup)

    unit := NewUnitService(
        mock.Get(mocks, NewServiceMock)
    ).(*Unit)

    return unit, mocks
}
```

**Note:** The `mock.Get(mocks, NewServiceMock)` is the standard pattern to
request an existing or new mock instance from the mock handler.


## Generic mock service call pattern

Now we need to define the mock service calls that follow a primitive, common
coding and naming pattern, that may be supported by code generation in the
future.

```go
func ServiceCall(input..., output..., error) mock.SetupFunc {
    return func(mocks *Mocks) any {
        mocks.WaitGroup().Add(1)
        return Get(mocks, NewServiceMock).EXPECT().
            ServiceCall(input...).Return(output..., error).Times(1).
            Do(func(input... interface{}) {
                defer mocks.WaitGroup().Done()
            })

    }
}
```

For simplicity the pattern combines regular as well as error behavior and is
prepared to handle tests with detached *goroutines*, however, this code can
be left out for further simplification.

For detached *goroutines*, i.e. functions that do not communicate with the
test, the mock handler provides a `WaitGroup` to registers expected mock calls
using `mocks.WaitGroup().Add(<times>)` and notifying the occurrence by calling
`mocks.WaitGroup().Done()` in as a mock callback function registered for the
match via `Do(<func>)`. The test waits for the detached *goroutines* to finish
by calling `mocks.WaitGroup().Wait()`.

A static series of mock service calls can now simply expressed by chaining the
mock service calls as follows using `mock.Chain` and while defining a new mock
call setup function:

```go
func ServiceCallChain(input..., output..., error) mock.SetupFunc {
    return func(mocks *Mocks) any {
        return mock.Chain(
            ServiceCallA(input...),
            ServiceCallB(input...),
            ...
    }
}
```


## Generic mock ordering patterns

With the above preparations for mocking service calls we can now define the
*mock setup* easily  using the following ordering methods:

* `Chain` allows to create an ordered chain of mock calls that can be combined
  with other setup methods that defermine the predecessors and successor mock
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
    ...
    expect*      *model.*
    expectError  error
}{
    "single mock setup": {
        mockSetup: ServiceCall(...),
    }
    "chain mock setup": {
        mockSetup: mock.Chain(
            ServiceCallA(...),
            ServiceCallB(...),
            ...
        )
    }
    "nested chain mock setup": {
        mockSetup: mock.Chain(
               ServiceCallA(...),
            mock.Chain(
                ServiceCallA(...),
                ServiceCallB(...),
                ...
            ),
               ServiceCallB(...),
               ...
        )
    }
    "parallel chain mock setup": {
        mockSetup: mock.Parallel(
            ServiceCallA(...),
            mock.Chain(
                ServiceCallB(...),
                ServiceCallC(...),
                ...
            ),
            mock.Chain(
                ServiceCallD(...),
                ServiceCallE(...),
                ...
            ),
            ...
        )
    }
    ...
}
```

This test parameter setup can now be use for a parameterized unit test using
the following common pattern:

```go
func TestUnitCall(t *testing.T) {
    for message, param := range testUnitCallParams {
        t.Run(message, func(t *testing.T) {
            require.NotEmpty(t, message)

            //Given
            unit, mocks := SetupTestUnit(t, param.mockSetup)

            //When
            result, err := unit.UnitCall(...)

            mock.WaitGroup().Wait()

            //Then
            if param.expectError != nil {
                assert.Equal(t, param.expectError, err)
            } else {
                require.NoError(t, err)
            }
            assert.Equal(t, param.expect*, result)
        })
    }
}
```
