package reflect

import (
	"fmt"
	"reflect"
)

// StringArgs is a helper function that converts a slice of any to a slice of
// strings. The function is used only by the reporter package to format the
// function arguments in `UnexpectedCall` and `ConsumedCall` in the same way
// as the controller is doing it.
//
// Unfortunately, this requires us to duplicate the internal controller logic
// here creating a brittle dependency on the implementation.
func StringArgs(args []any) []string {
	sargs := make([]string, len(args))
	for i, arg := range args {
		sargs[i] = getString(arg)
	}
	return sargs
}

// method, which avoids potential deadlocks.
func getString(x any) string {
	if isGeneratedMock(x) {
		return fmt.Sprintf("%T", x)
	}
	if s, ok := x.(fmt.Stringer); ok {
		return s.String()
	}
	return fmt.Sprintf("%v", x)
}

// isGeneratedMock checks if the given type has a "isgomock" field, indicating
// it is a generated mock.
func isGeneratedMock(x any) bool {
	typ := reflect.TypeOf(x)
	if typ == nil {
		return false
	}
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return false
	}
	_, isgomock := typ.FieldByName("isgomock")
	return isgomock
}
