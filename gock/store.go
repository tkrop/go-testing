package gock

import (
	"net/http"
	"net/url"
	"sync"

	"github.com/h2non/gock"
)

// MockStore store for HTTP request/response mocks.
type MockStore struct {
	mocks []gock.Mock
	mutex sync.RWMutex
	// Matcher template used when creating new HTTP request/response mocks.
	Matcher *gock.MockMatcher
}

// NewStore creates a new mock registry for HTTP request/response mocks. If nil
// is provided as matcher the default matcher is used.
func NewStore(matcher *gock.MockMatcher) *MockStore {
	if matcher != nil {
		return &MockStore{Matcher: matcher}
	}
	return &MockStore{Matcher: gock.NewMatcher()}
}

// NewMock creates and registers a new HTTP request/response mock with default
// settings and returns the mock.
func (s *MockStore) NewMock(uri string) gock.Mock {
	res := gock.NewResponse()
	req := gock.NewRequest()
	req.URLStruct, res.Error = url.Parse(uri)

	// Create the new mock expectation
	mock := gock.NewMock(req, res)
	mock.SetMatcher(s.Matcher.Clone())
	s.Register(mock)
	return mock
}

// Register registers a new HTTP request/response mocks in the current stack.
func (s *MockStore) Register(mock gock.Mock) *MockStore {
	if s.Exists(mock) {
		return s
	}

	// Expose mock in request/response for delegation
	mock.Request().Mock = mock
	mock.Response().Mock = mock

	// Make ops thread safe
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Registers the mock in the global store
	s.mocks = append(s.mocks, mock)

	return s
}

// Match matches the given `http.Request` with the registered HTTP request
// mocks. It is returning the mock that matches or an error, if a matching
// function fails. If no HTTP mock request matches nothing is returned.
func (s *MockStore) Match(req *http.Request) (gock.Mock, error) {
	for _, mock := range s.All() {
		matches, err := mock.Match(req)
		if err != nil {
			return nil, err //nolint:wrapcheck // transparent wrapper
		}
		if matches {
			return mock, nil
		}
	}
	return nil, nil
}

// IsDone returns true if all the registered  HTTP request/response mocks have
// been triggered successfully.
func (s *MockStore) IsDone() bool {
	return !s.IsPending()
}

// IsPending returns true if there are pending HTTP request/response mocks.
func (s *MockStore) IsPending() bool {
	return len(s.Pending()) > 0
}

// All returns the current list of registered HTTP request/response mocks.
func (s *MockStore) All() []gock.Mock {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return append([]gock.Mock{}, s.mocks...)
}

// Pending returns a slice of the pending HTTP request/response mocks. As a
// side effect the mock store is cleaned up similar as calling `Clean` on it.
func (s *MockStore) Pending() []gock.Mock {
	return s.Clean().All()
}

// Exists checks if the given HTTP request/response mocks is already registered.
func (s *MockStore) Exists(m gock.Mock) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	for _, mock := range s.mocks {
		if mock == m {
			return true
		}
	}
	return false
}

// Remove removes a registered HTTP request/response mocks by reference.
func (s *MockStore) Remove(m gock.Mock) *MockStore {
	for i, mock := range s.mocks {
		if mock == m {
			s.mutex.Lock()
			s.mocks = append(s.mocks[:i], s.mocks[i+1:]...)
			s.mutex.Unlock()
		}
	}

	return s
}

// Flush flushes the current stack of registered HTTP request/response mocks.
func (s *MockStore) Flush() *MockStore {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.mocks = []gock.Mock{}

	return s
}

// Clean cleans the store removing disabled or obsolete HTTP request/response
// mocks.
func (s *MockStore) Clean() *MockStore {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	buf := []gock.Mock{}
	for _, mock := range s.mocks {
		if mock.Done() {
			continue
		}
		buf = append(buf, mock)
	}
	s.mocks = buf

	return s
}
