package mock_test

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sync/atomic"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tkrop/go-testing/test"

	. "github.com/tkrop/go-testing/internal/mock"
)

type GenerateParams struct {
	file         string
	args         []string
	expectFile   string
	expectStdout string
	expectStderr string
	expectCode   int
}

var (
	// Test directory.
	testDirGenerate = func() string {
		dir, err := os.MkdirTemp("", "go-testing-*")
		if err != nil {
			panic(err)
		}
		return dir
	}()

	fileFailure = filepath.Join(testDirGenerate, dirTest, fileUnknown)
)

var testGenerateParams = map[string]GenerateParams{
	"iface": {
		file:       filepath.Join(testDirGenerate, MockFileDefault),
		args:       []string{pathTest},
		expectFile: expectIFace,
	},

	"failure parsing": {
		file: filepath.Join(testDirGenerate, MockFileDefault),
		args: []string{pathUnknown},
		expectStderr: "argument invalid [pos: 3, arg: " + pathUnknown +
			"]: not found\n",
		expectCode: 1,
	},

	"failure opening": {
		file: fileFailure,
		args: []string{pathTest},
		expectStderr: NewErrFileOpening(fileFailure, &fs.PathError{
			Op: "open", Path: fileFailure, Err: error(syscall.ENOENT),
		}).Error() + "\ninvalid argument\ninvalid argument\n",
		expectCode: 2,
	},
}

func TestGenerate(t *testing.T) {
	gen := NewGenerator(DirDefault, TargetDefault)
	test.Map(t, testGenerateParams).
		Run(func(t test.Test, param GenerateParams) {
			// Given
			args := []string{
				"--target-pkg=" + pkgMockTest,
				"--target-path=" + pathMock,
			}

			if param.file != "" {
				args = append(args, "--target-file="+param.file)
			}

			args = append(args, param.args...)

			outReader, outWriter, err := os.Pipe()
			require.NoError(t, err)
			errReader, errWriter, err := os.Pipe()
			require.NoError(t, err)

			// When
			code := atomic.Int32{}
			go func() {
				ret := gen.Generate(outWriter, errWriter, args...)
				code.Store(int32(ret))
				outWriter.Close()
				errWriter.Close()
			}()

			// Then
			buf, err := io.ReadAll(outReader)
			require.NoError(t, err)
			assert.Equal(t, param.expectStdout, string(buf))

			buf, err = io.ReadAll(errReader)
			require.NoError(t, err)
			assert.Equal(t, param.expectStderr, string(buf))

			if param.expectFile != "" {
				file, err := os.ReadFile(param.file)
				require.NoError(t, err)
				assert.Equal(t, param.expectFile, string(file))
			}
			assert.Equal(t, param.expectCode, int(code.Load()))
		}).
		Cleanup(func() {
			os.RemoveAll(testDirGenerate)
		})
}
