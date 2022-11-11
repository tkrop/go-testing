package slices

// Reverse reverses the given slice.
func Reverse[T any](slice []T) []T {
	for i, j := 0, len(slice)-1; i < j; i, j = i+1, j-1 {
		slice[i], slice[j] = slice[j], slice[i]
	}
	return slice
}

// Permute permutates the given slice starting at the position given by the
// index and call the `do` function on each permutation to collect the result.
// For a full permutation the `index` must start with `0`.
func Permute[T any](slice []T, do func([]T), i int) {
	if i <= len(slice) {
		Permute(slice, do, i+1)
		for j := i + 1; j < len(slice); j++ {
			slice[i], slice[j] = slice[j], slice[i]
			Permute(slice, do, i+1)
			slice[i], slice[j] = slice[j], slice[i]
		}
	} else {
		do(slice)
	}
}
