# Testing framework

[![Build][build-badge]][build-link]
[![Coverage][coveralls-badge]][coveralls-link]
[![Coverage][coverage-badge]][coverage-link]
[![Quality][quality-badge]][quality-link]
[![Report][report-badge]][report-link]
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

[license-badge]: https://img.shields.io/badge/License-MIT-green.svg
[license-link]: https://opensource.org/licenses/MIT

[docs-badge]: https://pkg.go.dev/badge/github.com/tkrop/go-testing.svg
[docs-link]: https://pkg.go.dev/github.com/tkrop/go-testing

<!--
[libs-badge]: https://img.shields.io/librariesio/release/github/tkrop/go-testing
[libs-link]: https://libraries.io/github/tkrop/go-testing

[security-badge]: https://snyk.io/test/github/tkrop/go-testing/main/badge.svg
[security-link]: https://snyk.io/test/github/tkrop/go-testing
-->

Go testing extension, that allows a simple setup of strongly isolated unit,
component, and integration test providing advanced mock support extending
[gomock][gomock] and [gock][gock].


## Introduction

Are you tired of endless boiler plate code when writing high quality [`go`][go]
tests with exhaustive mock setups for isolating the systems under test?

Then [`go-testing`][go-testing] may be your library of choice. It provides
unified building blocks for writing short and effective unit component, and
integration tests in [`go`][go] using simple, common patterns, that allow to
target a [sensible, high-quality code coverage][unit-testing] using different
mock frameworks.

To accomplish this, [`go-testing`][go-testing] provides highly sophisticated
extensions for [`go`][go]'s [`testing`][testing] package as well as for mock
packages, e.g. [`gomock`][gomock] and [`gock`][gock], that lift limitation and
foster a simple, common setup of *strongly isolated* and *parallel running*
tests - supporting diverse success and failure scenarios, even in the presence
of spawned [`go`-routines][go-routines], or if the system under test panics.

While the [`test`](test) package provides the building blocks for efficient
test setup and test isolation, the [`mock`](mock) and [`gock`](gock) packages
provide access to a short pragmatic domain language for defining detailed mock
requests and responses that allow to enforce validation. Finally, the
[`reflect`](reflect) package provides access to private properties of the
system under test.

**Now also providing support for micro-benchmarks!**

You can find more information in the [`go-testing` documentation][go-testing].

[go]: <https://go.dev/>
[go-routines]: <https://go.dev/tour/concurrency>
[go-testing]: <https://pkg.go.dev/github.com/tkrop/go-testing>
[unit-testing]: <https://ricomariani.medium.com/100-unit-testing-now-its-ante-f0e2384ffedf>


### Example Usage

First you have to define a unified test/benchmark parameter set. While this can
be done in many different ways, the following setup structure is considered to
be the [`go-testing`][go-testing] idiomatic way due to its wide coverage of
different use cases, its flexibility, and its non-the-last readability:

```go
type UnitParams struct {
    setup        mock.SetupFunc
    input*...    *model.*
    expect       test.Expect
    expect*...   *model.*
    expectError  error
}

var unitTestCases = map[string]UnitParams {
    "success" {
        setup: mock.Chain(
            CallMockA(input..., output...),
            ...
            test.Panic("failure message"),
       ),
        ...
        expect: test.ExpectSuccess
    }
}
```

Now you can set up a *strongly isolated* and *parallel running* test. While
there are again many ways to define tests (see package [test](test)), the
following pattern is considered to be the most [`go-testing`][go-testing]
idiomatic way again due to its wide coverage of different use cases, its
flexibility, and its non-the-last readability:

```go
func TestUnit(t *testing.T) {
    // Setup the test using a map fo parameterization.
    test.Map(t, unitTestCases).
        // Filter set of test cases temporary or permanent.
        Filter(test.Not(test.Pattern[T]("^test-case-prefix"))).
        // Focus on set of test cases temporary or permanent.
        Filter(test.Pattern[T]("^test-case-name$")).
        // Run the test in parallel.
        Run(func(t test.Test, param UnitParams){

            // Given
            mocks := mock.NewMock(t).
                SetArg("common-arg", local.input*)...
                Expect(param.setup)

            unit := NewUnitService(
                mock.Get(mocks, NewServiceMock),
                ...
            )

            // When
            result, err := unit.call(param.input*...)

            mocks.Wait()

            // Then
            assert.Equal(t, param.expectError, err)
            assert.Equal(t, param.expect*, result)
        })
}
```

