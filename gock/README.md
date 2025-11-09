# Package testing/gock

Goal of this package is to provide a small controller to isolate testing of
services (gateways) by mocking the network communication using [`gock`][gock].

**Note:** Since the controller is focused on testing, it does not support the
same networking and observation features of [`gock`][gock] and requires manual
transport interception setup. However, the interface is mainly compatible.


## Example usage

Just create a new controller on each test, connect it to your HTTP clients or
client wrappers, and then use it to create HTTP mock request/response cycles
as usual.


```go
func TestUnit(t *testing.T) {
    // Given
    gock := gock.NewController(t)

    client := &http.Client{}
    gock.InterceptClient(client)

    gock.New("http://foo.com").Get("/bar").
        {Reply(status)|ReplyError(err)}.BodyString("result")

    // When
    ...
}
```

The controller is fully integrated in to the [`mock`](../mock)-framework, so
that you can just request the controller via the [`gomock`][gomock] constructor
`mock.Get(mocks, gock.NewGock)` (see
[Example](#integration-with-mock-framework))

**Note:** The standard cardinality of mock requests using [`gock`][gock] is
`1`. So you can skip writing `Times(1)` and only use `Times(n)`, when you need
to increase the request cardinality.


## Migration from Gock

Migration from [`gock`][gock] to this package is straight forward. You just add
the controller creation at the begin of your test giving it the name `gock` and
hand it over to all methods creating HTTP request/response mocks. The mock
creation than happens as usual.

```go
func TestUnit(t *testing.T) {
    // Given
    gock := gock.NewController(t)

    ...

    gock.New("http://foo.com").Get("/bar").
        {Reply(status)|ReplyError(err)}.BodyString("result")

    // WHen
    ...
}
```

Since the controller does not intercept all transports by default, you need to
setup transport interception manually. This can happen in three different ways.
If you have access to the HTTP request/response client, you can use the usual
`InterceptClient` (and `RestoreClient`) methods.

```go
func TestUnit(t *testing.T) {
    // Given
    gock := gock.Controller(t)

    ...

    client := &http.Client{}
    gock.InterceptClient(client)
    defer gock.RestoreClient(client) // optional

    // When
    ...
}
```

Customized HTTP clients e.g. [`resty`][resty] may not give direct access to the
transport but offer the ability to set a `http.RoundTripper`. The controller
implements this interface and therefore can be simply used as drop in entity.

```go
func TestUnit(t *testing.T) {
    // Given
    gock := gock.Controller(t)

    ...

    client := resty.New()
    client.setTransport(gock)

    // When
    ...
}
```

As a last resort, you can also intercept the `http.DefaultTransport`, however,
this is not advised, since it will destroy the test isolation that is goal of
this controller framework. In this case you should use [`gock`][gock] directly.


## Integration with `mock`-framework

The `Gock`-controller framework supports a simple integration with the
[`mock`](../mock) framework for [gomock][gomock]: it simply provides
constructor that accepts the `gomock`-controller. Using constructor, it is
possible to create the usual setup methods similar as described in the
[generic mock service call pattern](../mock#generic-mock-service-call-pattern).

```go
func GockCall(
    url, path string, input..., status int, output..., error,
) mock.SetupFunc {
    return func(mocks *Mocks) any {
        mock.Get(mocks, gock.NewGock).New(url).Get(path).
            {Reply(status)|ReplyError(err)}...
        return nil
    }
}
```

**Note:** While this already nicely integrates the mock controller creation,
call setup, and validation, it currently provides no support for call order
validation as [`gomock`][gomock] supports it.


[gomock]: <https://go.uber.org/mock>
[gock]: <https://github.com/h2non/gock>
[resty]: <https://github.com/go-resty/resty>
