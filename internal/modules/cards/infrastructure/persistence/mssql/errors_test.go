package mssql_test

import (
	"errors"
	"testing"

	mssqldrv "github.com/microsoft/go-mssqldb"
	"github.com/stretchr/testify/assert"

	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/infrastructure/persistence/mssql"
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
			name:    "flag foreign key 547 maps to ErrFlagNotFound",
			input:   mssqldrv.Error{Number: 547, Message: "The INSERT statement conflicted with the FOREIGN KEY constraint \"FK_Card_Flag\"."},
			wantErr: domain.ErrFlagNotFound,
		},
		{
			name:    "other foreign key 547 passes through",
			input:   mssqldrv.Error{Number: 547, Message: "The INSERT statement conflicted with the FOREIGN KEY constraint \"FK_Card_User\"."},
			wantErr: mssqldrv.Error{Number: 547, Message: "The INSERT statement conflicted with the FOREIGN KEY constraint \"FK_Card_User\"."},
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
