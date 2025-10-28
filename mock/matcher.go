package mock

import (
	"fmt"
	"reflect"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/pmezard/go-difflib/difflib"
	"go.uber.org/mock/gomock"
)

const (
	// DefaultContext provides the default number of context lines to show
	// before and after the changes in a diff.
	DefaultContext = 3
	// DefaultSkippingSize is the maximum size of the string representation of a value
	// presented in the output.
	DefaultSkippingSize = 50 // bufio.MaxScanTokenSize - 100
	// DefaultSkippingTail is the size of the tail after the skipped value part.
	DefaultSkippingTail = 5
)

// MatcherConfig holds configuration settings for matchers.
type MatcherConfig struct {
	// Size of string representation before skipping part of it in output.
	skippingSize int
	// Tail size after the skipped part string representation.
	skippingTail int

	// Internal diff lib settings.
	dlib *difflib.UnifiedDiff
	// Internal spew config settings.
	spew *spew.ConfigState
	// Internal spew config settings (disabled methods for time).
	spewTime *spew.ConfigState
}

// NewMatcherConfig creates a new matcher configuration instance with default
// values.
func NewMatcherConfig() *MatcherConfig {
	return &MatcherConfig{
		skippingSize: DefaultSkippingSize,
		skippingTail: DefaultSkippingTail,
		dlib: &difflib.UnifiedDiff{
			Context:  DefaultContext,
			FromFile: "Want", FromDate: "",
			ToFile: "Got", ToDate: "",
		},
		spew: &spew.ConfigState{
			Indent:                  "  ",
			DisableMethods:          true,
			DisablePointerAddresses: true,
			DisableCapacities:       true,
			SortKeys:                true,
		},
		spewTime: &spew.ConfigState{
			Indent:                  "  ",
			DisablePointerAddresses: true,
			DisableCapacities:       true,
			SortKeys:                true,
		},
	}
}

// Context specifies the number of context lines to show before and after
// changes in a diff. The default, 3, means no context lines.
func Context(context int) ConfigFunc {
	return func(mocks *Mocks) {
		mocks.matcher.dlib.Context = context
	}
}

// FromFile specifies the label to use for the "from" side of the diff.
// Default is "Want".
func FromFile(file string) ConfigFunc {
	return func(mocks *Mocks) {
		mocks.matcher.dlib.FromFile = file
	}
}

// FromDate specifies the label to use for the "from" date of the diff.
// Default is empty.
func FromDate(date string) ConfigFunc {
	return func(mocks *Mocks) {
		mocks.matcher.dlib.FromDate = date
	}
}

// ToFile specifies the label to use for the "to" side of the diff. Default is
// "Got".
func ToFile(file string) ConfigFunc {
	return func(mocks *Mocks) {
		mocks.matcher.dlib.ToFile = file
	}
}

// ToDate specifies the label to use for the "to" date of the diff. Default is
// empty.
func ToDate(date string) ConfigFunc {
	return func(mocks *Mocks) {
		mocks.matcher.dlib.ToDate = date
	}
}

// Indent specifies the string to use for each indentation level. The global
// config instance that all top-level functions use set this to a single space
// by default. If you would like more indentation, you might set this to a tab
// with "\t" or perhaps two spaces with "  ".
func Indent(indent string) ConfigFunc {
	return func(mocks *Mocks) {
		mocks.matcher.spewTime.Indent = indent
		mocks.matcher.spew.Indent = indent
	}
}

// MaxDepth controls the maximum number of levels to descend into nested data
// structures. The default 0 means there is no limit. Circular data structures
// are properly detected, so it is not necessary to set this value unless you
// specifically want to limit deeply nested structures.
func MaxDepth(maxDepth int) ConfigFunc {
	return func(mocks *Mocks) {
		mocks.matcher.spewTime.MaxDepth = maxDepth
		mocks.matcher.spew.MaxDepth = maxDepth
	}
}

// DisableMethods specifies whether or not error and Stringer interfaces are
// invoked for types that implement them. Default is true, meaning that these
// methods will not be invoked.
func DisableMethods(disable bool) ConfigFunc {
	return func(mocks *Mocks) {
		mocks.matcher.spewTime.DisableMethods = true
		mocks.matcher.spew.DisableMethods = disable
	}
}

// DisablePointerMethods specifies whether or not to check for and invoke error
// and Stringer interfaces on types which only accept a pointer receiver when
// the current type is not a pointer.
//
// *Note:* This might be an unsafe action since calling one a pointer receiver
// could technically mutate the value. In practice, types which choose to
// satisfy an error or Stringer interface with a pointer receiver should not
// mutate their state inside these methods. As a result, this option relies on
// access to the unsafe package, so it will not have any effect when running in
// environments without access to the unsafe package such as Google App Engine
// or with the "safe" build tag specified.
func DisablePointerMethods(disable bool) ConfigFunc {
	return func(mocks *Mocks) {
		mocks.matcher.spewTime.DisablePointerMethods = disable
		mocks.matcher.spew.DisablePointerMethods = disable
	}
}

// DisablePointerAddresses specifies whether to disable the printing of pointer
// addresses. This is useful when diffing data structures in tests.
func DisablePointerAddresses(disable bool) ConfigFunc {
	return func(mocks *Mocks) {
		mocks.matcher.spewTime.DisablePointerAddresses = disable
		mocks.matcher.spew.DisablePointerAddresses = disable
	}
}

