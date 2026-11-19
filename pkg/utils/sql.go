package utils

import (
	"cmp"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/db"
)

type RowScanner interface {
	Scan(destinations ...any) error
}

// Lists the db-tagged columns of T, qualified by alias, so selects match the model exactly.
func BuildSelectColumns[T any](alias string) string {
	model := reflect.TypeFor[T]()
	columns := make([]string, 0, model.NumField())

	for index := range model.NumField() {
		column := model.Field(index).Tag.Get("db")
		if column != "" && column != "-" {
			columns = append(columns, alias+"."+column)
		}
	}

	return strings.Join(columns, ", ")
}

// Literal, not bound, so partial indexes still match. Pass compile-time constants only.
func BuildSQLEnumList[T ~string](values ...T) string {
	quoted := make([]string, 0, len(values))
	for _, value := range values {
		quoted = append(quoted, "'"+strings.ReplaceAll(string(value), "'", "''")+"'")
	}

	return strings.Join(quoted, ", ")
}

// Writing rows in one key order keeps concurrent batches from deadlocking.
func SortedIndexesByKey[T any](items []T, key func(item T) string) []int {
	order := make([]int, len(items))
	for index := range order {
		order[index] = index
	}

	slices.SortStableFunc(order, func(left, right int) int {
		return cmp.Compare(key(items[left]), key(items[right]))
	})

	return order
}

// A missing row is a nil record, not an error.
func ScanOptionalRow[T any](
	row RowScanner,
	scan func(row RowScanner) (*T, error),
	operation string,
) (*T, error) {
	record, err := scan(row)
	if err != nil {
		if db.IsNoRows(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	return record, nil
}

func FormatVectorLiteral(values []float32) string {
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
