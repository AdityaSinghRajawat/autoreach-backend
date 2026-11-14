package db

import "fmt"

type RowScanner interface {
	Scan(destinations ...any) error
}

func FindOne[T any](
	row RowScanner,
	scan func(row RowScanner) (*T, error),
	operation string,
) (*T, error) {
	record, err := scan(row)
	if err != nil {
		if IsNoRows(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	return record, nil
}
