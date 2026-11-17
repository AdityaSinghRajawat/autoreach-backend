package db

import (
	"cmp"
	"slices"
)

// Writing rows in one key order keeps concurrent batches from deadlocking.
func LockOrder[T any](items []T, key func(item T) string) []int {
	order := make([]int, len(items))
	for index := range order {
		order[index] = index
	}

	slices.SortStableFunc(order, func(left, right int) int {
		return cmp.Compare(key(items[left]), key(items[right]))
	})

	return order
}
