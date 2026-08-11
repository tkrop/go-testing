# Package testing/gock

The [`gock`][gock] package provides a small controller to isolate testing of
services (gateways) by mocking the network communication using
[`gock`][gock-h2non].

**Note:** Since the controller is focused on testing, it does not support the
full networking and observation features of [`gock`][gock-h2non] and requires
manual transport interception setup, however, the interface is mainly
compatible.


## Example usage

The most convenient way of applying the `gock` framework is using it together
with the [`mock`](../mock) framework as follows:

```go
func SetupClientGock(
    t test.Test, setup mock.SetupFunc, err error,
) Client {
    t.Helper()

    mocks := mock.NewMocks(t).Expect(setup)

    client := github.NewClient()

    // Intercept the client transport with the gock controller.
    http := test.Cast[*http.Client](reflect.NewAccessor(client).Get("client"))
    if err == nil {
        http.Transport = mock.Get(mocks, gock.NewGock)
    } else {
        http.Transport = NewErrorRoundTripper(err)
    }

    return client
}
```

This now allows to register mock HTTP request/response cycles using a simple
helper function that is compatible with the [`mock`](../mock) framework as
follows:

```go
func GockCall(
    url, path string, input..., status int, output..., error,
) mock.SetupFunc {
    return func(mocks *Mocks) any {
        mock.Get(mocks, gock.NewGock).New(url).Get(path).Times(1).
            {Reply(status)|ReplyError(err)}...
        return nil
    }
}
```

**Note:** The `return nil` is required to satisfy the `mock.SetupFunc`
function, and to account for the shortcoming that the integration currently
does not support ordering `gock.Response`.


## Standalone usage

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
        [Reply(status)|ReplyError(err)].BodyString("result")

    // When
    ...
}
```

The controller is fully integrated in to the [`mock`](../mock)-framework, so
that you can just request the controller via the [`gomock`][gomock] constructor
`mock.Get(mocks, gock.NewGock)` (see
[Example](#integration-with-mock-framework))

**Note:** The standard cardinality of mock requests using [`gock`][gock-h2non]
is `1`. So you can skip writing `Times(1)` and only use `Times(n)`, when you
need to increase the request cardinality.


## Migration from Gock

Migration from [`gock`][gock-h2non] to this package is straight forward. You
just add the controller creation at the begin of your test giving it the name
`gock` and hand it over to all methods creating HTTP request/response mocks.
The mock creation than happens as usual.

```go
func TestUnit(t *testing.T) {
    // Given
    gock := gock.NewController(t)

    ...

    gock.New("http://foo.com").Get("/bar").
        {Reply(status)|ReplyError(err)}.BodyString("result")

    // When
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

The [`Controller`][gock-ctrl] also supports a simple integration with the
[`mock`](../mock) framework for [gomock][gomock]. It provides a constructor
([`gock.NewGock`][gock-new]) that is compatible with [`mock.Get`][mock-get].
The controller is automatically registered in the mock controller and can be
retrieved using the `mock.Get` method.

The following example shows how to set up a client redirecting the transport
layer to the `gock` controller setting up the `mock` framework:

```go
func SetupClientGock(
    t test.Test, setup mock.SetupFunc, err error,
) Client {
    t.Helper()

    mocks := mock.NewMocks(t).Expect(setup)

    client := github.NewClient()

    // Intercept the client transport with the gock controller.
    http := test.Cast[*http.Client](reflect.NewAccessor(client).Get("client"))
    if err == nil {
        http.Transport = mock.Get(mocks, gock.NewGock)
    } else {
        http.Transport = gock.NewErrorRoundTripper(err)
    }

    return client
}
```

Using this constructor, it is possible to create standard setup methods
similar as described in the [generic mock service call
pattern](../mock#generic-mock-service-call-pattern).

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
call setup, and call validation, it currently provides no support for call
order validation as [`gomock`][gomock] supports it. As a consequence, the
call functions must not return the [`gock.Response`][gock] for further use.


[gock]: <https://pkg.go.dev/github.com/tkrop/go-testing/gock>
[gock-new]: <https://pkg.go.dev/github.com/tkrop/go-testing/gock#NewGock>
[gock-ctrl]: <https://pkg.go.dev/github.com/tkrop/go-testing/gock#Controller>
[gock-h2non]: <https://github.com/h2non/gock>
[mock-get]: <https://pkg.go.dev/github.com/tkrop/go-testing/mock#Get>
[gomock]: <https://go.uber.org/mock>
[resty]: <https://github.com/go-resty/resty>
