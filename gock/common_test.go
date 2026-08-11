package gock_test

import (
	"errors"
	"net/http"
	"testing"

	h2gock "github.com/h2non/gock"
	"github.com/stretchr/testify/assert"

	"github.com/tkrop/go-testing/gock"
	"github.com/tkrop/go-testing/test"
)

// NewFooMatcher creates a special foo matcher.
func NewFooMatcher() *h2gock.MockMatcher {
	matcher := h2gock.NewEmptyMatcher()
	matcher.Add(func(req *http.Request, _ *h2gock.Request) (bool, error) {
		if req.URL.Scheme == "https" {
			return true, assert.AnError
		}
		return true, nil
	})
	matcher.Add(func(req *http.Request, _ *h2gock.Request) (bool, error) {
		return req.URL.Host == "foo.com", nil
	})
	matcher.Add(func(req *http.Request, _ *h2gock.Request) (bool, error) {
		return req.URL.Path == "/baz" || req.URL.Path == "/bar", nil
	})
	return matcher
}

type ErrorRoundTripperParams struct {
	err    error
	expect error
}

var errorRoundTripperTestCases = map[string]ErrorRoundTripperParams{
	"with-error": {
		err:    assert.AnError,
		expect: assert.AnError,
	},
	"with-custom-error": {
		err:    errors.New("custom error"),
		expect: errors.New("custom error"),
	},
	"with-nil-error": {
		err:    nil,
		expect: nil,
	},
}

func TestNewErrorRoundTripper(t *testing.T) {
	test.Map(t, errorRoundTripperTestCases).
		Run(func(t test.Test, param ErrorRoundTripperParams) {
			// Given
			roundTripper := gock.NewErrorRoundTripper(param.err)

			// When
			response, err := roundTripper.RoundTrip(
				&http.Request{Method: http.MethodGet})

			// Then
			assert.Nil(t, response)
			assert.Equal(t, param.expect, err)
		})
}
