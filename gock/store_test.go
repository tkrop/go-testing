package gock_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tkrop/go-testing/gock"
	"github.com/tkrop/go-testing/test"
)

func TestStoreRegister(t *testing.T) {
	t.Parallel()

	// Given
	store := gock.NewStore(nil)

	// When
	mock := store.NewMock("foo")

	// Then
	mocks := store.All()
	assert.Len(t, mocks, 1)
	assert.Equal(t, mocks[0], mock)
	assert.Equal(t, mock.Request().Mock, mock)
	assert.Equal(t, mock.Response().Mock, mock)

	// When
	store.Register(mock)

	// Then
	mocks = store.All()
	assert.Len(t, mocks, 1)
	assert.Equal(t, mocks[0], mock)
}

func TestStoreGetAll(t *testing.T) {
	t.Parallel()

	// Given
	store := gock.NewStore(nil)
	mock := store.NewMock("foo")

	// When
	mocks := store.All()

	// Then
	assert.Len(t, mocks, 1)
	assert.Equal(t, mocks[0], mock)
	assert.Len(t, store.All(), 1)
}

func TestStoreExists(t *testing.T) {
	t.Parallel()

	// Given
	store := gock.NewStore(nil)
	mock := store.NewMock("foo")

	// When
	exists := store.Exists(mock)

	// Then
	mocks := store.All()
	assert.Len(t, mocks, 1)
	assert.True(t, exists, "mock exists")
	assert.Equal(t, mocks[0], mock)
}

func TestStorePending(t *testing.T) {
	t.Parallel()

	// Given
	store := gock.NewStore(NewFooMatcher())
	mock := store.NewMock("foo")
	done := store.NewMock("http://foo.com")
	done.Request().Get("/bar")
	ok, err := done.Match(&http.Request{Method: http.MethodGet, URL: &url.URL{
		Scheme: "http", Host: "foo.com", Path: "/baz",
	}})

	// Then
	mocks := store.All()
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Len(t, mocks, 2)

	// When
	pending := store.Pending()

	// Then
	mocks = store.All()
	assert.Len(t, mocks, 1)
	assert.Equal(t, mocks, pending)
	assert.Equal(t, mocks[0], mock)
}

func TestStoreIsPending(t *testing.T) {
	t.Parallel()

	// Given
	store := gock.NewStore(nil)
	store.NewMock("foo")

	// When
	pending := store.IsPending()

	// Then
	assert.True(t, pending, "mock pending")

	// When
	store.Flush()
	pending = store.IsPending()

	// Then
	assert.False(t, pending, "no mock pending")
}

func TestStoreIsDone(t *testing.T) {
	t.Parallel()

	// Given
	store := gock.NewStore(nil)
	store.NewMock("foo")

	// When
	done := store.IsDone()

	// Then
	assert.False(t, done, "mocks not done")

	// When
	store.Flush()
	done = store.IsDone()

	// Then
	assert.True(t, done, "mocks done")
}

func TestStoreRemove(t *testing.T) {
	t.Parallel()

	// Given
	store := gock.NewStore(nil)
	mock := store.NewMock("foo")

	// When
	exists := store.Exists(mock)

	// Then
	mocks := store.All()
	assert.Len(t, mocks, 1)
	assert.True(t, exists, "mock exists")

	// When
	store.Remove(mock)
	exists = store.Exists(mock)

	// Then
	assert.False(t, exists, "mock exists not")

	// When
	store.Remove(mock)
	exists = store.Exists(mock)

	// Then
	assert.False(t, exists, "mock exists not")
}

func TestStoreFlush(t *testing.T) {
	t.Parallel()

	// Given
	store := gock.NewStore(nil)
	mock1 := store.NewMock("foo")
	mock2 := store.NewMock("foo")

	// Then
	assert.Len(t, store.All(), 2)
	assert.True(t, store.Exists(mock1), "mock1 exists")
	assert.True(t, store.Exists(mock2), "mock2 exists")

	// When
	store.Flush()

	// Then
	assert.Len(t, store.All(), 0)
	assert.False(t, store.Exists(mock1), "mock1 exists not")
	assert.False(t, store.Exists(mock2), "mock2 exists not")
}

type MatchParams struct {
	url         string
	expectMatch bool
	expectError error
}

var matchTestCases = map[string]MatchParams{
	"match-with-bar": {
		url:         "http://foo.com/bar",
		expectMatch: true,
	},
	"match-with-baz": {
		url:         "http://foo.com/baz",
		expectMatch: true,
	},
	"missing-host": {
		url: "http://bar.com/baz",
	},
	"missing-path": {
		url: "http://foo.com/foo",
	},
	"missing-schema": {
		url:         "https://foo.com/bar",
		expectError: assert.AnError,
	},
}

func TestMatch(t *testing.T) {
	test.Map(t, matchTestCases).
		Run(func(t test.Test, param MatchParams) {
			// Given
			store := gock.NewStore(NewFooMatcher())
			mock := store.NewMock(param.url)

			// When
			uri, _ := url.Parse(param.url)
			match, err := store.Match(
				&http.Request{URL: uri})

			// Then
			if param.expectError != nil {
				assert.Equal(t, param.expectError, err)
			} else {
				assert.NoError(t, err)
			}
			if param.expectMatch {
				assert.Equal(t, match, mock)
			} else {
				assert.Nil(t, match)
			}
		})
}
