# Usage patterns of testing/gock

This package provides a small controller and mock storage to isolate testing of
HTTP request/response cycles extending [Gock][gock]. Since framework is focused
on testing it does not support the networking and observation features of
[Gock][gock] and requires manual transport interception, however, the interface
is mainly compatible with.


## Migration from Gock to Tock

Migration from [Gock][gock] to this package is straigt forward. You can just
add the controller creation as at the begin of your test giving it the name
`gock` and hand it over to all methods creating HTTP request/response mocks. The
mock creation than happens as usual.

```go
func MyTest(t *testing.T) {
	gock := tock.Controller(t)

	...

	gock.New("http://foo.com").Get("/bar").Times(1).
		Reply(200).BodyString("result")

	...
}
```

Since the controller does not intercept all transports by default, you need to
setup transport interception manually. This can happen in three different ways.
If you have access to the HTTP request/response client, you can either use the
usual `InterceptClient` (and `RestoreClient`) methods.

```go
func MyTest(t *testing.T) {
	gock := tock.Controller(t)

	...

	client := &http.Client{}
	gock.InterceptClient(client)
	defer gock.RestoreClient(client) // optional

	...
}
```

Some customized HTTP clients e.g. [resty][resty] offer the ability to set the
transport manually based on `http.RoundTripper` interface. The controller
supports customized HTTP clients by implementing `http.RoundTripper` so that
you can utilize their setup methods.

```go
func MyTest(t *testing.T) {
	gock := tock.Controller(t)

	...

	client := resty.New()
	client.setTransport(gock)

	...
}
```

As last resort, you can also intercept the `http.DefaultTransport`, however,
this is not advised, since it will destroy the test isolation that is goal of
this wrapper framework. In this case you should use
[gock][gock] directly.


## Integration with `mock`-framework

The `Gock`-controller framework supports a simple integration with the
[mock](../mock) framework for [gomock][gomock]: it simply provides constructor
that accepts the `gomock`-controller. Using constructor, it is possible to
create the usual setup methods similar as described in
[mock](../mock#generic-mock-service-call-pattern).

```go
func GockCall(
	url, path string, input..., status int, output..., error,
) mock.SetupFunc {
    return func(mocks *Mocks) any {
        mock.Get(mocks, gock.NewGock).New(url).Get(path).Times(1).
			Reply(status)...
		return nil
    }
}
```

**Note:** While this already nicely integrates the mock controller creation,
call setup, and validation, it currently provides no support for call order
validation as [GoMock][gomock] supports it.


[gomock]: https://github.com/golang/mock "GoMock"
[gock]: https://github.com/h2non/gock "Gock"
[resty]: https://github.com/go-resty/resty "Resty"
