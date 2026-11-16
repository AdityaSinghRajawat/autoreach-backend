package db

import (
	"cmp"
	"slices"
)

// LockOrder returns the positions of items sorted by the key their writes lock. A batch runs in one
// transaction, so writing rows in the same order everywhere keeps two batches from deadlocking.
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
