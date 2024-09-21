package test_test

// File contains the split out logic to calculate the caller for automated
// testing of test failure validation.
import (
	"os"
	"path"
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
func (c *Caller) Errorf(_ string, _ ...any) {
	_, path, line, _ := runtime.Caller(1)
	c.path = path + ":" + strconv.Itoa(line)
}

// Fatalf is the caller reporter function to capture the callers file and line
// number of the `Fatalf` call.
func (c *Caller) Fatalf(_ string, _ ...any) {
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

func (c *Caller) Panic(_ any) {
	_, path, line, _ := runtime.Caller(1)
	c.path = path + ":" + strconv.Itoa(line)
	panic("finished") // prevents goexit.
}

// getCaller implements the capturing logic for the callers file and line
// number for the given call.
func getCaller(call func(t test.Reporter)) string {
	t := test.NewTester(&testing.T{}, test.Failure)
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

	// Generic source directory for caller path evaluation.
	SourceDir = func() string {
		dir, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		return dir
	}()
	// CallerTestErrorf provides the file with the line number of the `Errorf`
	// call in testing.
	CallerTestErrorf = path.Join(SourceDir, "testing.go:206")
	// CallerGomockErrorf provides the file with the line number of the
	// `Errorf` call in gomock.
	CallerGomockErrorf = path.Join(SourceDir, "gomock.go:61")
)
