# Package testing/reflect

Setting up tests and comparing test results is most efficient, when you can
directly set up and compare the actual test and result objects. However, this
is sometimes prevented by objects not being open for construction and having
private states.

To cope with this challenge the [`reflect`][reflect] package supports helpers
to access, i.e. read and write, private fields of objects using reflection.

* [`reflect.NewBuilder[T]()`][builder] — provides a builder that allows to
  construct a new struct from scratch using methods to write to fields by
  field name.
* [`reflect.NewGetter(...)`][getter] — provides a getter that allows to read the
  private fields of an struct by field name.
* [`reflect.NewSetter(...)`][setter] — provides a setter that allows to write
  the private fields of an struct by field name.
* [`reflect.NewFinder(...)`][finder] — provides a finder that allows to find a
  private field by type and name.
* [`reflect.NewAccessor(...)`][accessor] — provides an accessor that allows to
  read and to write private fields of an struct by names name.


[reflect]: <https://pkg.go.dev/github.com/tkrop/go-testing/reflect>
[builder]: <https://pkg.go.dev/github.com/tkrop/go-testing/reflect#NewBuilder>
[getter]: <https://pkg.go.dev/github.com/tkrop/go-testing/reflect#NewGetter>
[setter]: <https://pkg.go.dev/github.com/tkrop/go-testing/reflect#NewSetter>
[accessor]: <https://pkg.go.dev/github.com/tkrop/go-testing/reflect#NewAccessor>
[finder]: <>


## Example usage

The following example shows how the private properties of a closed error can be
set up using the `test.NewBuilder[...]()`.

```go
    err := test.NewBuilder[viper.ConfigFileNotFoundError]().
        Set("locations", fmt.Sprintf("%s", "...path...")).
        Set("name", "test").Build()
```

Similar you can set up input objects with private properties to minimize the
dependencies in the test setup, however, using this features exposes the test
to internal changes of the tested classes and makes tests brittle.
