package reflect

import "reflect"

// Run executes a test with given name and test callback function on the given
// target. The target is typically a test runner instance, such as `*testing.T`
// or `*testing.B` and must implement `Run(name string, call func(T))` that is
// called to execute the test.
//
// A panic is raised if the target has no `Run` method or if target type does
// not implement T.
func Run[T any](target any, name string, call func(T)) {
	value := reflect.ValueOf(target)
	if !value.IsValid() {
		panic(name + ": target must not be nil")
	}

	method := value.MethodByName("Run")
	if !method.IsValid() {
		panic(name + ": target does not implement method [" +
			reflect.TypeOf(target).String() + " => Run]")
	}

	// Build a typed func(*concreteT) that wraps the inner argument as T
	// before forwarding to the provided call function.
	wrapper := reflect.MakeFunc(method.Type().In(1),
		func(args []reflect.Value) []reflect.Value {
			iface := args[0].Interface()
			if typ, ok := iface.(T); ok {
				call(typ)
				return nil
			}

			panic(name + ": target does not implement type [" +
				reflect.TypeOf(iface).String() + " => " +
				reflect.TypeOf((*T)(nil)).Elem().String() + "]")
		})

	method.Call([]reflect.Value{reflect.ValueOf(name), wrapper})
}
