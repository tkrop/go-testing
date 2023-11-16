// Package mock contains the basic collection of functions and types for
// controlling mocks and mock request/response setup. It is part of the public
// interface and starting to get stable, however, we are still experimenting
// to optimize the interface and the user experience.
package mock

import (
	"errors"
	"fmt"

	"github.com/golang/mock/gomock"

	"github.com/tkrop/go-testing/internal/reflect"
	"github.com/tkrop/go-testing/internal/sync"
)

// DetachMode defines the mode for detaching mock calls.
type DetachMode int

const (
	// None mode to not detach mode.
	None DetachMode = 0
	// Head mode to detach head, i.e. do not order mock calls after predecessor
	// mock calls provided via context.
	Head DetachMode = 1
	// Tail mode to detach tail, i.e. do not order mock calls before successor
	// mock calls provided via context.
	Tail DetachMode = 2
	// Both mode to detach tail and head, i.e. do neither order mock calls after
	// predecessor nor before successor provided via context.
	Both DetachMode = 3
)

// String return string representation of detach mode.
func (m DetachMode) String() string {
	switch m { //nolint:exhaustive // case is not needed, yet!
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
	// Call alias for [gomock.Call].
	Call = gomock.Call
	// Controller alias for [gomock.Controller].
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
	// detached from its predecessor as well as from its successor.
	detachBoth any
)

// SetupFunc common mock setup function signature.
type SetupFunc func(*Mocks) any

// Mocks common mock handler.
type Mocks struct {
	// The mock controller used.
	Ctrl *Controller
	// The lenient wait group.
	wg sync.WaitGroup
	// The map of mock singletons.
	mocks map[reflect.Value]any
	// A map of mock key value pairs.
	args map[any]any
}

// NewMocks creates a new mock handler using given test reporter, e.g.
// [*testing.T], or [test.Test].
func NewMocks(t gomock.TestReporter) *Mocks {
	return (&Mocks{
		Ctrl:  gomock.NewController(t),
		wg:    sync.NewLenientWaitGroup(),
		mocks: map[reflect.Value]any{},
		args:  map[any]any{},
	}).syncWith(t)
}

// Get resolves the singleton mock from the mock handler by providing the
// reflection value of the constructor function generated by [gomock] to create
// a new mock. The mock is only created once and stored in an internal creator
// function to mock map.
func (mocks *Mocks) Get(creator reflect.Value) any {
	mock, ok := mocks.mocks[creator]
	if ok && mock != nil {
		return mock
	}
	mock = reflect.ArgOf(creator.Call(
		reflect.ValuesIn(creator.Type(), mocks.Ctrl))[0])
	mocks.mocks[creator] = mock
	return mock
}

// Expect configures the mock handler to expect the given mock function calls.
func (mocks *Mocks) Expect(fncalls SetupFunc) *Mocks {
	if fncalls != nil {
		Setup(fncalls)(mocks)
	}
	return mocks
}

// GetArg gets the mock argument value for the given argument key. This can be
// used to access a common test arguments from a mock call.
func (mocks *Mocks) GetArg(key any) any {
	return mocks.args[key]
}

// SetArg sets the given mock argument value for the given argument key. This
// can be used to pass a common test arguments to mock calls.
func (mocks *Mocks) SetArg(key any, value any) *Mocks {
	mocks.args[key] = value
	return mocks
}

// SetArgs sets the given mock argument values for the given argument keys.
// This can be used to pass a set of common test arguments to mock calls.
func (mocks *Mocks) SetArgs(args map[any]any) *Mocks {
	for key, value := range args {
		mocks.args[key] = value
	}
	return mocks
}

// syncWith used to synchronize the wait group of the mock setup with the wait
// group of the given test reporter. This function is called automatically on
// mock creation and therefore does not need to be called on the same reporter
// again.
func (mocks *Mocks) syncWith(t gomock.TestReporter) *Mocks {
	if s, ok := t.(sync.Synchronizer); ok {
		s.WaitGroup(mocks.wg)
	}
	return mocks
}

// Wait waits for all mock calls registered via [Call], [Do], [Return],
// [Panic], and [Times] to be consumed before testing can continue. This method
// implements the [sync.WaitGroup] interface to support testing of detached
// *goroutines* in an isolated [test](../test) environment.
func (mocks *Mocks) Wait() {
	mocks.wg.Wait()
}

// Add adds the given delta on the wait group to register the expected or
// notify the consumed mock calls. This method implements the [sync.WaitGroup]
// interface to support testing of detached *goroutines* in an isolated
// [test](../test) environment.
//
// **Note:** Usually call expectation setup is completely handled via `Call`,
// `Do`, `Return`, and `Panic`. Use this method only for synchronizing tests
// *goroutines*.
func (mocks *Mocks) Add(delta int) int {
	mocks.wg.Add(delta)
	return delta
}

// Done removes exactly one expected mock call from the wait group to notify
// a consumed mock call. This method implements the [sync.WaitGroup] interface
// to support testing of detached `go-routines` in an isolated [test](../test)
// environment.
//
// **Note:** Usually call expectation setup is completely handled via `Call`,
// `Do`, `Return`, and `Panic`. Use this method only for synchronizing tests
// *goroutines*.
func (mocks *Mocks) Done() {
	mocks.wg.Done()
}

// Times is creating the expectation that exactly the given number of mock call
// are consumed. This call is supposed to be used as input for [gomock.Times]
// in combination with [Call], [Do], [Return], and [Panic]. Setting up [Times]
// is considering that these methods add one expected call by reducing the
// registration by one.
func (mocks *Mocks) Times(num int) int {
	mocks.wg.Add(num - 1)
	return num
}

// Call is a convenience method to setup a call back function for [gomock.Do]
// and [gomock.DoAndReturn]. Using this method signals an expected mock call
// during setup as well as a consumed mock call when executing the given call
// back function. The function is supplied with the regular call parameters and
// expected to return the mock result - if required, as [gomock.Do] ignores
// arguments.
//
// **Note:** Call registers exactly one expected call automatically.
func (mocks *Mocks) Call(fn any, call func(...any) []any) any {
	return mocks.notify(fn, false, call)
}

// Do is a convenience method to setup a call back function for [gomock.Do]
// or [gomock.DoAndReturn]. Using this method signals an expected mock call
// during setup as well as a consumed mock call when executing the given call
// back function returning the given optional arguments as mock result - if
// necessary, as [gomock.Do] ignores arguments.
//
// **Note:** Do registers exactly one expected call automatically.
func (mocks *Mocks) Do(fn any, args ...any) any {
	return mocks.notify(fn, true, nil, args...)
}

// Return is a convenience method to setup a call back function for [gomock.Do]
// or [gomock.DoAndReturn]. Using this method signals an expected mock call
// during setup as well as a consumed mock call when executing the given call
// back function returning the given optional arguments as mock result - if
// necessary, as [gomock.Do] ignores arguments.
//
// **Note:** Return registers exactly one expected call automatically.
func (mocks *Mocks) Return(fn any, args ...any) any {
	return mocks.notify(fn, false, nil, args...)
}

// Panic is a convenience method to setup a call back function that panics with
// given reason for [gomock.Do] or [gomock.DoAndReturn]. Using this method
// signals an expected mock call during setup as well as a consumed mock call
// when executing the given call back function.
//
// **Note:** Return registers exactly one expected call automatically.
func (mocks *Mocks) Panic(fn any, reason any) any {
	return mocks.notify(fn, false, func(...any) []any { panic(reason) })
}

// notify is a generic method for providing a customized notification function
// of the given function call type with given custom call behavior and given
// return arguments for usage in [gomock.Do] or `[gomock.DoAndReturn]. When
// notify is called, it registers exactly one expected service call that is
// consumed, when the created call back function is called. If lenient is
// given, the provided output arguments are not enforced to match the number
// of return arguments in the notification function.
func (mocks *Mocks) notify(
	fn any, lenient bool, call func(...any) []any, args ...any,
) any {
	mocks.wg.Add(1)

	ftype := reflect.TypeOf(fn)
	btype := reflect.BaseFuncOf(ftype, 1, 0)
	notify := reflect.MakeFuncOf(btype,
		func(in []reflect.Value) []reflect.Value {
			mocks.Ctrl.T.Helper()

			defer mocks.wg.Done()
			if call != nil {
				args = call(reflect.ArgsOf(in...)...)
			}

			return reflect.ValuesOut(ftype, lenient, args...)
		})

	return notify
}

// TODO: Reconsider approach - complex signature. Test setup look as follows:
//
//	func CallBX(input string, output string) mock.SetupFunc {
//		return mock.Mock(NewMockIFace, func(mock *MockIFace) *gomock.Call {
//			return mock.EXPECT().CallB(input).Return(output)
//		})
//	}
//
// Mock defines an advanced mock setup function for exactly one mock call setup
// by resolving the singleton mock instance and handing it over to the provided
// function for calling the mock method and providing the return values. The
// created function automatically sets up the wait group for advanced testing
// strategies.
//
// func Mock[T any](
// 	creator func(*Controller) *T, call func(*T) *gomock.Call,
// ) SetupFunc {
// 	return func(mocks *Mocks) any {
// 		call := call(Get(mocks, creator))
// 		value := reflect.ValueOf(call).Elem()
// 		field := value.FieldByName("methodType")
// 		ftype := *(*reflect.Type)(unsafe.Pointer(field.UnsafeAddr()))
// 		btype := reflect.BaseFuncOf(ftype, 0, ftype.NumOut())
// 		return call.Do(mocks.notify(btype, nil))
// 	}
// }

// Get resolves the actual mock from the mock handler by providing the
// constructor function generated by `gomock` to create a new mock.
func Get[T any](mocks *Mocks, creator func(*Controller) *T) *T {
	//revive:disable-next-line:unchecked-type-assertion // cannot happen
	return mocks.Get(reflect.ValueOf(creator)).(*T)
}

// TODO: decide on mock strategy.
//
// Expect resolves the mock recorder from the mock handler by providing the
// constructor function generated by `gomock`.
// func Expect[T any](mocks *Mocks, creator func(*Controller) any) T {
// 	return reflect.ValueOf(mocks.Get(creator)).
// 		MethodByName("EXPECT").Call(nil)[0].Interface().(T)
// }

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
// `gomock`. While the parallel setup provides some freedom, this still defines
// constraints with respect to parent and child setup methods, e.g. when setting
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
			panic(NewErrDetachMode(mode))
		}
	}
}

