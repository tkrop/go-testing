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

// Log is the caller reporter function to capture the callers file and line
// number of the `Log` call.
func (c *Caller) Log(_ ...any) {
	_, path, line, _ := runtime.Caller(1)
	c.path = path + ":" + strconv.Itoa(line)
}

// Logf is the caller reporter function to capture the callers file and line
// number of the `Logf` call.
func (c *Caller) Logf(_ string, _ ...any) {
	_, path, line, _ := runtime.Caller(1)
	c.path = path + ":" + strconv.Itoa(line)
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
func getCaller(call func(t test.Panicer)) string {
	t := test.New(&testing.T{}, false).Expect(test.Failure)
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
	// CallerLog provides the file with line number of the `Log` call.
	CallerLog = getCaller(func(t test.Panicer) {
		t.Log("log")
	})
	// CallerLogf provides the file with line number of the `Logf` call.
	CallerLogf = getCaller(func(t test.Panicer) {
		t.Logf("%s", "log")
	})
	// CallerError provides the file with line number of the `Error` call.
	CallerError = getCaller(func(t test.Panicer) {
		t.Error("fail")
	})
	// CallerErrorf provides the file with line number of the `Errorf` call.
	CallerErrorf = getCaller(func(t test.Panicer) {
		t.Errorf("%s", "fail")
	})
	// CallerFatal provides the file with line number of the `Fatal` call.
	CallerFatal = getCaller(func(t test.Panicer) {
		t.Fatal("fail")
	})
	// CallerFatalf provides the file with line number of the `Fatalf` call.
	CallerFatalf = getCaller(func(t test.Panicer) {
		t.Fatalf("%s", "fail")
	})
	// CallerFail provides the file with line number of the `Fail` call.
	CallerFail = getCaller(func(t test.Panicer) {
		t.Fail()
	})
	// CallerFailNow provides the file with line number of the `FailNow` call.
	CallerFailNow = getCaller(func(t test.Panicer) {
		t.FailNow()
	})
	// CallerPanic provides the file with line number of the `FailNow` call.
	CallerPanic = getCaller(func(t test.Panicer) {
		t.Panic("fail")
	})

	// Generic source directory for caller path evaluation.
	SourceDir = test.Must(os.Getwd())
	// CallerTestError provides the file with the line number of the `Error`
	// call in the test context implementation.
	CallerTestError = path.Join(SourceDir, "context.go:344")
	// CallerReporterErrorf provides the file with the line number of the
	// `Errorf` call in the test reporter/validator implementation.
	CallerReporterError = path.Join(SourceDir, "reporter.go:135")

	// CallerTestErrorf provides the file with the line number of the `Errorf`
	// call in the test context implementation.
	CallerTestErrorf = path.Join(SourceDir, "context.go:362")
	// CallerReporterErrorf provides the file with the line number of the
	// `Errorf` call in the test reporter/validator implementation.
	CallerReporterErrorf = path.Join(SourceDir, "reporter.go:158")
)
