package mock

import (
	"fmt"
	"reflect"

	"github.com/golang/mock/gomock"

	"github.com/tkrop/go-testing/sync"
)

// DetachMode defines the mode for detaching mock calls.
type DetachMode int

const (
	// None mode to not detach mode.
	None DetachMode = 0
	// Head mode to detach head, i.e. do not order mock calls after predecessor
	// mock calls provided via context.
	Head DetachMode = 1
	// Tail mode to deteach tail, i.e. do not order mock calls before successor
	// mock calls provided via context.
	Tail DetachMode = 2
	// Both mode to detach tail and head, i.e. do neither order mock calls after
	// predecessor nor before successor provided via context.
	Both DetachMode = 3
)

// String return string representation of detach mode.
func (m DetachMode) String() string {
	switch m {
	// Case is not needed, yet!
	// case None:
	// 	return "None"
	case Head:
		return "Head"
	case Tail:
		return "Tail"
	case Both:
		return "Both"
	default:
		return "Unknown"
	}
}

type (
	// Call alias for `gomock.Call`.
	Call = gomock.Call
	// Controller alias for `gomock.Controller`.
	Controller = gomock.Controller

	// chain is the type to signal that mock calls must and will be orders in a
	// chain of mock calls.
	chain any
	// parallel is the type to signal that mock calls must and will be orders
	// in a parallel set of mock calls.
	parallel any
	// detachHead is the type to signal that the leading mock call must and
	// will be detached from its predecessor.
	detachHead any
	// detachTail is the type to signal that the trailing mock call must and
	// will be detached from its successor.
	detachTail any
	// detachBoth is the type to signal that the mock call must and will be
	// deteched from its predecessor as well as from its successor.
	detachBoth any
)

// SetupFunc common mock setup function signature.
type SetupFunc func(*Mocks) any

// Mocks common mock handler.
type Mocks struct {
	// The mock controller used.
	ctrl *Controller
	// The lenient wait group.
	wg sync.WaitGroup
	// The map of mock singletons.
	mocks map[reflect.Type]any
}

// NewMock creates a new mock handler using given test reporter (`*testing.T`).
func NewMock(t gomock.TestReporter) *Mocks {
	return (&Mocks{
		ctrl:  gomock.NewController(t),
		wg:    sync.NewLenientWaitGroup(),
		mocks: map[reflect.Type]any{},
	}).syncWith(t)
}

// Expect configures the mock handler to expect the given mock function calls.
func (mocks *Mocks) Expect(fncalls SetupFunc) *Mocks {
	if fncalls != nil {
		Setup(fncalls)(mocks)
	}
	return mocks
}

// syncWith used to synchronize the waitgroup of the mock setup with the wait
// group of the given test reporter. This function is called automatically on
// mock creation and therefore does not need to be called on the same reporter
// again.
func (mocks *Mocks) syncWith(t gomock.TestReporter) *Mocks {
	if s, ok := t.(sync.Synchronizer); ok {
		s.WaitGroup(mocks.wg)
	}
	return mocks
}

// Wait waits for all mock calls registered via `mocks.Times(<#>)` to be
// consumed before testing continuing. This method implements the `WaitGroup`
// interface to support testing of detached `go-routines` in an isolated
// [test](../test) environment.
func (mocks *Mocks) Wait() {
	mocks.wg.Wait()
}

// TODO: not needed yet - optional extension.
//
// Add adds the given delta on the waiting group handling the expected or
// consumed mock calls.  This method implements the `WaitGroup` interface to
// support testing of detached `go-routines` in an isolated [test](../test)
// environment.
//
// func (mocks *Mocks) Add(delta int) {
// 	mocks.wg.Add(delta)
// }

// TODO: not needed yet - optional extension.
//
// Done removes exactly one expected mock call from the wait group handling the
// expected or consumed mock calls. This method implements the `WaitGroup`
// interface to support testing of detached `go-routines` in an isolated
// [test](../test) environment.
//
// func (mocks *Mocks) Done() {
// 	mocks.wg.Done()
// }

