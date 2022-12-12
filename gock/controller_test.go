package gock

import (
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"

	"github.com/tkrop/testing/test"
)

var testControllerParams = map[string]struct {
	url         string
	expectMatch test.Expect
	expectError error
}{
	"match with bar": {
		url:         "http://foo.com/bar",
		expectMatch: test.ExpectSuccess,
	},
	"match with baz": {
		url:         "http://foo.com/baz",
		expectMatch: test.ExpectSuccess,
	},
	"missing host": {
		url:         "http://bar.com/baz",
		expectError: gock.ErrCannotMatch,
	},
	"missing path": {
		url:         "http://foo.com/foo",
		expectError: gock.ErrCannotMatch,
	},
	"missing schema": {
		url:         "https://foo.com/bar",
		expectError: errAny,
	},
}

func TestController(t *testing.T) {
	t.Parallel()

	for message, param := range testControllerParams {
		message, param := message, param
		t.Run(message, func(t *testing.T) {
			t.Parallel()

			// Given
			tt := test.NewTestingT(t, param.expectMatch)
			ctrl := NewGock(gomock.NewController(tt))
			ctrl.MockStore.Matcher = NewFooMatcher()
			ctrl.New("http://foo.com").Get("/bar").Times(1).Reply(200)
			client := &http.Client{}

			// When
			ctrl.RestoreClient(client)
			ctrl.InterceptClient(client)
			ctrl.InterceptClient(client)
			response, err := client.Get(param.url)
			ctrl.RestoreClient(client)

			// Then
			if param.expectError != nil {
				assert.Equal(t, NewRoundTrippertError(
					http.MethodGet, param.url, param.expectError), err)
			} else {
				assert.NoError(t, err)
			}
			if param.expectMatch {
				assert.Equal(t, 200, response.StatusCode)
				assert.True(t, ctrl.MockStore.IsDone(), "mock done")
			} else {
				assert.False(t, ctrl.MockStore.IsDone(), "mock not done")
			}
			tt.Run(func(t test.Test) {
				ctrl.cleanup()
			})
		})
	}
}
