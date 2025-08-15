package gock_test

import (
	"net/http"
	"net/url"

	"github.com/h2non/gock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// NewFooMatcher creates a special foo matcher.
func NewFooMatcher() *gock.MockMatcher {
	matcher := gock.NewEmptyMatcher()
	matcher.Add(func(req *http.Request, _ *gock.Request) (bool, error) {
		if req.URL.Scheme == "https" {
			return true, assert.AnError
		}
		return true, nil
	})
	matcher.Add(func(req *http.Request, _ *gock.Request) (bool, error) {
		return req.URL.Host == "foo.com", nil
	})
	matcher.Add(func(req *http.Request, _ *gock.Request) (bool, error) {
		return req.URL.Path == "/baz" || req.URL.Path == "/bar", nil
	})
	return matcher
}

// NewRoundTripperError creates the same round trip error usually returned in
// case of a transport error. This method is used for validating tests that are
// replacing the transport against the error `RoundTripper` via
// `NewErrorRoundTripper`.
func NewRoundTripperError(method string, _url string, err error) error {
	op := cases.Title(language.Und).String(method)
	return &url.Error{Op: op, URL: _url, Err: err}
}