// Times is creating the expectation that exactly the given number of mock call
// are consumed. This call is best provided as input for `Times`.
func (mocks *Mocks) Times(num int) int {
	mocks.wg.Add(num)
	return num
}

// GetDone is a convenience method for providing a standardized notification
// function call with the given number of arguments for `Do` to signal that a
// mock call setup was consumed.
func (mocks *Mocks) GetDone(numargs int) any {
	return mocks.GetFunc(numargs, func() { mocks.wg.Done() })
}

// GetPanic is a convenience method for providing a customized notification
// function call with the given number of arguments for `Do` to signal that a
// mock call setup was consumed and as result paniced.
func (mocks *Mocks) GetPanic(numargs int, reason string) any {
	return mocks.GetFunc(numargs, func() { mocks.wg.Done(); panic(reason) })
}

// GetFunc is a convenience method for providing a customized function call
// with the given number of arguments for `Do`.
func (mocks *Mocks) GetFunc(numargs int, fn func()) any {
	switch numargs {
	case 0:
		return func() { fn() }
	case 1:
		return func(any) { fn() }
	case 2:
		return func(any, any) { fn() }
	case 3:
		return func(any, any, any) { fn() }
	case 4:
		return func(any, any, any, any) { fn() }
	case 5:
		return func(any, any, any, any, any) { fn() }
	case 6:
		return func(any, any, any, any, any, any) { fn() }
	case 7:
		return func(any, any, any, any, any, any, any) { fn() }
	case 8:
		return func(any, any, any, any, any, any, any, any) { fn() }
	case 9:
		return func(any, any, any, any, any, any, any, any, any) { fn() }
	default:
		panic(fmt.Sprintf("argument number not supported: %d", numargs))
	}
}

// TODO: Reconsider this approach. Seems not to be helpful yet. Test setup
// functions would look as follows:
//
//	func GetTokenX(url string, err error) mock.SetupFunc {
//	  return mock.Mock(NewMockTokenProvider, func(mock *MockTokenProvider) *gomock.Call {
//	    return mock.EXPECT().GetToken(url).Return(token)
//	  })
//	}
//
// Mock defines an advanced mock setup function for exactly one mock call setup
// by resolving the singleton mock instance and handing it over to the provided
// function for calling the mock method and providing the return values. The
// created function automatically sets up the wait group for advanced testing
// strategies.
//
// func Mock[T any](
// 	creator func(*Controller) *T, caller func(*T) *gomock.Call,
// ) SetupFunc {
// 	return func(mocks *Mocks) any {
// 		call := caller(Get(mocks, creator)).
// 			Times(mocks.Times(1))
// 		value := reflect.ValueOf(call).Elem()
// 		field := value.FieldByName("methodType")
// 		ftype := *(*reflect.Type)(unsafe.Pointer(field.UnsafeAddr()))
// 		return call.Do(mocks.GetDone(ftype.NumIn()))
// 	}
// }

// Get resolves the actual mock from the mock handler by providing the
// constructor function generated by `gomock` to create a new mock.
func Get[T any](mocks *Mocks, creator func(*Controller) *T) *T {
	ctype := reflect.TypeOf(creator)
	mock, ok := mocks.mocks[ctype]
	if ok && mock != nil {
		return mock.(*T)
	}
	mock = creator(mocks.ctrl)
	mocks.mocks[ctype] = mock
	return mock.(*T)
}

// Setup creates only a lazily ordered set of mock calls that is detached from
// the parent setup by returning no calls for chaining. The mock calls created
// by the setup are only validated in so far in relation to each other, that
// `gomock` delivers results for the same mock call receiver in the order
// provided during setup.
func Setup(fncalls ...func(*Mocks) any) func(*Mocks) any {
	return func(mocks *Mocks) any {
		for _, fncall := range fncalls {
			inOrder([]*Call{}, []detachBoth{fncall(mocks)})
		}
		return nil
	}
}

