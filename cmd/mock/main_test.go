package main

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

	"github.com/tkrop/go-testing/internal/mock"
	"github.com/tkrop/go-testing/test"
)

var (
	// Test directory.
	testDir = func() string {
		dir, err := os.MkdirTemp("", "go-testing-*")
		if err != nil {
			panic(err)
		}
		return dir
	}()

	// Expected IFace mock.
	expectIFace = func() string {
		file, err := os.ReadFile("../../internal/mock/mock_iface_test.gox")
		if err != nil {
			panic(err)
		}
		return string(file)
	}()

	pathTest    = "github.com/tkrop/go-testing/internal/mock/test"
	pathUnknown = "github.com/tkrop/go-testing/internal/mock/unknown"
	fileUnknown = filepath.Join(testDir, "unknown", "mock_iface_test.go")
)

type MainParam struct {
	file         string
	args         []string
	expectFile   string
	expectStdout string
	expectStderr string
	expectCode   int
}

var testMainParams = map[string]MainParam{
	"iface": {
		file:       filepath.Join(testDir, "mock_iface_test.go"),
		args:       []string{pathTest},
		expectFile: expectIFace,
	},

	"failure parsing": {
		file: filepath.Join(testDir, "mock_iface_test.go"),
		args: []string{pathUnknown},
		expectStderr: "argument failure [pos: 3, arg: *] => " +
			"package parsing [path: " + pathUnknown + "] => " +
			"[-: no required module provides package " + pathUnknown +
			"; to add it:\n\tgo get " + pathUnknown + "]\n",
		expectCode: 1,
	},

	"failure opening": {
		file: fileUnknown,
		args: []string{pathTest},
		expectStderr: mock.NewErrFileOpening(fileUnknown, &fs.PathError{
			Op: "open", Path: fileUnknown, Err: error(syscall.ENOENT),
		}).Error() + "\ninvalid argument\ninvalid argument\n",
		expectCode: 3,
	},
}

func TestMain(t *testing.T) {
	test.Map(t, testMainParams).
		Run(func(t test.Test, param MainParam) {
			// Given
			args := []string{
				"--target-pkg=mock",
				"--target-path=github.com/tkrop/go-testing/internal/mock",
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
				ret := run(outWriter, errWriter, args...)
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
			os.RemoveAll(testDir)
		})
}
