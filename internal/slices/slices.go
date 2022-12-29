// Package slices contains a collection of helpful generic functions for
// working with slices. It is currently not part of the public interface and
// must be consider as highly instable.
package slices

// Reverse reverses the given slice.
func Reverse[T any](slice []T) []T {
	for i, j := 0, len(slice)-1; i < j; i, j = i+1, j-1 {
		slice[i], slice[j] = slice[j], slice[i]
	}
	return slice
}

// Permute permutates the given slice as is.
func Permute[T any](slice []T) [][]T {
	perms := [][]T{}
	PermuteDo(slice, func(perm []T) {
		perms = append(perms, Copy(perm))
	}, 0)
	return perms
}

// PermuteDo permutates the given slice starting at the position given by the
// index and call the `do` function on each permutation to collect the result.
// For a full permutation the `index` must start with `0`.
func PermuteDo[T any](slice []T, do func([]T), i int) {
	if i <= len(slice) {
		PermuteDo(slice, do, i+1)
		for j := i + 1; j < len(slice); j++ {
			slice[i], slice[j] = slice[j], slice[i]
			PermuteDo(slice, do, i+1)
			slice[i], slice[j] = slice[j], slice[i]
		}
	} else {
		do(slice)
	}
}

// Copy makes a shallow copy of the given slice.
func Copy[T any](slice []T) []T {
	return append(make([]T, 0, len(slice)), slice...)
}
