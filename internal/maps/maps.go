// Package maps contains a collection of helpful generic functions for working
// with maps, that only exist because the standard map packages do not support
// these generic functions.
//
// It is not part of the public interface and considered highly instable. In
// the future these functions are hopefully supported by the standard library.
package maps

// Copy merges the given source maps into the target map, overriding existing
// values if the same key appears in a later source map.
func Copy[K comparable, V any](target map[K]V, sources ...map[K]V) map[K]V {
	for _, source := range sources {
		for k, v := range source {
			target[k] = v
		}
	}
	return target
}
