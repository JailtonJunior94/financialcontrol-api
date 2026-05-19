package mssql_test

import (
	"errors"
	"testing"

	mssqldrv "github.com/microsoft/go-mssqldb"
	"github.com/stretchr/testify/assert"

	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/infrastructure/persistence/mssql"
)

func TestMapDriverError(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("generic error")

	cases := []struct {
		name    string
		input   error
		wantErr error
	}{
		{
			name:    "nil passthrough",
			input:   nil,
			wantErr: nil,
		},
		{
			name:    "active-refund unique index 2601 maps to ErrRefundAlreadyExists",
			input:   mssqldrv.Error{Number: 2601, Message: "Cannot insert duplicate key row ... index 'UX_Tx_ActiveRefundPerOriginal'."},
			wantErr: domain.ErrRefundAlreadyExists,
		},
		{
			name:    "unique constraint 2627 on active-refund index maps to ErrRefundAlreadyExists",
			input:   mssqldrv.Error{Number: 2627, Message: "Violation of UNIQUE KEY constraint 'UX_Tx_ActiveRefundPerOriginal'."},
			wantErr: domain.ErrRefundAlreadyExists,
		},
		{
			name:    "unique violation on a different index passes through",
			input:   mssqldrv.Error{Number: 2601, Message: "Cannot insert duplicate key row ... index 'UX_Other'."},
			wantErr: mssqldrv.Error{Number: 2601, Message: "Cannot insert duplicate key row ... index 'UX_Other'."},
		},
		{
			name:    "other mssql error passes through",
			input:   mssqldrv.Error{Number: 515},
			wantErr: mssqldrv.Error{Number: 515},
		},
		{
			name:    "non-mssql error passes through",
			input:   sentinel,
			wantErr: sentinel,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := mssql.MapDriverError(tc.input)
			assert.Equal(t, tc.wantErr, got)
		})
	}
}
