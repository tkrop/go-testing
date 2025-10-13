package gock

import (
	"net/http"

	"github.com/h2non/gock"
	"go.uber.org/mock/gomock"

	"github.com/tkrop/go-testing/test"
)

// ErrCannotMatch is the re-exported error for a failed match.
var ErrCannotMatch = gock.ErrCannotMatch

// Transport is a small transport implementation delegating requests to the
// owning HTTP request/response mock controller.
type Transport struct {
	// controller is the responsible HTTP request/response mock controller.
	controller *Controller
	// transport encapsulates the original transport interface for delegation.
	transport http.RoundTripper
}

// NewTransport creates a new *Transport with no responders.
func NewTransport(
	controller *Controller, transport http.RoundTripper,
) *Transport {
	return &Transport{
		controller: controller,
		transport:  transport,
	}
}

// RoundTrip delegates to the controller `RoundTrip` method.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	return t.controller.RoundTrip(req)
}

// Controller is a Gock based HTTP request/response mock controller.
type Controller struct {
	// The attached test context.
	t test.Test
	// MockStore the attached HTTP request/response mock storage.
	MockStore *MockStore
}

// NewGock creates a new HTTP request/response mock controller from the given
// go mock controller.
func NewGock(ctrl *gomock.Controller) *Controller {
	if t, ok := ctrl.T.(test.Test); ok {
		return NewController(t)
	}
	panic("gock not supported by test setup")
}

// NewController creates a new HTTP request/response mock controller.
func NewController(t test.Test) *Controller {
	ctrl := &Controller{
		t:         t,
		MockStore: NewStore(gock.NewMatcher()),
	}
	if c, ok := ctrl.t.(test.Cleanuper); ok {
		c.Cleanup(func() {
			ctrl.Cleanup()
		})
	}
	return ctrl
}

// New creates and registers a new HTTP request/response mock with given full
// qualified URI and default settings. It returns the request builder for
// setup of HTTP request and response mock details.
func (ctrl *Controller) New(uri string) *gock.Request {
	return ctrl.MockStore.NewMock(uri).Request()
}

// InterceptClient allows to intercept HTTP traffic of a custom http.Client that
// uses a non default http.Transport/http.RoundTripper implementation.
func (ctrl *Controller) InterceptClient(client *http.Client) {
	if _, ok := client.Transport.(*Transport); ok {
		return
	}
	client.Transport = NewTransport(ctrl, client.Transport)
}

// RestoreClient allows to disable and restore the original transport in the
// given http.Client.
func (*Controller) RestoreClient(client *http.Client) {
	if transport, ok := client.Transport.(*Transport); ok {
		client.Transport = transport.transport
	}
}

// RoundTrip receives HTTP requests and matches them against the registered
// HTTP request/response mocks. If a match is found it is used to construct the
// response, else the request is tracked as unmatched. If networing is enabled,
// the original transport is used to handle the request to .
//
// This method implements the `http.RoundTripper` interface and is used by
// attaching the controller to a `http.client` via `SetTransport`.
func (ctrl *Controller) RoundTrip(req *http.Request) (*http.Response, error) {
	// find matching mock for the incoming request.
	mock, err := ctrl.MockStore.Match(req)
	if err != nil {
		return nil, err
	} else if mock == nil {
		return nil, gock.ErrCannotMatch
	}
	defer ctrl.MockStore.Clean()

	return gock.Responder(req, mock.Response(), nil) //nolint:wrapcheck // transparent wrapper
}

// Cleanup checks if all the HTTP request/response mocks that were expected to
// be called have been called. This function is automatically registered with
// the test controller and will be called when the test is finished.
func (ctrl *Controller) Cleanup() {
	if pending := ctrl.MockStore.Pending(); len(pending) != 0 {
		for _, call := range pending {
			ctrl.t.Errorf("missing call(s) to %v", call)
		}
		ctrl.t.Errorf("aborting test due to missing call(s)")
	}
	ctrl.MockStore.Flush()
}
