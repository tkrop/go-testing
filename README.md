# Testing

The testing projects contains a couple of small opinionated extensions for
[Golang](https://go.dev/) and [Gomock](https://github.com/golang/mock) to
simplify and enable complicated unit and component tests.

* [mock](mock) provides the means to setup a simple chain or a complex network
  of expected mock calls with minimal effort. This makes it easy to extend the
  usual narrow range of mocking to larger components using a unified pattern.

* [test](test) provides a small framework to simply isolate the test execution
  and safely check whether a test fails as expected. This is primarily very
  handy to validate a test framework as provided by the [mock](mock) package
  but may be handy in other cases too.

* [perm](perm) provides a small framework to simplify permutation tests, i.e.
  a consistent test set where conditions can be checked in all known orders
  with different outcome. This is very handy in combination with [test](test)
  to validated the [mock](mock) framework, but may be useful in other cases
  too.

 Please see the documentation of the sub-packages for more details.

# Terms of Usage

This software is open source as is under the MIT license. If you start using
the software, please give it a star, so that I know to be more careful with
changes. If this project has more than 25 Stars, I will introduce semantic
versioning for changes.

# Contributing

If you like to contribute, please create an issue and/or pull request with a
proper description of your proposal or contribution. I will review it and
provide feedback on it.