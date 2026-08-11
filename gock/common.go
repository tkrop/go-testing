package gock

import (
	"net/http"
	"net/url"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// TransportSetter interface to abstract the injection of `http.RoundTripper`
// compatible interfaces into http clients.
type TransportSetter interface {
	// SetTransport injects a custom `http.RoundTripper` compatible transport
	// interface into the http client.
	SetTransport(transport http.RoundTripper)
}

// RoundTripper functional interface to satisfy net/http.RoundTripper using an
// annonymous function.
type RoundTripper func(req *http.Request) (*http.Response, error)

// NewErrorRoundTripper static function to create new net/http.RoundTripper that
// can be used to test with default http errors while using simple anonymous
// functions.
func NewErrorRoundTripper(err error) RoundTripper {
	return RoundTripper(func(_ *http.Request) (*http.Response, error) {
		return nil, err
	})
}

// NewRoundTripperError creates the same round trip error usually returned in
// case of a transport error. This method is used for validating tests that are
// replacing the transport against the error `RoundTripper` via
// `NewErrorRoundTripper`.
func NewRoundTripperError(method string, _url string, err error) error {
	op := cases.Title(language.Und).String(method)
	return &url.Error{Op: op, URL: _url, Err: err}
}
