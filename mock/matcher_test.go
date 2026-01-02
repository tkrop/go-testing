package mock_test

import (
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/pmezard/go-difflib/difflib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tkrop/go-testing/mock"
	"github.com/tkrop/go-testing/reflect"
	"github.com/tkrop/go-testing/test"
)

const (
	nameDlib     = "dlib"
	nameSpew     = "spew"
	nameSpewTime = "spewTime"
	nameMatcher  = "matcher"
)

// diff creates the complete expected diff output in unified diff format. It
// takes a hunk header (e.g., "-1 +1" or "-1,3 +1,3") and the diff content, and
// constructs the full diff with standard headers and trailing space+newline.
func diff(hunk, content string) string {
	return "--- Want\n+++ Got\n@@ -" + hunk + " @@\n" + content
}

// GetDiffConfigAccessor returns a builder to access the matcher config of
// the given mocks.
func GetDiffConfigAccessor(
	mocks *mock.Mocks,
) reflect.Builder[*mock.DiffConfig] {
	return reflect.NewAccessor(reflect.NewAccessor(mocks).
		Get(nameMatcher).(*mock.DiffConfig))
}

type DiffParams struct {
	want any
	got  any
	diff string
}

var diffTestCases = map[string]DiffParams{
	// Nil cases.
	"nil want": {
		want: nil,
		got:  "something",
	},
	"nil got": {
		want: "something",
		got:  nil,
	},
	"both nil": {
		want: nil,
		got:  nil,
	},

	// Different type cases.
	"different types string vs int": {
		want: "hello",
		got:  42,
	},
	"different types struct vs map": {
		want: struct{ A int }{1},
		got:  map[string]int{"a": 1},
	},

	// Unsupported type cases.
	"int type": {
		want: 42,
		got:  43,
	},
	"float type": {
		want: 3.14,
		got:  2.71,
	},
	"bool type": {
		want: true,
		got:  false,
	},

	// String type cases.
	"string equal": {
		want: "hello",
		got:  "hello",
	},
	"string different": {
		want: "hello",
		got:  "world",
		diff: diff("1 +1",
			"-hello\n"+"+world\n"),
	},
	"string multiline": {
		want: "line1\nline2\nline3",
		got:  "line1\nmodified\nline3",
		diff: diff("1,3 +1,3",
			" line1\n"+
				"-line2\n"+
				"+modified\n"+
				" line3\n"),
	},

	// Struct type cases.
	"struct equal": {
		want: struct{ A int }{1},
		got:  struct{ A int }{1},
	},
	"struct different": {
		want: struct{ A int }{1},
		got:  struct{ A int }{2},
		diff: diff("1,4 +1,4",
			" (struct { A int }) {\n"+
				"-  A: (int) 1\n"+
				"+  A: (int) 2\n"+
				" }\n \n"),
	},
	"struct complex": {
		want: struct {
			Name string
			Age  int
		}{"Alice", 30},
		got: struct {
			Name string
			Age  int
		}{"Bob", 25},
		diff: diff("1,5 +1,5",
			" (struct { Name string; Age int }) {\n"+
				"-  Name: (string) (len=5) \"Alice\",\n"+
				"-  Age: (int) 30\n"+
				"+  Name: (string) (len=3) \"Bob\",\n"+
				"+  Age: (int) 25\n"+
				" }\n \n"),
	},

	// Map type cases.
	"map equal": {
		want: map[string]int{"a": 1},
		got:  map[string]int{"a": 1},
	},
	"map different values": {
		want: map[string]int{"a": 1, "b": 2},
		got:  map[string]int{"a": 1, "b": 3},
		diff: diff("1,5 +1,5",
			" (map[string]int) (len=2) {\n"+
				"   (string) (len=1) \"a\": (int) 1,\n"+
				"-  (string) (len=1) \"b\": (int) 2\n"+
				"+  (string) (len=1) \"b\": (int) 3\n"+
				" }\n \n"),
	},
	"map different keys": {
		want: map[string]int{"a": 1},
		got:  map[string]int{"b": 1},
		diff: diff("1,4 +1,4",
			" (map[string]int) (len=1) {\n"+
				"-  (string) (len=1) \"a\": (int) 1\n"+
				"+  (string) (len=1) \"b\": (int) 1\n"+
				" }\n \n"),
	},

	// Slice type cases.
	"slice equal": {
		want: []int{1, 2, 3},
		got:  []int{1, 2, 3},
	},
	"slice different": {
		want: []int{1, 2, 3},
		got:  []int{1, 4, 3},
		diff: diff("1,6 +1,6",
			" ([]int) (len=3) {\n"+
				"   (int) 1,\n"+
				"-  (int) 2,\n"+
				"+  (int) 4,\n"+
				"   (int) 3\n"+
				" }\n \n"),
	},
	"slice different length": {
		want: []int{1, 2, 3},
		got:  []int{1, 2},
		diff: diff("1,6 +1,5",
			"-([]int) (len=3) {\n"+
				"+([]int) (len=2) {\n"+
				"   (int) 1,\n"+
				"-  (int) 2,\n"+
				"-  (int) 3\n"+
				"+  (int) 2\n"+
				" }\n \n"),
	},
	"slice of strings": {
		want: []string{"a", "b", "c"},
		got:  []string{"a", "x", "c"},
		diff: diff("1,6 +1,6",
			" ([]string) (len=3) {\n"+
				"   (string) (len=1) \"a\",\n"+
				"-  (string) (len=1) \"b\",\n"+
				"+  (string) (len=1) \"x\",\n"+
				"   (string) (len=1) \"c\"\n"+
				" }\n \n"),
	},

	// Array type cases.
	"array equal": {
		want: [3]int{1, 2, 3},
		got:  [3]int{1, 2, 3},
	},
	"array different": {
		want: [3]int{1, 2, 3},
		got:  [3]int{1, 4, 3},
		diff: diff("1,6 +1,6",
			" ([3]int) (len=3) {\n"+
				"   (int) 1,\n"+
				"-  (int) 2,\n"+
				"+  (int) 4,\n"+
				"   (int) 3\n"+
				" }\n \n"),
	},

	// Pointer to supported type cases.
	"pointer to struct": {
		want: &struct{ A int }{1},
		got:  &struct{ A int }{2},
		diff: diff("1,4 +1,4",
			" (*struct { A int })({\n"+
				"-  A: (int) 1\n"+
				"+  A: (int) 2\n"+
				" })\n \n"),
	},

	// Time type cases.
	"time equal": {
		want: time.Date(2025, 10, 27, 12, 0, 0, 0, time.UTC),
		got:  time.Date(2025, 10, 27, 12, 0, 0, 0, time.UTC),
	},
	"time different": {
		want: time.Date(2025, 10, 27, 12, 0, 0, 0, time.UTC),
		got:  time.Date(2025, 10, 28, 12, 0, 0, 0, time.UTC),
		diff: diff("1,2 +1,2",
			"-(time.Time) 2025-10-27 12:00:00 +0000 UTC\n"+
				"+(time.Time) 2025-10-28 12:00:00 +0000 UTC\n \n"),
	},
}

