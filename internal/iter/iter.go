// Package iter contains a collection of helpful generic functions for working
// with iterators that only exist because the standard iterator packages do not
// support these generic functions.
//
// It is not part of the public interface and considered highly instable. In
// the future these functions are hopefully supported by the standard library.
package iter

import (
	"iter"
	"sync"

	"github.com/tkrop/go-testing/test"
)

// SyncMap returns an iter.Seq2 iterator over the entries of a *sync.Map,
// casting keys and values to the given type parameters K and V respectively.
func SyncMap[K comparable, V any](source *sync.Map) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		source.Range(func(k, v any) bool {
			return yield(test.Cast[K](k), test.Cast[V](v))
		})
	}
}
