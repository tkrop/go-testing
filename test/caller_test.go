package test_test

// File contains the split out logic to calculate the caller for automated
// testing of test failure validation.
import (
	"os"
	"path"
	"runtime"
	"strconv"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/tkrop/go-testing/mock"
	"github.com/tkrop/go-testing/test"
)

// Caller reporter that allows to capture the callers file and line number.
type Caller struct {
	path string
}

// Error is the caller reporter function to capture the callers file and line
// number of the `Error` call.
func (c *Caller) Error(_ ...any) {
	_, path, line, _ := runtime.Caller(1)
	c.path = path + ":" + strconv.Itoa(line)
}

// Errorf is the caller reporter function to capture the callers file and line
// number of the `Errorf` call.
func (c *Caller) Errorf(_ string, _ ...any) {
	_, path, line, _ := runtime.Caller(1)
	c.path = path + ":" + strconv.Itoa(line)
}

// Fatal is the caller reporter function to capture the callers file and line
// number of the `Fatal` call.
func (c *Caller) Fatal(_ ...any) {
	_, path, line, _ := runtime.Caller(1)
	c.path = path + ":" + strconv.Itoa(line)
	panic("finished") // prevents goexit.
}

// Fatalf is the caller reporter function to capture the callers file and line
// number of the `Fatalf` call.
func (c *Caller) Fatalf(_ string, _ ...any) {
	_, path, line, _ := runtime.Caller(1)
	c.path = path + ":" + strconv.Itoa(line)
	panic("finished") // prevents goexit.
}

// Fail is the caller reporter function to capture the callers file and line
// number of the `Fail` call.
func (c *Caller) Fail() {
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

// Panic is the caller reporter function to capture the callers file and line
// number of the `Panic` call.
func (c *Caller) Panic(_ any) {
	_, path, line, _ := runtime.Caller(1)
	c.path = path + ":" + strconv.Itoa(line)
	panic("finished") // prevents goexit.
}

// getCaller implements the capturing logic for the callers file and line
// number for the given call.
func getCaller(call func(t test.Reporter)) string {
	t := test.New(&testing.T{}, test.Failure, false)
	mocks := mock.NewMocks(t)
	caller := mock.Get(mocks,
		func(*gomock.Controller) *Caller {
			return &Caller{}
		})
	t.Reporter(caller)
	func() {
		defer func() { _ = recover() }()
		call(t)
	}()
	return caller.path
}

var (
	// CallerError provides the file with line number of the `Error` call.
	CallerError = getCaller(func(t test.Reporter) {
		t.Error("fail")
	})
	// CallerErrorf provides the file with line number of the `Errorf` call.
	CallerErrorf = getCaller(func(t test.Reporter) {
		t.Errorf("%s", "fail")
	})
	// CallerFatal provides the file with line number of the `Fatal` call.
	CallerFatal = getCaller(func(t test.Reporter) {
		t.Fatal("fail")
	})
	// CallerFatalf provides the file with line number of the `Fatalf` call.
	CallerFatalf = getCaller(func(t test.Reporter) {
		t.Fatalf("%s", "fail")
	})
	// CallerFail provides the file with line number of the `Fail` call.
	CallerFail = getCaller(func(t test.Reporter) {
		t.Fail()
	})
	// CallerFailNow provides the file with line number of the `FailNow` call.
	CallerFailNow = getCaller(func(t test.Reporter) {
		t.FailNow()
	})
	// CallerPanic provides the file with line number of the `FailNow` call.
	CallerPanic = getCaller(func(t test.Reporter) {
		t.Panic("fail")
	})

	// Generic source directory for caller path evaluation.
	SourceDir = test.Must(os.Getwd())
	// CallerTestError provides the file with the line number of the `Error`
	// call in the test context implementation.
	CallerTestError = path.Join(SourceDir, "context.go:352")
	// CallerReporterErrorf provides the file with the line number of the
	// `Errorf` call in the test reporter/validator implementation.
	CallerReporterError = path.Join(SourceDir, "reporter.go:87")

	// CallerTestErrorf provides the file with the line number of the `Errorf`
	// call in the test context implementation.
	CallerTestErrorf = path.Join(SourceDir, "context.go:370")
	// CallerReporterErrorf provides the file with the line number of the
	// `Errorf` call in the test reporter/validator implementation.
	CallerReporterErrorf = path.Join(SourceDir, "reporter.go:109")
)