// Chain creates a single chain of mock calls that is validated by `gomock`.
// If the execution order deviates from the order defined in the chain, the
// test validation fails. The method returns the full mock calls tree to allow
// chaining with other ordered setup method.
func Chain(fncalls ...func(*Mocks) any) func(*Mocks) any {
	return func(mocks *Mocks) any {
		calls := make([]chain, 0, len(fncalls))
		for _, fncall := range fncalls {
			calls = chainCalls(calls, fncall(mocks))
		}
		return calls
	}
}

// Parallel creates a set of parallel set of mock calls that is validated by
// `gomock`. While the parallel setup provids some freedom, this still defines
// constrainst with repect to parent and child setup methods, e.g. when setting
// up parallel chains in a chain, each parallel chains needs to follow the last
// mock call and finish before the following mock call.
//
// If the execution order deviates from the order defined by the parallel
// context, the test validation fails. The method returns the full set of mock
// calls to allow combining them with other ordered setup methods.
func Parallel(fncalls ...func(*Mocks) any) func(*Mocks) any {
	return func(mocks *Mocks) any {
		calls := make([]parallel, 0, len(fncalls))
		for _, fncall := range fncalls {
			calls = append(calls, fncall(mocks).(parallel))
		}
		return calls
	}
}

// Detach detach given mock call setup using given detach mode. It is possible
// to detach the mock call from the preceding mock calls (`Head`), from the
// succeeding mock calls (`Tail`), or from both as used in `Setup`.
func Detach(mode DetachMode, fncall func(*Mocks) any) func(*Mocks) any {
	return func(mocks *Mocks) any {
		switch mode {
		case None:
			return fncall(mocks)
		case Head:
			return []detachHead{fncall(mocks)}
		case Tail:
			return []detachTail{fncall(mocks)}
		case Both:
			return []detachBoth{fncall(mocks)}
		default:
			panic(ErrDetachMode(mode))
		}
	}
}

// Sub returns the sub slice of mock calls starting at index `from` up to index
// `to` inclduing. A negative value is used to calculate an index from the end
// of the slice. If the index of `from` is higher as the index `to`, the
// indexes are automatically switched. The returned sub slice of mock calls
// keeps its original semantic.
func Sub(from, to int, fncall func(*Mocks) any) func(*Mocks) any {
	return func(mocks *Mocks) any {
		calls := fncall(mocks)
		switch calls := calls.(type) {
		case *Call:
			inOrder([]*Call{}, calls)
			return GetSubSlice(from, to, []any{calls})
		case []chain:
			inOrder([]*Call{}, calls)
			return GetSubSlice(from, to, calls)
		case []parallel:
			inOrder([]*Call{}, calls)
			return GetSubSlice(from, to, calls)
		case []detachBoth:
			panic(ErrDetachNotAllowed(Both))
		case []detachHead:
			panic(ErrDetachNotAllowed(Head))
		case []detachTail:
			panic(ErrDetachNotAllowed(Tail))
		case nil:
			return nil
		default:
			panic(ErrNoCall(calls))
		}
	}
}

// GetSubSlice returns the sub slice of mock calls starting at index `from`
// up to index `to` inclduing. A negative value is used to calculate an index
// from the end of the slice. If the index `from` is after the index `to`, the
// indexes are automatically switched.
func GetSubSlice[T any](from, to int, calls []T) any {
	from = getPos(from, calls)
	to = getPos(to, calls)
	if from > to {
		return calls[to : from+1]
	} else if from < to {
		return calls[from : to+1]
	}
	return calls[from]
}

// getPos returns the actual call position evaluating negative positions
// from the back of the mock call slice.
func getPos[T any](pos int, calls []T) int {
	len := len(calls)
	if pos < 0 {
		pos = len + pos
		if pos < 0 {
			return 0
		}
		return pos
	} else if pos < len {
		return pos
	}
	return len - 1
}

