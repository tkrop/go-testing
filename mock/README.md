# Usage patterns of testing/mock

This package contains a small extension library to handle mock call setup in a
standardized, highly reusable way. Unfortunately, we had to sacrifice a bit of
type-safety to allow for chaining mock calls in an arbitrary way during setup.
Anyhow, the offered in runtime validation is a sufficient strategy to cover
for the missing type-safty.


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
        return mock.Get(mocks, NewServiceMock).EXPECT().
            ServiceCall(input...).Return(output..., error).
			Times(mocks.Times(1)).Do(mocks.GetDone(<#input-args>))
    }
}
```

For simplicity the pattern combines regular as well as error behavior and is
prepared to handle tests with detached *goroutines*, i.e. functions that are
spawned by the system-under-test without waiting for their result.

The mock handler therefore provides a `WaitGroup` and automatically registers
the expected mock calls via `mock.Times(<#>)` and notifies the completion via
`Do(mocks.GetDone(<#input-args>))`. The test needs to wait for the detached
*goroutines* to finish by calling `mocks.Wait()` before checking whether mock
calls are completely consumed.

**Note:** Since waiting for mock calls can take unitl the test timeout appears
in case of test failures, you need to tests using `mocks.Wait()` in an isolated
[test environment](../test) that unlocks the waiting test in case of failures
and fatal errors using:

```go
test.Success(func(t *TestingT) {
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

This test parameter setup can now be use for all parameterized unit test using
the following common pattern, that includes the optional `mocks.Wait()` for
detached *goroutines* as well as the isoated [test environment](../test) to
unlocks the waiting test in case of failures:

```go
func TestUnitCall(t *testing.T) {
    for message, param := range testUnitCallParams {
        t.Run(message, test.Success(func(t *testing.T) {
            //Given
            unit, mocks := SetupTestUnit(t, param.mockSetup)

            //When
            result, err := unit.UnitCall(...)

            mocks.Wait()

            //Then
            if param.expectError != nil {
                assert.Equal(t, param.expectError, err)
            } else {
                require.NoError(t, err)
            }
            assert.Equal(t, param.expect*, result)
        }, false))
    }
}
```


## Paralle parameterized test pattern

Generally, the [test](../test) and [mock](.) framework supports parallel test
execution, however, there are some general requirements for running tests in
parallel:

1. The tests *must not modify* environment variables dynamically.
2. The tests *must not require* reserved service ports and open listeners.
3. The tests *must not use* [Gock][gock] for mocking on HTTP transport level
   (please use the internal [gock](../gock)-controller package),
4. The tests *must not use* [monkey patching][monkey] to modify commonly used
   functions, e.g. `time.Now()`, and finaly
5. The tests *must not share* any other resources, e.g. objects or database
   schemas, that need to be updated during the test execution.

If this conditions hold, the general pattern provided above can be extened to
support parallel test execution.

```go
func TestUnitCall(t *testing.T) {
    t.Parallel()
    for message, param := range testUnitCallParams {
        message, param := message, param
        t.Run(message, test.Success(func(t *testing.T) {
            //Given
            unit, mocks := SetupTestUnit(t, param.mockSetup)

            //When
            result, err := unit.UnitCall(...)

            mocks.Wait()

            //Then
            if param.expectError != nil {
                assert.Equal(t, param.expectError, err)
            } else {
                require.NoError(t, err)
            }
            assert.Equal(t, param.expect*, result)
        }, true))
    }
}
```

**Note:** In the above pattern the setup for parallel tests hidden in the setup
of the isolated [test](../test) environment.

[gock]: https://github.com/h2non/gock "Gock"
[monkey]: https://github.com/bouk/monkey "Monkey Patching"
