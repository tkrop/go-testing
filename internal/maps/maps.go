// Package maps contains a collection of helpful generic functions for working
// with maps, that only exist because the standard map packages do not support
// these generic functions.
//
// It is not part of the public interface and considered highly instable. In
// the future these functions are hopefully supported by the standard library.
package maps

import (
	"iter"
)

// Copy created a shallow copy of the given map.
func Copy[K comparable, V any](source map[K]V) map[K]V {
	target := make(map[K]V)
	for key, value := range source {
		target[key] = value
	}
	return target
}

// Add the given maps to a common base map overriding existing key values pairs
// added from previous maps if a new entry exists in a latter source map.
func Add[K comparable, V any](target map[K]V, sources ...map[K]V) map[K]V {
	for _, source := range sources {
		for k, v := range source {
			target[k] = v
		}
	}
	return target
}

// Collect collects all entries from the given iterator into a map.
func Collect[K comparable, V any](source iter.Seq2[K, V]) map[K]V {
	target := make(map[K]V)
	for k, v := range source {
		target[k] = v
	}
	return target
}