func TestDiff(t *testing.T) {
	test.Map(t, diffTestCases).
		Run(func(t test.Test, param DiffParams) {
			// Given
			config := mock.NewDiffConfig()

			// When
			diff := config.Diff(param.want, param.got)

			// Then
			assert.Equal(t, param.diff, diff)
		})
}

type ConfigParams struct {
	config mock.ConfigFunc
	access func(mocks *mock.Mocks) any
	expect any
}

var configTestCases = map[string]ConfigParams{
	// Diff config options.
	"diff context": {
		config: mock.Context(7),
		access: func(mocks *mock.Mocks) any {
			return GetDiffConfigAccessor(mocks).
				Get(nameDlib).(*difflib.UnifiedDiff).Context
		},
		expect: 7,
	},
	"diff from-file": {
		config: mock.FromFile("expect"),
		access: func(mocks *mock.Mocks) any {
			return GetDiffConfigAccessor(mocks).
				Get(nameDlib).(*difflib.UnifiedDiff).FromFile
		},
		expect: "expect",
	},
	"diff from-date": {
		config: mock.FromDate("2025-10-27"),
		access: func(mocks *mock.Mocks) any {
			return GetDiffConfigAccessor(mocks).
				Get(nameDlib).(*difflib.UnifiedDiff).FromDate
		},
		expect: "2025-10-27",
	},
	"diff to-file": {
		config: mock.ToFile("actual.txt"),
		access: func(mocks *mock.Mocks) any {
			return GetDiffConfigAccessor(mocks).
				Get(nameDlib).(*difflib.UnifiedDiff).ToFile
		},
		expect: "actual.txt",
	},
	"diff to-date": {
		config: mock.ToDate("2025-10-28"),
		access: func(mocks *mock.Mocks) any {
			return GetDiffConfigAccessor(mocks).
				Get(nameDlib).(*difflib.UnifiedDiff).ToDate
		},
		expect: "2025-10-28",
	},

	// Spew config options.
	"spew indent": {
		config: mock.Indent("\t"),
		access: func(mocks *mock.Mocks) any {
			return GetDiffConfigAccessor(mocks).
				Get(nameSpew).(*spew.ConfigState).Indent
		},
		expect: "\t",
	},
	"spew max-depth": {
		config: mock.MaxDepth(5),
		access: func(mocks *mock.Mocks) any {
			return GetDiffConfigAccessor(mocks).
				Get(nameSpew).(*spew.ConfigState).MaxDepth
		},
		expect: 5,
	},
	"spew disable-methods": {
		config: mock.DisableMethods(false),
		access: func(mocks *mock.Mocks) any {
			return GetDiffConfigAccessor(mocks).
				Get(nameSpew).(*spew.ConfigState).DisableMethods
		},
		expect: false,
	},
	"spew disable-pointer-methods": {
		config: mock.DisablePointerMethods(true),
		access: func(mocks *mock.Mocks) any {
			return GetDiffConfigAccessor(mocks).
				Get(nameSpew).(*spew.ConfigState).DisablePointerMethods
		},
		expect: true,
	},
	"spew disable-pointer-addresses": {
		config: mock.DisablePointerAddresses(false),
		access: func(mocks *mock.Mocks) any {
			return GetDiffConfigAccessor(mocks).
				Get(nameSpew).(*spew.ConfigState).DisablePointerAddresses
		},
		expect: false,
	},
	"spew disable-capacities": {
		config: mock.DisableCapacities(false),
		access: func(mocks *mock.Mocks) any {
			return GetDiffConfigAccessor(mocks).
				Get(nameSpew).(*spew.ConfigState).DisableCapacities
		},
		expect: false,
	},
	"spew continue-on-method": {
		config: mock.ContinueOnMethod(true),
		access: func(mocks *mock.Mocks) any {
			return GetDiffConfigAccessor(mocks).
				Get(nameSpew).(*spew.ConfigState).ContinueOnMethod
		},
		expect: true,
	},
	"spew sort-keys": {
		config: mock.SortKeys(false),
		access: func(mocks *mock.Mocks) any {
			return GetDiffConfigAccessor(mocks).
				Get(nameSpew).(*spew.ConfigState).SortKeys
		},
		expect: false,
	},
	"spew spew-keys": {
		config: mock.SpewKeys(true),
		access: func(mocks *mock.Mocks) any {
			return GetDiffConfigAccessor(mocks).
				Get(nameSpew).(*spew.ConfigState).SpewKeys
		},
		expect: true,
	},

	// Spew config options.
	"spew-time indent": {
		config: mock.Indent("\t"),
		access: func(mocks *mock.Mocks) any {
			return GetDiffConfigAccessor(mocks).
				Get(nameSpewTime).(*spew.ConfigState).Indent
		},
		expect: "\t",
	},
	"spew-time max-depth": {
		config: mock.MaxDepth(5),
		access: func(mocks *mock.Mocks) any {
			return GetDiffConfigAccessor(mocks).
				Get(nameSpewTime).(*spew.ConfigState).MaxDepth
		},
		expect: 5,
	},
	"spew-time disable-methods": {
		config: mock.DisableMethods(false),
		access: func(mocks *mock.Mocks) any {
			return GetDiffConfigAccessor(mocks).
				Get(nameSpewTime).(*spew.ConfigState).DisableMethods
		},
		expect: true, // exception: default is true for spewtime.
	},
	"spew-time disable-pointer-methods": {
		config: mock.DisablePointerMethods(true),
		access: func(mocks *mock.Mocks) any {
			return GetDiffConfigAccessor(mocks).
				Get(nameSpewTime).(*spew.ConfigState).DisablePointerMethods
		},
		expect: true,
	},
	"spew-time disable-pointer-addresses": {
		config: mock.DisablePointerAddresses(false),
		access: func(mocks *mock.Mocks) any {
			return GetDiffConfigAccessor(mocks).
				Get(nameSpewTime).(*spew.ConfigState).DisablePointerAddresses
		},
		expect: false,
	},
	"spew-time disable-capacities": {
		config: mock.DisableCapacities(false),
		access: func(mocks *mock.Mocks) any {
			return GetDiffConfigAccessor(mocks).
				Get(nameSpewTime).(*spew.ConfigState).DisableCapacities
		},
		expect: false,
	},
	"spew-time continue-on-method": {
		config: mock.ContinueOnMethod(true),
		access: func(mocks *mock.Mocks) any {
			return GetDiffConfigAccessor(mocks).
				Get(nameSpewTime).(*spew.ConfigState).ContinueOnMethod
		},
		expect: true,
	},
	"spew-time sort-keys": {
		config: mock.SortKeys(false),
		access: func(mocks *mock.Mocks) any {
			return GetDiffConfigAccessor(mocks).
				Get(nameSpewTime).(*spew.ConfigState).SortKeys
		},
		expect: false,
	},
	"spew-time spew-keys": {
		config: mock.SpewKeys(true),
		access: func(mocks *mock.Mocks) any {
			return GetDiffConfigAccessor(mocks).
				Get(nameSpewTime).(*spew.ConfigState).SpewKeys
		},
		expect: true,
	},
}

