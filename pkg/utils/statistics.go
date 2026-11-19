package utils

import (
	"math"
	"slices"
)

func Percentile(values []float64, fraction float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sorted := slices.Clone(values)
	slices.Sort(sorted)

	position := max(0, min(1, fraction)) * float64(len(sorted)-1)
	lower := int(math.Floor(position))
	upper := int(math.Ceil(position))

	return sorted[lower] + (sorted[upper]-sorted[lower])*(position-float64(lower))
}

func CosineSimilarity(left, right []float32) float64 {
	if len(left) != len(right) || len(left) == 0 {
		return 0
	}

	var dot, leftNorm, rightNorm float64

	for index := range left {
		dot += float64(left[index]) * float64(right[index])
		leftNorm += float64(left[index]) * float64(left[index])
		rightNorm += float64(right[index]) * float64(right[index])
	}

	if leftNorm == 0 || rightNorm == 0 {
		return 0
	}

	return dot / (math.Sqrt(leftNorm) * math.Sqrt(rightNorm))
}
