package gock

import (
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"

	"github.com/tkrop/testing/test"
)

type ControllerParams struct {
	url         string
	expectMatch test.Expect
	expectError error
}

var testControllerParams = map[string]ControllerParams{
	"match with bar": {
		url:         "http://foo.com/bar",
		expectMatch: test.Success,
	},
	"match with baz": {
		url:         "http://foo.com/baz",
		expectMatch: test.Success,
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
	test.Map(t, testControllerParams).
		Run(func(t test.Test, param ControllerParams) {
			// Given
			ctrl := NewGock(gomock.NewController(t))
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
			ctrl.cleanup()
		})
}

func TestPanic(t *testing.T) {
	// Given
	defer func() {
		if err := recover(); err == nil {
			assert.Fail(t, "did not panic")
		}
	}()

	// When
	NewGock(gomock.NewController(struct{ gomock.TestReporter }{}))

	// Then
	assert.Fail(t, "did not panic")
}
