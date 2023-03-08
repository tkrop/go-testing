<!-- markdownlint-disable -->
[![Build](https://github.com/tkrop/go-testing/actions/workflows/go.yaml/badge.svg)](https://github.com/tkrop/go-testing/actions/workflows/go.yaml)
[![Coverage](https://coveralls.io/repos/github/tkrop/go-testing/badge.svg?branch=main)](https://coveralls.io/github/tkrop/go-testing?branch=main)
[![Coverage](https://app.codacy.com/project/badge/Coverage/cc1c47ec5ce0493caf15c08fa72fc78c)](https://www.codacy.com/gh/tkrop/go-testing/dashboard?utm_source=github.com&utm_medium=referral&utm_content=tkrop/go-testing&utm_campaign=Badge_Coverage)
[![Quality](https://app.codacy.com/project/badge/Grade/cc1c47ec5ce0493caf15c08fa72fc78c)](https://www.codacy.com/gh/tkrop/go-testing/dashboard?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=tkrop/go-testing&amp;utm_campaign=Badge_Grade)
[![Report](https://goreportcard.com/badge/github.com/tkrop/go-testing)](https://goreportcard.com/report/github.com/tkrop/go-testing)
[![FOSSA](https://app.fossa.com/api/projects/git%2Bgithub.com%2Ftkrop%2Ftesting.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Ftkrop%2Ftesting?ref=badge_shield)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Docs](https://pkg.go.dev/badge/github.com/tkrop/go-testing.svg)](https://pkg.go.dev/github.com/tkrop/go-testing)
<!--
[![Libraries](https://img.shields.io/librariesio/release/github/tkrop/go-testing)](https://libraries.io/github/tkrop/go-testing)
[![Security](https://snyk.io/test/github/tkrop/go-testing/main/badge.svg)](https://snyk.io/test/github/tkrop/go-testing)
-->
<!-- markdownlint-enable -->


# Testing framework

Goal of the `testing` framework is to provide simple and efficient tools to for
writing effective unit and component tests in [Go][go].

To accomplish this, the `testing` framework contains a couple of opinionated
small extensions for [testing][testing], [Gomock][gomock], and [Gock][gock] to
enable isolated, parallel, parameterized tests using a common pattern to setup
strongly validating mock request and response chains that work across detached
`go` routines and various error scenarios.


## Example Usage

The core idea of the [mock](mock)/[gock](gock) packages is to provide a short
pragmatic domain language for defining mock requests with response that enforce
validation, while the [test](test) package provides the building blocks for
test isolation.

```go
type UnitParams struct {
    mockSetup    mock.SetupFunc
    input*...    *model.*
    expect       test.Expect
    expect*...   *model.*
    expectError  error
}

var testUnitParams = map[string]UnitParams {
    "success" {
        mockSetup: mock.Chain(
            CallMockA(input..., output...),
            ...
            test.Panic("failure message"),
       ),
        ...
        expect: test.ExpectSuccess
    }
}

func TestUnit(t *testing.T) {
    test.Map(t, testParams).
        Run(func(t test.Test, param UnitParams){

        // Given
        mocks := mock.NewMock(t).Expect(
          param.mockSetup,
        )

        unit := NewUnitService(
            mock.Get(mocks, NewServiceMock),
            ...
        )

        // When
        result, err := unit.call(param.input*...)

        mocks.Wait()

        // Then
        if param.expectError != nil {
            assert.Equal(t, param.expectError, err)
        } else {
            require.NoError(t, err)
        }
        assert.Equal(t, param.expect*, result)
    })
}
```

This opinionated test pattern supports a wide range of test in a standardized
way.


## Why parameterized test?

Parameterized test are an efficient way to setup a high number of related test
cases cover the system under test in a black box mode from feature perspective.
With the right tools and concepts - as provided by this `testing` framework,
parameterized test allow to cover all success and failure paths of a system
under test as outlined above.


## Why parallel tests?

Running tests in parallel make the feedback loop on failures faster, help to
detect failures from concurrent access and race conditions using `go test
-race`, that else only appear randomly in production, and foster a design with
clear responsibilities. This side-effects compensate for the small additional
effort needed to write parallel tests.


## Why isolation of tests?

Test isolation is a precondition to have stable running test - especially run
in parallel. Isolation must happen from input perspective, i.e. the outcome of
a test must not be affected by any previous running test, but also from output
perspective, i.e. it must not affect any later running test. This is often
complicated since many tools, patterns, and practices break the test isolation
(see [requirements for parallel isolated
tests](#requirements-for-parallel-isolated-tests).


## Why strong validation?

Test are only meaningful, if they validate ensure pre-conditions and validate
post-conditions sufficiently strict. Without validation test cannot ensure that
the system under test behaves as expected - even with 100% code and branch
coverage. As a consequence, a system may fail in unexpected ways in production.

Thus it is advised to validate mock input parameters for mocked requests and
to carefully define the order of mock requests and responses. The [mock](mock)
framework makes this approach as simple as possible, but it is still the
responsibility of the developer to setup the validation correctly.


# Framework structure

The `testing` framework consists of the following sub-packages:

* [test](test) provides a small framework to simply isolate the test execution
  and safely check whether a test fails or succeeds as expected in coordination
  with the [mock](mock) package - even in if a system under test spans detached
  go-routines.

* [mock](mock) provides the means to setup a simple chain or a complex network
  of expected mock calls with minimal effort. This makes it easy to extend the
  usual narrow range of mocking to larger components using a unified pattern.

* [gock](gock) provides a drop-in extension for [Gock][gock] consisting of a
  controller and a mock storage that allows to run tests isolated. This allows
  to parallelize simple test and parameterized tests.

* [perm](perm) provides a small framework to simplify permutation tests, i.e.
  a consistent test set where conditions can be checked in all known orders
  with different outcome. This is very handy in combination with [test](test)
  to validated the [mock](mock) framework, but may be useful in other cases
  too.

Please see the documentation of the sub-packages for more details.


# Requirements for parallel isolated tests

Running tests in parallel not only makes test faster, but also helps to detect
race conditions that else randomly appear in production  when running tests
with `go test -race`.

**Note:** there are some general requirements for running test in parallel:

1. Tests *must not modify* environment variables dynamically - utilize test
   specific configuration instead.
2. Tests *must not require* reserved service ports and open listeners - setup
   services to acquire dynamic ports instead.
3. Tests *must not share* files, folder and pipelines, e.g. `stdin`, `stdout`,
   or `stderr` - implement logic by using wrappers that can be redirected and
   mocked.
4. Tests *must not share* database schemas or tables, that are updated during
   execution of parallel tests - implement test to setup test specific database
   schemas.
5. Tests *must not share* process resources, that are update during execution
   of parallel tests. Many frameworks make use of common global resources that
   make them unsuitable for parallel tests.

Examples for such shared resources in common frameworks are:

* Using of [monkey patching][monkey] to modify commonly used global functions,
  e.g. `time.Now()` - implement access to these global functions using lambdas
  and interfaces to allow for mocking.
* Using of [Gock][gock] to mock HTTP responses on transport level - make use
  of the [gock](gock)-controller provided by this framework.
* Using the [Gin][gin] HTTP web framework which uses a common `json`-parser
  setup instead of a service specific configuration. While this is not a huge
  deal, the repeated global setup creates race alerts. Instead use [chi][chi]
  that supports a service specific configuration.

With a careful design the general pattern provided above can be used to support
parallel test execution.


# Terms of Usage

This software is open source as is under the MIT license. If you start using
the software, please give it a star, so that I know to be more careful with
changes. If this project has more than 25 Stars, I will introduce semantic
versions for changes.


# Contributing

If you like to contribute, please create an issue and/or pull request with a
proper description of your proposal or contribution. I will review it and
provide feedback on it.


[go]: https://go.dev/ "Golang"
[testing]: https://pkg.go.dev/testing "Go Testing"
[gomock]: https://github.com/golang/mock "GoMock"
[gock]: https://github.com/h2non/gock "Gock"
[monkey]: https://github.com/bouk/monkey "Monkey Patching"
[gin]: https://github.com/gin-gonic/gin "Gin HTTP web framework"
[chi]: https://github.com/go-chi/chi "Chi HTTP web service"
