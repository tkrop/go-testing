# Package testing/reflect

Setting up tests and comparing test results is most efficient, when you can
directly set up and compare the actual test and result objects. However, this
is sometimes prevented by objects not being open for construction and having
private states.

To cope with this challenge the `test`-package supports helpers to access, i.e.
read and write, private fields of objects using reflection.

* `test.NewBuilder[...]()` allows constructing a new object from scratch.
* `test.NewGetter(...)` allows reading private fields of an object by name.
* `test.NewSetter(...)` allows writing private fields by name, and finally
* `test.NewAccessor(...)` allows reading and writing of private fields by name.


## Example usage

The following example shows how the private properties of a closed error can be
set up using the `test.NewBuilder[...]()`.

```go
    err := test.NewBuilder[viper.ConfigFileNotFoundError]().
        Set("locations", fmt.Sprintf("%s", "...path...")).
        Set("name", "test").Build()
```

Similar we can set up input objects with private properties to minimize the
dependencies in the test setup, however, using this features exposes the test
to internal changes.
