package mssql

import (
	"errors"
	"strings"

	mssqldrv "github.com/denisenkom/go-mssqldb"
	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain"
)

// foreignKeyConstraintNumber is the MSSQL error code 547 raised when a
// foreign key constraint is violated (e.g. attempting to insert a Card
// referencing a non-existent Flag). See:
// https://learn.microsoft.com/sql/relational-databases/errors-events/mssqlserver-547-database-engine-error
const foreignKeyConstraintNumber = 547
const cardFlagConstraintName = "FK_Card_Flag"

func MapDriverError(err error) error {
	var mssqlErr mssqldrv.Error
	if errors.As(err, &mssqlErr) && isCardFlagForeignKeyViolation(mssqlErr) {
		return domain.ErrFlagNotFound
	}
	return err
}

func isCardFlagForeignKeyViolation(err mssqldrv.Error) bool {
	if err.Number != foreignKeyConstraintNumber {
		return false
	}
	return strings.Contains(strings.ToLower(err.Message), strings.ToLower(cardFlagConstraintName))
}