func TestMatcherConfig(t *testing.T) {
	test.Map(t, configTestCases).
		Run(func(t test.Test, param ConfigParams) {
			// Given
			mocks := mock.NewMocks(t)
			require.NotEqual(t, param.expect, param.access(mocks))

			// When
			mocks.Config(param.config)

			// Then
			assert.Equal(t, param.expect, param.access(mocks))
		})
}

type EqualMatchesParams struct {
	want   any
	got    any
	expect bool
}

var equalMatchesTestCases = map[string]EqualMatchesParams{
	"equal primitives": {
		want:   42,
		got:    42,
		expect: true,
	},
	"different primitives": {
		want:   42,
		got:    43,
		expect: false,
	},
	"equal strings": {
		want:   "hello",
		got:    "hello",
		expect: true,
	},
	"different strings": {
		want:   "hello",
		got:    "world",
		expect: false,
	},
	"equal structs": {
		want:   struct{ A int }{1},
		got:    struct{ A int }{1},
		expect: true,
	},
	"different structs": {
		want:   struct{ A int }{1},
		got:    struct{ A int }{2},
		expect: false,
	},
	"equal slices": {
		want:   []int{1, 2, 3},
		got:    []int{1, 2, 3},
		expect: true,
	},
	"different slices": {
		want:   []int{1, 2, 3},
		got:    []int{1, 4, 3},
		expect: false,
	},
}

