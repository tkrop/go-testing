# Package testing/perm

Goal of this package is to provide a small framework to simplify tests based
on permutating a test parameter. This is helpful when a test input must be
checked in all known orders with differing outcomes depending on the order.

This was mainly developed to validate the setup functions of the [mock](../mock)
framework, but may be useful in other use cases.


## Standard permutation pattern

To setup a permutation test you need to define two parts:

1. A `TestMap` defined `map[string]func(t *test.TestingT)` setting up the
   tests that are object to permutation and,
2. An `ExpectMap` defined as `map[string]Expect` setting up a set of distinct
   expectation for various permutation.

The permutation in the `ExpectMap` key is given by a list of names using `-`
as separator.

**Example:** If a `TestMap` defines the names `a` and `b`, than the permutation
keys in the `ExpectMap` are written as `a-b` and `b-a`.

In a test, the `TestMap` will usually be defined via a function expecting a
mock handler to permutate the mock call setup:

```go
func SetupPermTestABCD(mocks *mock.Mocks) *perm.Test {
    iface := mock.Get(mocks, mock.NewMockIFace)
    return perm.NewTest(mocks,
        perm.TestMap{
            "a": func(t *test.TestingT) { iface.CallA("a") },
            "b": func(t *test.TestingT) { iface.CallA("b") },
            "c": func(t *test.TestingT) {
                assert.Equal(t, "d", iface.CallB("c"))
            },
            "d": func(t *test.TestingT) {
                assert.Equal(t, "e", iface.CallB("d"))
            },
        })
}
```

As you can see the functions contain the necessary test assertions beside the
mock calls. The `ExpectMap` itself can be incomplete and contain any subset of
the full permutation list with the expected results. E.g.

```go
var testPermParams = perm.ExpectMap{
    "a-b-c-d": test.ExpectSuccess,
    "a-b-d-c": test.ExpectSuccess,
    "a-d-b-c": test.ExpectSuccess,
    "b-a-c-d": test.ExpectSuccess,
    "b-a-d-c": test.ExpectSuccess,
    "b-d-a-c": test.ExpectSuccess,
    "d-a-b-c": test.ExpectSuccess,
    "d-b-a-c": test.ExpectSuccess,
}
```

The nice part of the permutation framework is, that it now allows to complete
the permutation by defining a default value for all remaining permutations by
calling `testPermParams.Remain(test.ExpectSuccess)` fluently. This can be than
used in a parameterized test.

```go
func TestDetach(t *testing.T) {
    for message, expect := range testPermParams.Remain(test.ExpectFailure) {
        t.Run(message, test.Run(expect, func(t *test.TestingT) {
            require.NotEmpty(t, message)

            // Given
            perm := strings.Split(message, "-")
            mockSetup := mock.Chain(
                mock.Detach(mock.None, CallA("a")),
                mock.Detach(mock.Head, CallA("b")),
                mock.Detach(mock.Tail, CallB("c", "d")),
                mock.Detach(mock.Both, CallB("d", "e")),
            )
            mock := MockSetup(t, mockSetup)

            // When
            test := SetupPermTestABCD(mock)

            // Then
            test.Test(t, perm, expect)
        }))
    }
}
```
