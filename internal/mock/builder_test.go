package mock_test

import (
	"testing"

	"github.com/huandu/go-clone"
	"github.com/stretchr/testify/assert"

	. "github.com/tkrop/go-testing/internal/mock"
	test_mock "github.com/tkrop/go-testing/mock"
	"github.com/tkrop/go-testing/test"
)

const (
	aliasMock   = "mock_" + pkgMock
	aliasInt    = "internal_" + aliasMock
	aliasRepo   = "testing_" + aliasInt
	aliasOrg    = "tkrop_" + aliasRepo
	aliasCom    = "com_" + aliasOrg
	aliasGitHub = "github_" + aliasCom
)

var (
	importMock = &Import{Alias: pkgMock, Path: pathMock}
)

func importAlias(alias string) *Import {
	return &Import{Alias: alias, Path: alias}
}

func methodsMockIFaceFunc(alias string) []*Method {
	return []*Method{{
		Name: "Call",
		Params: []*Param{{
			Name: "value", Type: "*" + alias + ".Struct",
		}, {
			Name: "args", Type: "[]*reflect.Value",
		}},
		Results: []*Param{{
			Type: "[]any",
		}},
		Variadic: true,
	}}
}

type FileBuilderParams struct {
	setup      test_mock.SetupFunc
	imports    []*Import
	mocks      []*Mock
	expectFile *File
}

var testFileBuilderParams = map[string]FileBuilderParams{
	"no imports": {
		mocks: []*Mock{{
			Methods: methodsMockIFace,
		}},
		expectFile: &File{
			Imports: []*Import{
				importMock, ImportReflect,
			},
			Mocks: []*Mock{{
				Methods: methodsMockIFaceFunc(pkgMock),
			}},
		},
	},

	"others imported": {
		imports: []*Import{
			ImportReflect, ImportGomock,
		},
		mocks: []*Mock{{
			Methods: methodsMockIFace,
		}},
		expectFile: &File{
			Imports: []*Import{
				ImportReflect, ImportGomock, importMock,
			},
			Mocks: []*Mock{{
				Methods: methodsMockIFaceFunc(pkgMock),
			}},
		},
	},

	"pre imported": {
		imports: []*Import{
			ImportReflect, importMock,
		},
		mocks: []*Mock{{
			Methods: methodsMockIFace,
		}},
		expectFile: &File{
			Imports: []*Import{
				ImportReflect, importMock,
			},
			Mocks: []*Mock{{
				Methods: methodsMockIFaceFunc(pkgMock),
			}},
		},
	},

	"alias imported": {
		imports: []*Import{importAlias(pkgMock)},
		mocks: []*Mock{{
			Methods: methodsMockIFace,
		}},
		expectFile: &File{
			Imports: []*Import{
				importAlias(pkgMock), {
					Alias: aliasMock, Path: pathMock,
				}, ImportReflect,
			},
			Mocks: []*Mock{{
				Methods: methodsMockIFaceFunc(aliasMock),
			}},
		},
	},

	"package imported": {
		imports: []*Import{
			importAlias(pkgMock),
			importAlias(aliasMock),
		},
		mocks: []*Mock{{
			Methods: methodsMockIFace,
		}},
		expectFile: &File{
			Imports: []*Import{
				importAlias(pkgMock),
				importAlias(aliasMock), {
					Alias: aliasInt, Path: pathMock,
				}, ImportReflect,
			},
			Mocks: []*Mock{{
				Methods: methodsMockIFaceFunc(aliasInt),
			}},
		},
	},

	"internal imported": {
		imports: []*Import{
			importAlias(pkgMock),
			importAlias(aliasMock),
		},
		mocks: []*Mock{{
			Methods: methodsMockIFace,
		}},
		expectFile: &File{
			Imports: []*Import{
				importAlias(pkgMock),
				importAlias(aliasMock), {
					Alias: aliasInt, Path: pathMock,
				}, ImportReflect,
			},
			Mocks: []*Mock{{
				Methods: methodsMockIFaceFunc(aliasInt),
			}},
		},
	},

	"organization imported": {
		imports: []*Import{
			importAlias(pkgMock),
			importAlias(aliasMock),
			importAlias(aliasInt),
			importAlias(aliasRepo),
		},
		mocks: []*Mock{{
			Methods: methodsMockIFace,
		}},
		expectFile: &File{
			Imports: []*Import{
				importAlias(pkgMock),
				importAlias(aliasMock),
				importAlias(aliasInt),
				importAlias(aliasRepo), {
					Alias: aliasOrg, Path: pathMock,
				}, ImportReflect,
			},
			Mocks: []*Mock{{
				Methods: methodsMockIFaceFunc(aliasOrg),
			}},
		},
	},

	"import double": {
		setup:   test.Panic(NewErrAliasConflict(pathMock, pkgMock)),
		imports: []*Import{importMock, importMock},
	},

	"import conflict": {
		setup: test.Panic(NewErrAliasConflict(pathMock, aliasGitHub)),
		imports: []*Import{
			importAlias(pkgMock),
			importAlias(aliasMock),
			importAlias(aliasInt),
			importAlias(aliasRepo),
			importAlias(aliasOrg),
			importAlias(aliasCom),
			importAlias(aliasGitHub),
		},
		mocks: []*Mock{{
			Methods: methodsMockIFace,
		}},
	},
}

func TestFileBuilder(t *testing.T) {
	test.Map(t, testFileBuilderParams).Run(func(t test.Test, param FileBuilderParams) {
		test_mock.NewMocks(t).Expect(param.setup)

		// Given
		builder := NewFileBuilder(Type{})
		mocks := clone.Clone(param.mocks).([]*Mock)
		assert.Equal(t, param.mocks, mocks)

		// When
		file := builder.AddImports(param.imports...).
			AddMocks(mocks...).Build()

		// Then
		assert.Equal(t, param.expectFile, file)
	})
}