func TestEqualMatches(t *testing.T) {
	test.Map(t, equalMatchesTestCases).
		Run(func(t test.Test, param EqualMatchesParams) {
			// Given
			mocks := mock.NewMocks(t)
			matcher := mocks.Equal(param.want)

			// When
			result := matcher.Matches(param.got)

			// Then
			assert.Equal(t, param.expect, result)
		})
}

type EqualGotParams struct {
	value  any
	expect string
}

var equalGotTestCases = map[string]EqualGotParams{
	"int": {
		value:  42,
		expect: "int(42)",
	},
	"string": {
		value:  "hello",
		expect: "string(\"hello\")",
	},
	"struct": {
		value:  struct{ A int }{1},
		expect: "struct { A int }(struct { A int }{A:1})",
	},
	"slice": {
		value:  []int{1, 2, 3},
		expect: "[]int([]int{1, 2, 3})",
	},
}

func TestEqualGot(t *testing.T) {
	test.Map(t, equalGotTestCases).
		Run(func(t test.Test, param EqualGotParams) {
			// Given
			mocks := mock.NewMocks(t)
			matcher := mocks.Equal("dummy")

			// When
			result := matcher.Got(param.value)

			// Then
			assert.Equal(t, param.expect, result)
		})
}

var longValue = func() string {
	longValue := make([]byte, 100000)
	for i := range longValue {
		longValue[i] = 'a'
	}
	return string(longValue)
}()

type EqualStringParams struct {
	want   any
	got    any
	expect string
}

var equalStringTestCases = map[string]EqualStringParams{
	"without diff": {
		want:   "hello",
		got:    nil, // Don't call Matches, so no diff
		expect: "string(\"hello\")",
	},
	"with diff": {
		want: "hello",
		got:  "world",
		expect: "string(\"hello\")\nDiff (-want, +got):\n" +
			diff("1 +1", "-hello\n+world\n"),
	},
	"with long value": {
		want: longValue,
		got:  nil, // Don't call Matches, so no diff.
		expect: "string(\"" +
			longValue[0:mock.DefaultSkippingSize-mock.DefaultSkippingTail-1] +
			"<... skipped ...>" +
			longValue[len(longValue)-mock.DefaultSkippingTail+1:] +
			"\")",
	},
}

func TestEqualString(t *testing.T) {
	test.Map(t, equalStringTestCases).
		Run(func(t test.Test, param EqualStringParams) {
			// Given
			mocks := mock.NewMocks(t)
			matcher := mocks.Equal(param.want)
			matcher.Matches(param.got)

			// When
			result := matcher.String()

			// Then
			assert.Equal(t, param.expect, result)
		})
}
