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
	expectIFaceStub = string(test.Must(os.ReadFile("mock_stub_test.gox")))
	// Generated IFace mock with methods.
	expectIFace = string(test.Must(os.ReadFile("mock_iface_test.gox")))
)

type TemplateParams struct {
	mocks  []*Mock
	expect string
}

var testTemplateParams = map[string]TemplateParams{
	"iface no methods": {
		mocks: []*Mock{{
			Source: sourceIFaceAny,
			Target: targetMockTestIFace.With(&Type{File: "-"}),
		}},
		expect: expectIFaceStub,
	},
	"iface with methods": {
		mocks: []*Mock{{
			Source:  sourceIFaceAny,
			Target:  targetMockTestIFace.With(&Type{File: "-"}),
			Methods: methodsLoadIFace,
		}},
		expect: expectIFace,
	},
}

func TestTemplate(t *testing.T) {
	temp, imports, err := NewTemplate()
	require.NoError(t, err)

	test.Map(t, testTemplateParams).
		Run(func(t test.Test, param TemplateParams) {
			// Given
			imports := clone.Clone(imports).([]*Import)
			mocks := clone.Clone(param.mocks).([]*Mock)
			files := NewFiles(mocks, imports...)
			require.Len(t, files, 1)

			file := files[0]
			writer := &bytes.Buffer{}

			// When
			err := temp.Execute(writer, file)

			// Then
			assert.NoError(t, err)
			assert.Equal(t, param.expect, writer.String())
		})
}
