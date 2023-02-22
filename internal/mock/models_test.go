package mock_test

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/huandu/go-clone"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "github.com/tkrop/go-testing/internal/mock"
	"github.com/tkrop/go-testing/mock"
	"github.com/tkrop/go-testing/test"
)

//go:generate mockgen -package=mock_test -destination=mock_template_test.go -source=template.go Template

var (
	targetDefault = &Type{
		Package: pkgMock, Path: pathMock, File: fileMock,
	}
	targetOtherPkg  = targetDefault.With(&Type{Package: pkgMockTest})
	targetOtherPath = targetDefault.With(&Type{Path: pathTest})
	targetOtherName = targetDefault.With(nameIFace)
	targetOtherFile = targetDefault.With(&Type{File: fileOther})
)

func Execute(writer io.Writer, file *File, err error) mock.SetupFunc {
	return func(mocks *mock.Mocks) any {
		return mock.Get(mocks, NewMockTemplate).EXPECT().
			Execute(writer, file).DoAndReturn(mocks.Call(
			Template.Execute,
			func(args ...any) []any {
				if err != nil {
					return []any{err}
				}
				return []any{template.Execute(
					args[0].(io.Writer), args[1].(*File),
				)}
			}))
	}
}

type NewFilesParams struct {
	mocks       []*Mock
	expectFiles []*File
}

var testNewFilesParams = map[string]NewFilesParams{
	"target once": {
		mocks: []*Mock{{Target: targetDefault}},
		expectFiles: []*File{{
			Target: targetDefault,
			Mocks:  []*Mock{{Target: targetDefault}},
		}},
	},

	"target twice": {
		mocks: []*Mock{
			{Target: targetDefault}, {Target: targetDefault},
		},
		expectFiles: []*File{{
			Target: targetDefault,
			Mocks: []*Mock{
				{Target: targetDefault}, {Target: targetDefault},
			},
		}},
	},

	"target other name (ignored)": {
		mocks: []*Mock{
			{Target: targetDefault}, {Target: targetOtherName},
		},
		expectFiles: []*File{{
			Target: targetDefault,
			Mocks: []*Mock{
				{Target: targetDefault}, {Target: targetOtherName},
			},
		}},
	},

	"target other package": {
		mocks: []*Mock{
			{Target: targetDefault}, {Target: targetOtherPkg},
		},
		expectFiles: []*File{{
			Target: targetDefault,
			Mocks:  []*Mock{{Target: targetDefault}},
		}, {
			Target: targetOtherPkg,
			Mocks:  []*Mock{{Target: targetOtherPkg}},
		}},
	},

	"target other file": {
		mocks: []*Mock{
			{Target: targetDefault}, {Target: targetOtherFile},
		},
		expectFiles: []*File{{
			Target: targetDefault,
			Mocks:  []*Mock{{Target: targetDefault}},
		}, {
			Target: targetOtherFile,
			Mocks:  []*Mock{{Target: targetOtherFile}},
		}},
	},

	"target other path": {
		mocks: []*Mock{
			{Target: targetDefault}, {Target: targetOtherPath},
		},
		expectFiles: []*File{{
			Target: targetDefault,
			Mocks:  []*Mock{{Target: targetDefault}},
		}, {
			Target: targetOtherPath,
			Mocks:  []*Mock{{Target: targetOtherPath}},
		}},
	},
}

func TestNewFiles(t *testing.T) {
	test.Map(t, testNewFilesParams).
		Run(func(t test.Test, param NewFilesParams) {
			// Given
			mocks := clone.Clone(param.mocks).([]*Mock)

			// When
			files := NewFiles(mocks)

			// Then
			assert.Equal(t, param.expectFiles, files)
		})
}

var (
	// Test directory.
	testDirModels = func() string {
		dir, err := os.MkdirTemp("", "go-testing-*")
		if err != nil {
			panic(err)
		}
		return dir
	}()

	targetStdout = &Type{Package: pkgMock, Path: pathMock, File: "-"}
	targetCustom = &Type{
		Package: pkgMock, Path: pathMock,
		File: filepath.Join(testDirModels, fileMock),
	}
	fileStdout   = &File{Target: targetStdout}
	fileCustom   = &File{Target: targetCustom}
	fileNoTarget = &File{}
)

type FileParams struct {
	file        *File
	mocks       []*Mock
	error       error
	expectName  string
	expectOpen  error
	expectWrite error
	expectClose error
}

var testFileParams = map[string]FileParams{
	"file stdout": {
		file: fileStdout,
		mocks: []*Mock{{
			Source: sourceIFaceAny,
			Target: targetStdout,
		}},
		expectName: os.Stdout.Name(),
	},

	"file custom": {
		file: fileCustom,
		mocks: []*Mock{{
			Source: sourceIFaceAny,
			Target: targetCustom,
		}},
		expectName: filepath.Join(testDirModels, fileMock),
	},

	"file error": {
		file: fileStdout,
		mocks: []*Mock{{
			Source: sourceIFaceAny,
			Target: targetStdout,
		}},
		expectName:  os.Stdout.Name(),
		error:       errAny,
		expectWrite: errAny,
	},

	"no such directory": {
		file: fileNoTarget,
		mocks: []*Mock{{
			Source: sourceIFaceAny,
		}},
		expectOpen: func(path string) error {
			return NewErrFileOpening(path, &fs.PathError{
				Op: "open", Path: path, Err: error(syscall.ENOENT),
			})
		}(""),
		expectWrite: errors.New("invalid argument"),
		expectClose: errors.New("invalid argument"),
	},
}

func TestFile(t *testing.T) {
	test.Map(t, testFileParams).
		Run(func(t test.Test, param FileParams) {
			// Given
			fmocks := clone.Clone(param.mocks).([]*Mock)
			files := NewFiles(fmocks, ImportsTemplate...)
			require.Len(t, files, 1)
			file := files[0]

			// When
			err := file.Open(os.Stdout)

			// Then
			assert.Equal(t, param.expectOpen, err)
			if param.expectName != "" {
				assert.Equal(t, param.expectName, file.Writer.Name())
			} else {
				assert.Nil(t, file.Writer)
			}

			// Given
			mocks := mock.NewMocks(t).Expect(
				Execute(file.Writer, file, param.error))

			// When
			err = file.Write(mock.Get(mocks, NewMockTemplate))

			// Then
			assert.Equal(t, param.expectWrite, err)

			// When
			err = file.Close()

			// Then
			assert.Equal(t, param.expectClose, err)
		}).
		Cleanup(func() {
			os.RemoveAll(testDirModels)
		})
}
