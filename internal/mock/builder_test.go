package mock_test

import (
	"testing"

	"github.com/huandu/go-clone"
	"github.com/stretchr/testify/assert"

	. "github.com/tkrop/go-testing/internal/mock"
	mock "github.com/tkrop/go-testing/mock"
	"github.com/tkrop/go-testing/test"
)

var (
	importMock     = &Import{Path: pathMock}
	importMockTest = &Import{Alias: pkgTest, Path: pathMockTest}
	importTest     = &Import{Alias: pkgTesting, Path: pathTest}
	importIllegal  = &Import{Alias: pkgMock, Path: pathMock}
)

func importAlias(alias string) *Import {
	return &Import{Alias: alias, Path: alias}
}

type FileBuilderParams struct {
	setup      mock.SetupFunc
	target     Type
	imports    []*Import
	mocks      []*Mock
	expectFile *File
}

var testFileBuilderParams = map[string]FileBuilderParams{
	"import double": {
		setup: test.Panic(NewErrAliasConflict(
			importMockTest, pathMockTest)),
		target:  targetMockIFace,
		imports: []*Import{importMockTest, importMockTest},
	},

	"import illegal": {
		setup:   test.Panic(NewErrIllegalImport(importIllegal)),
		target:  targetMockIFace,
		imports: []*Import{importIllegal},
	},

	"no imports": {
		target: targetMockIFace,
		mocks: []*Mock{{
			Methods: methodsMockIFace,
		}},
		expectFile: &File{
			Target: targetMockIFace,
			Imports: []*Import{
				importMockTest, ImportReflect, importTest, importMock,
			},
			Mocks: []*Mock{{
				Methods: methodsMockIFaceFunc(pkgTest, pkgTesting, ""),
			}},
		},
	},

	"others imported": {
		target: targetMockIFace,
		imports: []*Import{
			ImportReflect, ImportGomock, ImportMock,
		},
		mocks: []*Mock{{
			Methods: methodsMockIFace,
		}},
		expectFile: &File{
			Target: targetMockIFace,
			Imports: []*Import{
				ImportReflect, ImportGomock, ImportMock,
				importMockTest, importTest, importMock,
			},
			Mocks: []*Mock{{
				Methods: methodsMockIFaceFunc(pkgTest, pkgTesting, ""),
			}},
		},
	},

	"pre imported": {
		target: targetMockIFace,
		imports: []*Import{
			ImportReflect, importMock, importMockTest, importTest,
		},
		mocks: []*Mock{{
			Methods: methodsMockIFace,
		}},
		expectFile: &File{
			Target: targetMockIFace,
			Imports: []*Import{
				ImportReflect, importMock, importMockTest, importTest,
			},
			Mocks: []*Mock{{
				Methods: methodsMockIFaceFunc(pkgTest, pkgTesting, ""),
			}},
		},
	},

	"alias imported": {
		target:  targetMockIFace,
		imports: []*Import{importAlias(pkgTest)},
		mocks: []*Mock{{
			Methods: methodsMockIFace,
		}},
		expectFile: &File{
			Target: targetMockIFace,
			Imports: []*Import{
				importAlias(pkgTest), {
					Alias: aliasMock, Path: pathMockTest,
				}, ImportReflect, importTest, importMock,
			},
			Mocks: []*Mock{{
				Methods: methodsMockIFaceFunc(aliasMock, pkgTesting, ""),
			}},
		},
	},

	"package imported": {
		target: targetMockIFace,
		imports: []*Import{
			importAlias(pkgTest),
			importAlias(aliasMock),
		},
		mocks: []*Mock{{
			Methods: methodsMockIFace,
		}},
		expectFile: &File{
			Target: targetMockIFace,
			Imports: []*Import{
				importAlias(pkgTest),
				importAlias(aliasMock),
				{
					Alias: aliasInt, Path: pathMockTest,
				},
				ImportReflect, importTest, importMock,
			},
			Mocks: []*Mock{{
				Methods: methodsMockIFaceFunc(aliasInt, pkgTesting, ""),
			}},
		},
	},

	"internal imported": {
		target: targetMockIFace,
		imports: []*Import{
			importAlias(pkgTest),
			importAlias(aliasMock),
		},
		mocks: []*Mock{{
			Methods: methodsMockIFace,
		}},
		expectFile: &File{
			Target: targetMockIFace,
			Imports: []*Import{
				importAlias(pkgTest),
				importAlias(aliasMock),
				{
					Alias: aliasInt, Path: pathMockTest,
				},
				ImportReflect, importTest, importMock,
			},
			Mocks: []*Mock{{
				Methods: methodsMockIFaceFunc(aliasInt, pkgTesting, ""),
			}},
		},
	},

	"organization imported": {
		target: targetMockIFace,
		imports: []*Import{
			importAlias(pkgTest),
			importAlias(aliasMock),
			importAlias(aliasInt),
			importAlias(aliasRepo),
		},
		mocks: []*Mock{{
			Methods: methodsMockIFace,
		}},
		expectFile: &File{
			Target: targetMockIFace,
			Imports: []*Import{
				importAlias(pkgTest),
				importAlias(aliasMock),
				importAlias(aliasInt),
				importAlias(aliasRepo),
				{
					Alias: aliasOrg, Path: pathMockTest,
				},
				ImportReflect, importTest, importMock,
			},
			Mocks: []*Mock{{
				Methods: methodsMockIFaceFunc(aliasOrg, pkgTesting, ""),
			}},
		},
	},

	"import conflict": {
		setup: test.Panic(NewErrAliasConflict(&Import{
			Alias: aliasGitHub, Path: pathMockTest,
		}, pathMockTest)),
		target: targetMockIFace,
		imports: []*Import{
			importAlias(pkgTest),
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

// TODO: var testFileBuilderXParams = map[string]FileBuilderParams{}

func TestFileBuilder(t *testing.T) {
	test.Map(t, testFileBuilderParams).
		Run(func(t test.Test, param FileBuilderParams) {
			mock.NewMocks(t).Expect(param.setup)

			// Given
			builder := NewFileBuilder(param.target)
			mocks := clone.Clone(param.mocks).([]*Mock)
			assert.Equal(t, param.mocks, mocks)

			// When
			file := builder.AddImports(param.imports...).
				AddMocks(mocks...).Build()

			// Then
			assert.Equal(t, param.expectFile, file)
		})
}
