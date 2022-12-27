package test_test

// File contains the split out logic to calculate the caller for automated
// testing of test failure validation.
import (
	"runtime"
	"strconv"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/tkrop/go-testing/mock"
	"github.com/tkrop/go-testing/test"
)

// Caller reporter that allows to capture the callers file and line number.
type Caller struct {
	path string
}

// Errorf is the caller reporter function to capture the callers file and line
// number of the `Errorf` call.
func (c *Caller) Errorf(format string, args ...any) {
	_, path, line, _ := runtime.Caller(1)
	c.path = path + ":" + strconv.Itoa(line)
}

// Fatalf is the caller reporter function to capture the callers file and line
// number of the `Fatalf` call.
func (c *Caller) Fatalf(format string, args ...any) {
	_, path, line, _ := runtime.Caller(1)
	c.path = path + ":" + strconv.Itoa(line)
	panic("finished") // prevents goexit.
}

// FailNow is the caller reporter function to capture the callers file and line
// number of the `FailNow` call.
func (c *Caller) FailNow() {
	_, path, line, _ := runtime.Caller(1)
	c.path = path + ":" + strconv.Itoa(line)
	panic("finished") // prevents goexit.
}

func (c *Caller) Panic(arg any) {
	_, path, line, _ := runtime.Caller(1)
	c.path = path + ":" + strconv.Itoa(line)
	panic("finished") // prevents goexit.
}

// getCaller implements the capturing logic for the callers file and line
// number for the given call.
func getCaller(call func(t test.Reporter)) string {
	t := test.NewTester(&testing.T{}, test.Failure)
	mocks := mock.NewMock(t)
	caller := mock.Get(mocks,
		func(ctrl *gomock.Controller) *Caller {
			return &Caller{}
		})
	t.Reporter(caller)
	func() {
		defer func() { recover() }()
		call(t)
	}()
	return caller.path
}

var (
	// CallerErrorf provides the file with line number of the `Errorf` call.
	CallerErrorf = getCaller(func(t test.Reporter) {
		t.Errorf("fail")
	})
	// CallerFatalf provides the file with line number of the `Fatalf` call.
	CallerFatalf = getCaller(func(t test.Reporter) {
		t.Fatalf("fail")
	})
	// CallerFailNow provides the file with line number of the `FailNow` call.
	CallerFailNow = getCaller(func(t test.Reporter) {
		t.FailNow()
	})
	// CallerPanic provides the file with line number of the `FailNow` call.
	CallerPanic = getCaller(func(t test.Reporter) {
		t.Panic("fail")
	})
)