// DisableCapacities specifies whether to disable the printing of capacities
// for arrays, slices, maps and channels. This is useful when diffing data
// structures in tests.
func DisableCapacities(disable bool) ConfigFunc {
	return func(mocks *Mocks) {
		mocks.matcher.spewTime.DisableCapacities = disable
		mocks.matcher.spew.DisableCapacities = disable
	}
}

// ContinueOnMethod specifies whether or not recursion should continue once a
// custom error or Stringer interface is invoked.  The default, false, means it
// will print the results of invoking the custom error or Stringer interface
// and return immediately instead of continuing to recurse into the internals
// of the data type.
//
// *Note:* This flag does not have any effect if method invocation is disabled
// via the DisableMethods or DisablePointerMethods options.
func ContinueOnMethod(enable bool) ConfigFunc {
	return func(mocks *Mocks) {
		mocks.matcher.spewTime.ContinueOnMethod = enable
		mocks.matcher.spew.ContinueOnMethod = enable
	}
}

// SortKeys specifies map keys should be sorted before being printed. Use this
// to have a more deterministic, diffable output.  Note that only native types
// (bool, int, uint, floats, uintptr and string) and types that support the
// error or Stringer interfaces (if methods are enabled) are supported, with
// other types sorted according to the reflect.Value.String() output which
// guarantees display stability.
func SortKeys(sort bool) ConfigFunc {
	return func(mocks *Mocks) {
		mocks.matcher.spewTime.SortKeys = sort
		mocks.matcher.spew.SortKeys = sort
	}
}

// SpewKeys specifies that, as a last resort attempt, map keys should be spewed
// to strings and sorted by those strings.  This is only considered if SortKeys
// is true.
func SpewKeys(spew bool) ConfigFunc {
	return func(mocks *Mocks) {
		mocks.matcher.spewTime.SpewKeys = spew
		mocks.matcher.spew.SpewKeys = spew
	}
}

// Equal is an improved `gomock.Matcher` that matches via `reflect.DeepEqual`
// showing detailed diff when there is a mismatch.
type Equal struct {
	config *MatcherConfig
	want   any
	diff   string
}

// Equal returns an improved equals matcher showing a detailed diff when there
// is a mismatch in the expected and actual values.
func (mocks *Mocks) Equal(want any) *Equal {
	return &Equal{
		config: mocks.matcher,
		want:   want,
		diff:   "",
	}
}

// Matches returns whether the actual value is equal to the expected value.
func (eq *Equal) Matches(got any) bool {
	if !gomock.Eq(eq.want).Matches(got) {
		eq.diff = eq.Diff(eq.want, got)
		return false
	}
	return true
}

// Got returns a string representation of the actual value.
func (eq *Equal) Got(got any) string {
	return fmt.Sprintf("%T(%s)", got, eq.skip(got))
}

// String returns a string representation of the expected value along with the
// diff between expected and actual values as long as both are of the same type
// and are a struct, map, slice, array or string. Otherwise the diff is hidden.
func (eq *Equal) String() string {
	if eq.diff != "" {
		return fmt.Sprintf("%T(%s)\nDiff (-want, +got):\n%s",
			eq.want, eq.skip(eq.want), eq.diff)
	}
	return fmt.Sprintf("%T(%s)", eq.want, eq.skip(eq.want))
}

// skip returns a truncated string representation of the given value.
func (eq *Equal) skip(v any) string {
	config, value := eq.config, fmt.Sprintf("%#v", v)
	if len(value) > config.skippingSize {
		return value[0:config.skippingSize-config.skippingTail] +
			"<... skipped ...>" +
			value[len(value)-config.skippingTail:]
	}
	return value
}

// Diff returns a diff of the expected value and the actual value as long as
// both are of the same type and are a struct, map, slice, array or string.
// Otherwise it returns an empty string.
func (eq *Equal) Diff(want, got any) string {
	if want == nil || got == nil {
		return ""
	}

	etype := reflect.TypeOf(want)
	atype := reflect.TypeOf(got)

	if etype != atype {
		return ""
	}

	ekind := etype.Kind()
	if ekind == reflect.Ptr {
		ekind = etype.Elem().Kind()
	}
	if ekind != reflect.Struct && ekind != reflect.Map &&
		ekind != reflect.Slice && ekind != reflect.Array &&
		ekind != reflect.String {
		return ""
	}

	var estr, astr string

	config := eq.config
	switch etype {
	case reflect.TypeOf(""):
		estr = reflect.ValueOf(want).String()
		astr = reflect.ValueOf(got).String()
	case reflect.TypeOf(time.Time{}):
		estr = config.spewTime.Sdump(want)
		astr = config.spewTime.Sdump(got)
	default:
		estr = config.spew.Sdump(want)
		astr = config.spew.Sdump(got)
	}

	dlib := config.dlib
	diff, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A: difflib.SplitLines(estr), B: difflib.SplitLines(astr),
		FromFile: dlib.FromFile, FromDate: dlib.FromDate,
		ToFile: dlib.ToFile, ToDate: dlib.ToDate,
		Context: dlib.Context,
	})

	return diff
}
