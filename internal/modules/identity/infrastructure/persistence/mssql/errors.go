package mssql

import (
	"errors"

	mssqldrv "github.com/microsoft/go-mssqldb"

	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain"
)

// SQL Server error numbers raised when a unique index/constraint is violated.
// 2601: Cannot insert duplicate key row in object (unique index).
// 2627: Violation of UNIQUE KEY constraint.
// https://learn.microsoft.com/sql/relational-databases/errors-events/database-engine-events-and-errors
const (
	uniqueIndexViolationNumber      = 2601
	uniqueConstraintViolationNumber = 2627
)

// MapDriverError translates SQL Server driver errors raised by the identity
// repository into domain sentinels. The Users table enforces a UNIQUE Email
// constraint, so a duplicate insert maps to ErrUserAlreadyExists. Returns the
// original error untouched when no mapping applies, preserving the wrapped
// chain for the caller.
func MapDriverError(err error) error {
	var mssqlErr mssqldrv.Error
	if errors.As(err, &mssqlErr) && isUniqueViolation(mssqlErr) {
		return domain.ErrUserAlreadyExists
	}
	return err
}

func isUniqueViolation(err mssqldrv.Error) bool {
	return err.Number == uniqueIndexViolationNumber ||
		err.Number == uniqueConstraintViolationNumber
}
