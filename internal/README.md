# Internal utils

The `internal` utils contain the all helpful generic solutions developed to
support the functionality of the `testing` framework, but are not supposed to
be exported and used by others.

The `internal` utils consist of the following sub-packages:

* [math](math) provides generic `Min`/`Max` functions that are used by the
  [reflect](reflect) package.

* [reflect](reflect) contains a collection of helpful generic functions that
  support reflection. The functions are used by the [mock](../mock) and the
  [test](../test) packages to implement major features.

* [slices](slices) contains a collection of helpful generic functions for
  working with slices. The functions are mainly used by the [perm](../perm)
  and the [test](../test) package to implement minor features.

* [sync](sync) provides a lenient wait group implementation for coordinating
  [mock](../mock)s in the isolated [test](../test)s to gracefully unlock all
  waiters after test failures to finish the test.
