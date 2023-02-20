package mock_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/huandu/go-clone"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "github.com/tkrop/go-testing/internal/mock"
	"github.com/tkrop/go-testing/test"
)

var (
	// Generated IFace mock without methods.
	expectIFaceStub = func() string {
		file, err := os.ReadFile("mock_stub_test.gox")
		if err != nil {
			panic(err)
		}
		return string(file)
	}()
	// Generated IFace mock with methods.
	expectIFace = func() string {
		file, err := os.ReadFile("mock_iface_test.gox")
		if err != nil {
			panic(err)
		}
		return string(file)
	}()
)

type GenerateParams struct {
	mocks  []*Mock
	expect string
}

var testGenerateParams = map[string]GenerateParams{
	"iface no methods": {
		mocks: []*Mock{{
			Source: sourceIFaceAny,
			Target: targetMockIFace.With(Type{File: "-"}),
		}},
		expect: expectIFaceStub,
	},
	"iface with methods": {
		mocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetMockIFace.With(Type{File: "-"}),
			Methods: methodsMockIFace,
		}},
		expect: expectIFace,
	},
}

func TestTemplate(t *testing.T) {
	temp, imports, err := NewTemplate()
	require.NoError(t, err)

	test.Map(t, testGenerateParams).
		Run(func(t test.Test, param GenerateParams) {
			// Given
			imports := clone.Clone(imports).([]*Import)
			mocks := clone.Clone(param.mocks).([]*Mock)
			files := NewFiles(mocks, imports...)
			require.Len(t, files, 1)

			file := files[0]
			writer := &bytes.Buffer{}

			// When
			temp.Execute(writer, file)

			// Then
			assert.Equal(t, param.expect, writer.String())
		})
}
