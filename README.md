# Testing framework

[![Build][build-badge]][build-link]
[![Coverage][coveralls-badge]][coveralls-link]
[![Coverage][coverage-badge]][coverage-link]
[![Quality][quality-badge]][quality-link]
[![Report][report-badge]][report-link]
[![FOSSA][fossa-badge]][fossa-link]
[![License][license-badge]][license-link]
[![Docs][docs-badge]][docs-link]
<!--
[![Libraries][libs-badge]][libs-link]
[![Security][security-badge]][security-link]
-->

[build-badge]: https://github.com/tkrop/go-testing/actions/workflows/go.yaml/badge.svg
[build-link]: https://github.com/tkrop/go-testing/actions/workflows/go.yaml

[coveralls-badge]: https://coveralls.io/repos/github/tkrop/go-testing/badge.svg?branch=main
[coveralls-link]: https://coveralls.io/github/tkrop/go-testing?branch=main

[coverage-badge]: https://app.codacy.com/project/badge/Coverage/cc1c47ec5ce0493caf15c08fa72fc78c
[coverage-link]: https://app.codacy.com/gh/tkrop/go-testing/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_coverage

[quality-badge]: https://app.codacy.com/project/badge/Grade/cc1c47ec5ce0493caf15c08fa72fc78c
[quality-link]: https://app.codacy.com/gh/tkrop/go-testing/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_grade

[report-badge]: https://goreportcard.com/badge/github.com/tkrop/go-testing
[report-link]: https://goreportcard.com/report/github.com/tkrop/go-testing

[fossa-badge]: https://app.fossa.com/api/projects/git%2Bgithub.com%2Ftkrop%2Ftesting.svg?type=shield&issueType=license
[fossa-link]: https://app.fossa.com/projects/git%2Bgithub.com%2Ftkrop%2Ftesting?ref=badge_shield&issueType=license

[license-badge]: https://img.shields.io/badge/License-MIT-yellow.svg
[license-link]: https://opensource.org/licenses/MIT

[docs-badge]: https://pkg.go.dev/badge/github.com/tkrop/go-testing.svg
[docs-link]: https://pkg.go.dev/github.com/tkrop/go-testing

<!--
[libs-badge]: https://img.shields.io/librariesio/release/github/tkrop/go-testing
[libs-link]: https://libraries.io/github/tkrop/go-testing

[security-badge]: https://snyk.io/test/github/tkrop/go-testing/main/badge.svg
[security-link]: https://snyk.io/test/github/tkrop/go-testing
-->

## Introduction

Goal of the `testing` framework is to provide simple and efficient tools to for
writing effective unit, component, and integration tests in [`go`][go].

To accomplish this, the `testing` framework provides a couple of extensions for
to standard [`testing`][testing] package of [`go`][go] that support a simple
setup of [`gomock`][gomock] and [`gock`][gock] in isolated, parallel, and
parameterized tests using a common pattern to setup with strong validation of
mock request and response that work under various failure scenarios and even in
the presense of [`go`-routines][go-routines].

[go-routines]: <https://go.dev/tour/concurrency>


### Example Usage

The core idea of the [`mock`](mock)/[`gock`](gock) packages is to provide a
short pragmatic domain language for defining mock requests with responses that
enforce validation, while the [`test`](test) package provides the building
blocks for test isolation.

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
        mocks := mock.NewMock(t).
            SetArg("common-arg", local.input*)...
            Expect(param.mockSetup)

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
way. For variations have a closer look at the [test](test) package.


### Why parameterized test?

Parameterized test are an efficient way to setup a high number of related test
cases cover the system under test in a black box mode from feature perspective.
With the right tools and concepts - as provided by this `testing` framework,
parameterized test allow to cover all success and failure paths of a system
under test as outlined above.


### Why parallel tests?

Running tests in parallel make the feedback loop on failures faster, help to
detect failures from concurrent access and race conditions using `go test
-race`, that else only appear randomly in production, and foster a design with
clear responsibilities. This side-effects compensate for the small additional
effort needed to write parallel tests.


### Why isolation of tests?

Test isolation is a precondition to have stable running test - especially run
in parallel. Isolation must happen from input perspective, i.e. the outcome of
a test must not be affected by any previous running test, but also from output
perspective, i.e. it must not affect any later running test. This is often
complicated since many tools, patterns, and practices break the test isolation
(see [requirements for parallel isolated
tests](#requirements-for-parallel-isolated-tests).


### Why strong validation?

Test are only meaningful, if they validate ensure pre-conditions and validate
post-conditions sufficiently strict. Without validation test cannot ensure that
the system under test behaves as expected - even with 100% code and branch
coverage. As a consequence, a system may fail in unexpected ways in production.

Thus it is advised to validate mock input parameters for mocked requests and
to carefully define the order of mock requests and responses. The
[`mock`](mock) framework makes this approach as simple as possible, but it is
still the responsibility of the developer to setup the validation correctly.


## Framework structure

The `testing` framework consists of the following sub-packages:

* [`test`](test) provides a small framework to simply isolate the test execution
  and safely check whether a test fails or succeeds as expected in coordination
  with the [`mock`](mock) package - even in if a system under test spans
  detached [`go`-routines][go-routines].

* [`mock`](mock) provides the means to setup a simple chain or a complex network
  of expected mock calls with minimal effort. This makes it easy to extend the
  usual narrow range of mocking to larger components using a unified pattern.

* [`gock`](gock) provides a drop-in extension for [Gock][gock] consisting of a
  controller and a mock storage that allows to run tests isolated. This allows
  to parallelize simple test and parameterized tests.

* [`perm`](perm) provides a small framework to simplify permutation tests, i.e.
  a consistent test set where conditions can be checked in all known orders
  with different outcome. This is very handy in combination with [`test`](test)
  to validated the [`mock`](mock) framework, but may be useful in other cases
  too.

Please see the documentation of the sub-packages for more details.


## Requirements for parallel isolated tests

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
* Using of [`gock`][gock] to mock HTTP responses on transport level - make use
  of the [`gock`](gock)-controller provided by this framework.
* Using the [Gin][gin] HTTP web framework which uses a common `json`-parser
  setup instead of a service specific configuration. While this is not a huge
  deal, the repeated global setup creates race alerts. Instead use [`chi`][chi]
  that supports a service specific configuration.

With a careful design the general pattern provided above can be used to support
parallel test execution.


## Terms of Usage

This software is open source as is under the MIT license. If you start using
the software, please give it a star, so that I know to be more careful with
changes. If this project has more than 25 Stars, I will introduce semantic
versions for changes.


## Building

This project is using [go-make][go-make] for building, which provides default
implementations for most common tasks. Read the [go-make manual][go-make-man]
for more information about how to build, test, lint, etc.

[go-make]: <https://github.com/tkrop/go-make>
[go-make-man]: <https://github.com/tkrop/go-make/blob/main/MANUAL.md>


## Contributing

If you like to contribute, please create an issue and/or pull request with a
proper description of your proposal or contribution. I will review it and
provide feedback on it.


[go]: <https://go.dev/>
[testing]: <https://pkg.go.dev/testing>
[gomock]: <https://github.com/golang/mock>
[gock]: <https://github.com/h2non/gock>
[monkey]: <https://github.com/bouk/monkey>
[gin]: <https://github.com/gin-gonic/gin>
[chi]: <https://github.com/go-chi/chi>
