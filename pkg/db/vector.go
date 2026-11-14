package db

import (
	"strconv"
	"strings"
)

func FormatVector(values []float32) string {
	builder := strings.Builder{}
	builder.Grow(len(values)*10 + 2)
	builder.WriteByte('[')

	for index, value := range values {
		if index > 0 {
			builder.WriteByte(',')
		}

		builder.WriteString(strconv.FormatFloat(float64(value), 'g', -1, 32))
	}

	builder.WriteByte(']')

	return builder.String()
}
