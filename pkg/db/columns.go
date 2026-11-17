package db

import (
	"reflect"
	"strings"
)

// Columns lists the db-tagged columns of T, qualified by alias, so selects match the model exactly.
func Columns[T any](alias string) string {
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
