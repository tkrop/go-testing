# Internal utils

The `internal` packages contain all the helpful generic solutions developed to
support the functionality of the [`go-testing`][go-testing] framework, but are
not supposed to be exported and used by others.

The `internal` packages consist of the following sub-packages:

* [`iter`](iter) provides iterator wrappers for structures in standard packages
  that do not provide these yet.

* [`maps`](maps) provides helpful functions for maps that are not supported by
  the standard map packages yet.

* [`reflect`](reflect) contains a collection of helpful generic functions that
  support reflection. The functions are used by the [`mock`](../mock) and the
  [`test`](../test) packages to implement major features.

* [`slices`](slices) contains a collection of helpful generic functions for
  working with slices. The functions are mainly used by the [`perm`](../perm)
  and the [`test`](../test) package to implement minor features.

* [`sync`](sync) provides a lenient wait group implementation for coordinating
  [`mock`](../mock)s in the isolated [`test`](../test)s to gracefully unlock all
  waiters after test failures to finish the test.

[go-testing]: <https://pkg.go.dev/github.com/tkrop/go-testing>
