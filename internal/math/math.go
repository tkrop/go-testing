// Package math contains a collection of helpful generic functions helping with
// mathematical problems. It is currently not part of the public interface and
// must be consider as highly instable.
package math

import (
	"golang.org/x/exp/constraints"
)

// Max returns the maximal argument of the given arguments.
func Max[T constraints.Ordered](args ...T) T {
	max := args[0]
	for _, arg := range args[1:] {
		if arg > max {
			max = arg
		}
	}
	return max
}

// Min returns the minimal argument of the given arguments.
func Min[T constraints.Ordered](args ...T) T {
	min := args[0]
	for _, arg := range args[1:] {
		if arg < min {
			min = arg
		}
	}
	return min
}