// chainCalls joins arbitrary slices, single mock calls, and parallel mock calls
// into a single mock call slice and slice of mock slices. If the provided mock
// calls do not contain mock calls or slices of them, the join fails with a
// `panic`.
func chainCalls(calls []chain, more ...any) []chain {
	for _, call := range more {
		switch call := call.(type) {
		case *Call:
			calls = append(calls, call)
		case []chain:
			calls = append(calls, call...)
		case []parallel:
			calls = append(calls, call)
		case []detachBoth:
			calls = append(calls, call)
		case []detachHead:
			calls = append(calls, call)
		case []detachTail:
			calls = append(calls, call)
		case nil:
		default:
			panic(ErrNoCall(call))
		}
	}
	return calls
}

// inOrder creates an order of the given mock call using given anchors as
// predecessor and return the mock call as next anchor. The created order
// depends on the actual type of the mock call (slice).
func inOrder(anchors []*Call, call any) []*Call {
	switch call := call.(type) {
	case *Call:
		return inOrderCall(anchors, call)
	case []parallel:
		return inOrderParallel(anchors, call)
	case []chain:
		return inOrderChain(anchors, call)
	case []detachBoth:
		return inOrderDetachBoth(anchors, call)
	case []detachHead:
		return inOrderDetachHead(anchors, call)
	case []detachTail:
		return inOrderDetachTail(anchors, call)
	case nil:
		return anchors
	default:
		panic(ErrNoCall(call))
	}
}

// inOrderCall creates an order for the given mock call using the given achors
// as predecessor and resturn the call as next anchor.
func inOrderCall(anchors []*Call, call *Call) []*Call {
	if len(anchors) != 0 {
		for _, anchor := range anchors {
			if anchor != call {
				call.After(anchor)
			}
		}
	}
	return []*Call{call}
}

// inOrderChain creates a chain order of the given mock calls using given
// anchors as predecessor and return the last mocks call as next anchor.
func inOrderChain(anchors []*Call, calls []chain) []*Call {
	for _, call := range calls {
		anchors = inOrder(anchors, call)
	}
	return anchors
}

// inOrderParallel creates a parallel order the given mock calls using the
// anchors as predecessors and return list of all (last) mock calls as next
// anchors.
func inOrderParallel(anchors []*Call, calls []parallel) []*Call {
	nanchors := make([]*Call, 0, len(calls))
	for _, call := range calls {
		nanchors = append(nanchors, inOrder(anchors, call)...)
	}
	return nanchors
}

// inOrderDetachBoth creates a detached set of mock calls without using the
// anchors as predecessor nor returning the last mock calls as next anchor.
func inOrderDetachBoth(anchors []*Call, calls []detachBoth) []*Call {
	for _, call := range calls {
		inOrder(nil, call)
	}
	return anchors
}

// inOrderDetachHead creates a head detached set of mock calls without using
// the anchors as predecessor. The anchors are forwarded together with the new
// mock calls as next anchors.
func inOrderDetachHead(anchors []*Call, calls []detachHead) []*Call {
	for _, call := range calls {
		anchors = append(anchors, inOrder(nil, call)...)
	}
	return anchors
}

// inOrderDetachTail creates a tail detached set of mock calls using the
// anchors as predessors but without adding the mock calls as next anchors.
// The provided anchors are provided as next anchors.
func inOrderDetachTail(anchors []*Call, calls []detachTail) []*Call {
	for _, call := range calls {
		inOrder(anchors, call)
	}
	return anchors
}

// ErrNoCall creates an error with given call type to panic on inorrect call
// type.
func ErrNoCall(call any) error {
	return fmt.Errorf("type [%v] is not based on *gomock.Call",
		reflect.TypeOf(call))
}

// ErrDetachMode creates an error that the given detach mode is not supported.
func ErrDetachMode(mode DetachMode) error {
	return fmt.Errorf("detach mode [%v] is not supported", mode)
}

// ErrDetachNotAllowed creates an error that detach.
func ErrDetachNotAllowed(mode DetachMode) error {
	return fmt.Errorf("detach [%v] not supported in sub", mode)
}