// Sub returns the sub slice of mock calls starting at index `from` up to index
// `to` including. A negative value is used to calculate an index from the end
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
			panic(NewErrDetachNotAllowed(Both))
		case []detachHead:
			panic(NewErrDetachNotAllowed(Head))
		case []detachTail:
			panic(NewErrDetachNotAllowed(Tail))
		case nil:
			return nil
		default:
			panic(NewErrNoCall(calls))
		}
	}
}

// GetSubSlice returns the sub slice of mock calls starting at index `from`
// up to index `to` including. A negative value is used to calculate an index
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
	end := len(calls)
	if pos < 0 {
		pos = end + pos
		if pos < 0 {
			return 0
		}
		return pos
	} else if pos < end {
		return pos
	}
	return end - 1
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
			panic(NewErrNoCall(call))
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
		panic(NewErrNoCall(call))
	}
}

// inOrderCall creates an order for the given mock call using the given anchors
// as predecessor and return the call as next anchor.
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
// anchors as predecessors but without adding the mock calls as next anchors.
// The provided anchors are provided as next anchors.
func inOrderDetachTail(anchors []*Call, calls []detachTail) []*Call {
	for _, call := range calls {
		inOrder(anchors, call)
	}
	return anchors
}

var (
	// Error type for unsupported type errors.
	ErrTypeNotSupported = errors.New("type not supported")

	// Error type for unsupported mode errors.
	ErrModeNotSupprted = errors.New("mode not supported")
)

// NewErrNoCall creates an error with given call type to panic on incorrect
// call type.
func NewErrNoCall(call any) error {
	return fmt.Errorf("%w [type: %v] must be *gomock.Call",
		ErrTypeNotSupported, reflect.TypeOf(call))
}

// NewErrDetachMode creates an error that the given detach mode is not
// supported.
func NewErrDetachMode(mode DetachMode) error {
	return fmt.Errorf("%w [mode: %v]",
		ErrModeNotSupprted, mode)
}

// NewErrDetachNotAllowed creates an error that the detach mode is not
// supported.
func NewErrDetachNotAllowed(mode DetachMode) error {
	return fmt.Errorf("%w [mode: %v] not supported in sub",
		ErrModeNotSupprted, mode)
}
