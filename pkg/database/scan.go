package database

import devkitdb "github.com/JailtonJunior94/devkit-go/pkg/database"

// ScanAll iterates rows using scan, closes rows, and propagates rows.Err().
// Returns an empty slice (not nil) for zero rows.
func ScanAll[T any](rows devkitdb.Rows, scan func(devkitdb.Rows) (T, error)) (_ []T, retErr error) {
	defer func() {
		if closeErr := rows.Close(); closeErr != nil && retErr == nil {
			retErr = closeErr
		}
	}()
	result := make([]T, 0)
	for rows.Next() {
		item, err := scan(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
