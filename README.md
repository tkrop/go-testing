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

[build-badge]: https://github.com/tkrop/go-testing/actions/workflows/build.yaml/badge.svg
[build-link]: https://github.com/tkrop/go-testing/actions/workflows/build.yaml

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
setup of test cases using [`gomock`][gomock] and [`gock`][gock] in isolated,
parallel, and parameterized tests using a common pattern with strong validation
of mock request and response that work under various failure scenarios and even
in the presence of spawned [`go`-routines][go-routines].

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
        Timeout(50 * time.Millisecond)
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

Parameterized test are an effective way to set up a systematic set of test
cases covering a system under test in a black box mode. With the right tools
and concepts — as provided by this `testing` framework, parameterized test
allow to cover all success and failure paths of a system under test.


### Why parallel tests?

Running tests in parallel make the feedback loop on failures faster, help to
detect failures from concurrent access and race conditions using `go test
-race`, that else only appear randomly in production, and foster a design with
clear responsibilities. This side-effects compensate for the small additional
effort needed to write parallel tests.


### Why isolation of tests?

Test isolation is a precondition to have stable running test — especially run
in parallel. Isolation must happen from input perspective, i.e. the outcome of
a test must not be affected by any previous running test, but also from output
perspective, i.e. it must not affect any later running test. This is often
complicated since many tools, patterns, and practices break the test isolation
(see [requirements for parallel isolated
tests](#requirements-for-parallel-isolated-tests).


### Why strong validation?

Test are only meaningful, if they ensure/validate pre-conditions as well as
validate/ensure post-conditions sufficiently strict. Without validation test
cannot ensure that the system under test behaves as expected — even with 100%
code and branch coverage. As a consequence, a system may fail in unexpected
ways in production.

Thus, it is advised to validate input parameters for mocked requests and to
carefully define the order of mock requests and responses. The [`mock`](mock)
framework makes this approach as simple as possible, but it is still the
responsibility of the test developer to set up the validation correctly.


## Framework structure

The `testing` framework consists of the following sub-packages:

* [`test`](test) provides a small framework to isolate the test execution and
  safely check whether a test fails or succeeds as expected in combination with
  the [`mock`](mock) package — even if a system under test spans detached
  [`go`-routines][go-routines].

* [`mock`](mock) provides the means to set up a simple chain as well as a
  complex network of expected mock calls with minimal effort. This makes it
  easy to extend the usual narrow range of mocking to larger components using
  a unified test pattern.

* [`gock`](gock) provides a drop-in extension for the [Gock][gock] package
  consisting of a controller and a mock storage that allows running tests
  isolated. This allows parallelizing simple test as well as parameterized
  tests.

* [`perm`](perm) provides a small framework to simplify permutation tests, i.e.
  a consistent test set where conditions can be checked in all known orders
  with different outcome. This was very handy in combination with [`test`](test)
  for validating the [`mock`](mock) framework, but may be useful in other cases
  too.

Please see the documentation of the sub-packages for more details.


## Requirements for parallel isolated tests

Running tests in parallel makes test not only faster, but also helps to detect
race conditions that else randomly appear in production, when running tests
with `go test -race`.

**Note:** there are some general requirements for running test in parallel:

1. Tests *must not modify* environment variables dynamically — utilize test
   specific configuration instead.
2. Tests *must not require* reserved service ports and open listeners — setup
   services to acquire dynamic ports instead.
3. Tests *must not share* files, folder and pipelines, e.g. `stdin`, `stdout`,
   or `stderr` — implement logic by using wrappers that can be redirected and
   mocked.
4. Tests *must not share* database schemas or tables, that are updated during
   execution of parallel tests — implement test to set up test specific database
   schemas.
5. Tests *must not share* process resources, that are update during execution
   of parallel tests. Many frameworks make use of common global resources that
   make them unsuitable for parallel tests.

Examples for such shared resources in common frameworks are:

* Using of [monkey patching][monkey] to modify commonly used global functions,
  e.g. `time.Now()` — implement access to these global functions using lambdas
  and interfaces to allow for mocking.
* Using of [`gock`][gock] to mock HTTP responses on transport level — make use
  of the [`gock`](gock)-controller provided by this framework.
* Using the [Gin][gin] HTTP web framework which uses a common `json`-parser
  setup instead of a service specific configuration. While this is not a huge
  deal, the repeated global setup creates race alerts. Instead, use
  [`chi`][chi] that supports a service specific configuration.

With a careful system design, the general pattern provided above can be used
to create parallel test for a wide range of situations.


## Building

This project is using [go-make][go-make], which provides default targets for
most common tasks, to initialize, build, test, and run the software of this
project. Read the [go-make manual][go-make-man] for more information about
targets and configuration options.

[go-make]: <https://github.com/tkrop/go-make>
[go-make-man]: <https://github.com/tkrop/go-make/blob/main/MANUAL.md>

The [`Makefile`](Makefile) depends on a preinstalled [`go`][go] for version
management, and makes heavy use of GNU tools, i.e. [`coretils`][core],
[`findutils`][find], ['(g)make'][make], [`(g)awk`][awk], [`(g)sed`][sed], and
not the least [`bash`][bash]. For certain non-core-features it also requires
[`docker`][docker]/[`podman`][podman] and [`curl`][curl]. On MacOS, it uses
[brew][brew] to ensure that the latest versions with the exception
[`docker`][docker]/[`podman`][podman] are.

[go]: <https://go.dev/>
[brew]: <https://brew.sh/>
[curl]: <https://curl.se/>
[docker]: <https://www.docker.com/>
[podman]: <https://podman.io/>
[make]: <https://www.gnu.org/software/make/>
[bash]: <https://www.gnu.org/software/bash/>
[core]: <https://www.gnu.org/software/coreutils/>
[find]: <https://www.gnu.org/software/findutils/>
[awk]: <https://www.gnu.org/software/awk/>
[sed]: <https://www.gnu.org/software/sed/>

**Not:** [go-make][go-make] automatically installs `pre-commit` and `commit-msg`
[hooks][git-hooks] overwriting and deleting pre-existing hooks (see also
[Customizing Git - Git Hooks][git-hooks]). The `pre-commit` hook calls
`make commit` as an alias for executing  `test-go`, `test-unit`, `lint-<level>`,
and `lint-markdown` to enforce successful testing and linting. The `commit-msg`
hook calls `make git-verify message` for validating whether the commit message
is following the [conventional commit][convent-commit] best practice.

[git-hooks]: <https://git-scm.com/book/en/v2/Customizing-Git-Git-Hooks>
[convent-commit]: <https://www.conventionalcommits.org/en/v1.0.0/>


## Terms of Usage

This software is open source under the MIT license. You can use it without
restrictions and liabilities. Please give it a star, so that I know. If the
project has more than 25 Stars, I will introduce semantic versions `v1`.


## Contributing

If you like to contribute, please create an issue and/or pull request with a
proper description of your proposal or contribution. I will review it and
provide feedback on it as fast as possible.


[testing]: <https://pkg.go.dev/testing>
[gomock]: <https://github.com/golang/mock>
[gock]: <https://github.com/h2non/gock>
[monkey]: <https://github.com/bouk/monkey>
[gin]: <https://github.com/gin-gonic/gin>
[chi]: <https://github.com/go-chi/chi>
