package gock

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStoreRegister(t *testing.T) {
	t.Parallel()

	// Given
	store := NewStore(nil)

	// When
	mock := store.NewMock("foo")

	// Then
	assert.Len(t, store.mocks, 1)
	assert.Equal(t, store.mocks[0], mock)
	assert.Equal(t, mock.Request().Mock, mock)
	assert.Equal(t, mock.Response().Mock, mock)

	// When
	store.Register(mock)
	assert.Len(t, store.mocks, 1)
	assert.Equal(t, store.mocks[0], mock)
}

func TestStoreGetAll(t *testing.T) {
	t.Parallel()

	// Given
	store := NewStore(nil)
	mock := store.NewMock("foo")

	// WHen
	mocks := store.All()

	// Then
	assert.Len(t, store.mocks, 1)
	assert.Len(t, mocks, 1)
	assert.Equal(t, mocks[0], mock)
}

func TestStoreExists(t *testing.T) {
	t.Parallel()

	// Given
	store := NewStore(nil)
	mock := store.NewMock("foo")

	// When
	exists := store.Exists(mock)

	assert.Len(t, store.mocks, 1)
	assert.True(t, exists, "mock exists")
}

func TestStorePending(t *testing.T) {
	t.Parallel()

	// Given
	store := NewStore(NewFooMatcher())
	mock := store.NewMock("foo")
	done := store.NewMock("http://foo.com")
	done.Request().Get("/bar")
	done.Match(&http.Request{Method: http.MethodGet, URL: &url.URL{
		Scheme: "http", Host: "foo.com", Path: "/baz",
	}})

	// Then
	assert.Len(t, store.mocks, 2)

	// When
	pending := store.Pending()

	// Then
	assert.Len(t, store.mocks, 1)
	assert.Equal(t, store.mocks, pending)
	assert.Equal(t, store.mocks[0], mock)
}

func TestStoreIsPending(t *testing.T) {
	t.Parallel()

	// Given
	store := NewStore(nil)
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
	store := NewStore(nil)
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
	store := NewStore(nil)
	mock := store.NewMock("foo")

	// When
	exists := store.Exists(mock)

	// Then
	assert.Len(t, store.mocks, 1)
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
	store := NewStore(nil)
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

var testMatchParams = map[string]struct {
	url         string
	expectMatch bool
	expectError error
}{
	"match with bar": {
		url:         "http://foo.com/bar",
		expectMatch: true,
	},
	"match with baz": {
		url:         "http://foo.com/baz",
		expectMatch: true,
	},
	"missing host": {
		url: "http://bar.com/baz",
	},
	"missing path": {
		url: "http://foo.com/foo",
	},
	"missing schema": {
		url:         "https://foo.com/bar",
		expectError: errAny,
	},
}

func TestMatch(t *testing.T) {
	t.Parallel()
	for message, param := range testMatchParams {
		message, param := message, param
		t.Run(message, func(t *testing.T) {
			t.Parallel()

			// Given
			store := NewStore(NewFooMatcher())
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
}