As an addon, you can also use the same pattern to define benchmarks for a
system under test based on the before defined test parameter set. The following
setup structure is considered to be the most [`go-testing`][go-testing]
framework idiomatic way (see also [Test benchmark
setup](test#parameterized-benchmark-setup)):

```go
func BenchmarkUnit(b *testing.B) {
    test.Map(test.Benchmark(b), unitTestCases).
        // Filter set of test cases temporary or permanent.
        Filter(test.Not(test.Pattern[T]("^test-case-prefix"))).
        // Focus on set of test cases temporary or permanent.
        Filter(test.Pattern[T]("^test-case-name$")).
        // Execute benchmark setup and loop phases.
        Benchmark(func(b *testing.B, param UnitParams) func(b *testing.B) {
            // Setup
            unit := NewUnitService(param.input*...)

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

For more test patterns and variations have a closer look at details in the
[test](test) package or read the [package docs][docs-test].

[docs-test]: <https://pkg.go.dev/github.com/tkrop/go-testing/test>


### Why parameterized test?

Parameterized (table-driven) test are an effective way to set up a systematic
set of test cases covering a system under test in a black or white box mode.
With the right tools and concepts — such as supported by this test framework —,
parameterized test allow to cover all success and failure paths of a system
under test.


### Why parallel tests?

Running tests in parallel makes the feedback loop on failures faster and helps
to detect failures from concurrent access. By using `go test -race` we can
easily uncover race conditions, that else only appear randomly in production,
and foster a design with clear responsibilities. This side-effects compensate
for the small additional effort needed to write parallel tests.


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

The [`go-testing`][go-testing] framework consists of the following
sub-packages:

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
race conditions that else randomly appear in production, by running tests using
`go test -race`.

**Note:** there are some general requirements for running test in parallel:

1. Tests *must not modify* environment variables dynamically — utilize test
   friendly configuration concepts instead.
2. Tests *must not require* reserved service ports and open listeners — setup
   services to acquire dynamic ports instead.
3. Tests *must not share* any files, folders, and pipelines, e.g. `stdin`,
  `stdout`, or `stderr` — implement logic by using wrappers that can be easily
   redirected and mocked.
4. Tests *must not share* database schemas or tables, that are updated during
   execution of parallel tests — implement test to set up test specific
   database schemas.
5. Tests *must not share* process resources, that are update during execution
   of parallel tests. Many frameworks make use of common global resources that
   make them unsuitable for parallel tests — use frameworks that do not suffer
   by these flaws.

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


[testing]: <https://pkg.go.dev/testing>
[gomock]: <https://go.uber.org/mock>
[gock]: <https://github.com/h2non/gock>
[monkey]: <https://github.com/bouk/monkey>
[gin]: <https://github.com/gin-gonic/gin>
[chi]: <https://github.com/go-chi/chi>


## Building

This project is using a custom build system called [go-make][go-make], that
provides default targets for most common tasks. Makefile rules are generated
based on the project structure and files for common tasks, to initialize,
build, test, and run the components in this repository.

To get started, run one of the following commands.

```bash
make help
make show-targets
```

Read the [go-make manual][go-make-man] for more information about targets
and configuration options.

**Not:** [go-make][go-make] installs `pre-commit` and `commit-msg`
[hooks][git-hooks] calling `make commit` to enforce successful testing and
linting and `make git-verify message` to validate whether the commit message
is following the [conventional commit][convent-commit] best practice.

[go-make]: <https://github.com/tkrop/go-make>
[go-make-man]: <https://github.com/tkrop/go-make/blob/main/MANUAL.md>
[git-hooks]: <https://git-scm.com/book/en/v2/Customizing-Git-Git-Hooks>
[convent-commit]: <https://www.conventionalcommits.org/en/v1.0.0/>


## Terms of Usage

This software is open source under the MIT license. You can use, fork, and copy
it without restrictions and liabilities. Please give the project a star, when
you consider it worthy.


## Contributing

If you like to contribute, please create an issue and/or pull request with a
proper description of your proposal or contribution. I will review it and
provide feedback on it as fast as possible.


## Disclaimer

This software is developed with the help of AI following the highest human
standards. All actions executed by AI are carefully reviewed, counter-checked,
and corrected with the highest human standards and quality goals in mind. No
AI generate code is allowed to be merged or released without a careful human
reviews to prevent systematic degeneration of coding standards and code
quality.
