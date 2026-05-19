package mssql

import (
	"errors"
	"strings"

	mssqldrv "github.com/microsoft/go-mssqldb"

	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain"
)

// SQL Server error numbers raised when a unique index/constraint is violated.
// 2601: Cannot insert duplicate key row in object (unique index).
// 2627: Violation of UNIQUE KEY constraint.
// https://learn.microsoft.com/sql/relational-databases/errors-events/database-engine-events-and-errors
const (
	uniqueIndexViolationNumber      = 2601
	uniqueConstraintViolationNumber = 2627
)

// activeRefundUniqueIndexName is the filtered unique index that guarantees at
// most one active refund per original transaction (RF-13, 409 duplicate guard).
const activeRefundUniqueIndexName = "UX_Tx_ActiveRefundPerOriginal"

// MapDriverError translates SQL Server driver errors raised by the finance
// repositories into domain sentinels. A violation of the active-refund unique
// index maps to ErrRefundAlreadyExists, closing the concurrent-insert race that
// the application-level guard cannot cover. Returns the original error
// untouched when no mapping applies, preserving the wrapped chain.
func MapDriverError(err error) error {
	var mssqlErr mssqldrv.Error
	if errors.As(err, &mssqlErr) && isActiveRefundUniqueViolation(mssqlErr) {
		return domain.ErrRefundAlreadyExists
	}
	return err
}

func isActiveRefundUniqueViolation(err mssqldrv.Error) bool {
	if err.Number != uniqueIndexViolationNumber && err.Number != uniqueConstraintViolationNumber {
		return false
	}
	return strings.Contains(strings.ToLower(err.Message), strings.ToLower(activeRefundUniqueIndexName))
}
