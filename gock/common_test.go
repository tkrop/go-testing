package gock

import (
	"errors"
	"net/http"
	"net/url"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	gock "gopkg.in/h2non/gock.v1"
)

var errAny = errors.New("any error")

// NewFooMatcher creates a special foo matcher.
func NewFooMatcher() *gock.MockMatcher {
	matcher := gock.NewEmptyMatcher()
	matcher.Add(func(req *http.Request, ereq *gock.Request) (bool, error) {
		if req.URL.Scheme == "https" {
			return true, errAny
		}
		return true, nil
	})
	matcher.Add(func(req *http.Request, ereq *gock.Request) (bool, error) {
		return req.URL.Host == "foo.com", nil
	})
	matcher.Add(func(req *http.Request, ereq *gock.Request) (bool, error) {
		return req.URL.Path == "/baz" || req.URL.Path == "/bar", nil
	})
	return matcher
}

// NewRoundTrippertError creates the same round trip error usually returned in
// case of a transport error. This method is used for validating tests that are
// replacing the transport against the error `RoundTripper` via
// `NewErrorRoundTripper`.
func NewRoundTrippertError(method string, URL string, err error) error {
	op := cases.Title(language.Und).String(method)
	return &url.Error{Op: op, URL: URL, Err: err}
}
