package mssql_test

import (
	"context"
	"errors"
	"testing"
	"time"

	devkitmocks "github.com/JailtonJunior94/devkit-go/pkg/database/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/infrastructure/persistence/mssql"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

func newInstallmentRepo(t *testing.T) (*mssql.InstallmentRepository, *devkitmocks.MockDBTX) {
	t.Helper()
	db := devkitmocks.NewMockDBTX(t)
	return mssql.NewInstallmentRepository(db), db
}

func newTestInstallments(t *testing.T, n int) []*entities.Installment {
	t.Helper()
	txID := vos.NewTransactionID()
	invID := vos.NewInvoiceID()
	amount, _ := vos.NewMoney("100.00")
	count, _ := vos.NewInstallmentCount(n)
	now := time.Now().UTC()
	items := make([]*entities.Installment, n)
	for i := 0; i < n; i++ {
		num, _ := vos.NewInstallmentNumber(i + 1)
		items[i] = entities.RehydrateInstallment(
			vos.NewInstallmentID(), txID, invID,
			num, count, amount,
			vos.InstallmentStatusScheduled, nil,
			now, now, nil,
		)
	}
	return items
}

// ---- AddBatch ----

func TestInstallmentRepository_AddBatch_Success(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newInstallmentRepo(t)
	items := newTestInstallments(t, 3)

	result := devkitmocks.NewMockResult(t)
	// ctx + query + 12 named args = 14 total per INSERT
	db.On("ExecContext", anyArgs(14)...).Return(result, nil)

	require.NoError(t, repo.AddBatch(ctx, items))
}

func TestInstallmentRepository_AddBatch_DBError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newInstallmentRepo(t)
	items := newTestInstallments(t, 1)

	db.On("ExecContext", anyArgs(14)...).Return(nil, errors.New("insert fail"))

	err := repo.AddBatch(ctx, items)
	require.Error(t, err)
	assert.ErrorContains(t, err, "mssql: add installment")
}

func TestInstallmentRepository_AddBatch_Empty(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, _ := newInstallmentRepo(t)

	require.NoError(t, repo.AddBatch(ctx, nil))
}

// ---- UpdateBatch ----

func TestInstallmentRepository_UpdateBatch_Success(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newInstallmentRepo(t)
	items := newTestInstallments(t, 2)

	result := devkitmocks.NewMockResult(t)
	// ctx + query + invoiceId + status + updatedAt + deletedAt + id = 7 total per UPDATE
	db.On("ExecContext", anyArgs(7)...).Return(result, nil)

	require.NoError(t, repo.UpdateBatch(ctx, items))
}

func TestInstallmentRepository_UpdateBatch_DBError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newInstallmentRepo(t)
	items := newTestInstallments(t, 1)

	db.On("ExecContext", anyArgs(7)...).Return(nil, errors.New("update fail"))

	err := repo.UpdateBatch(ctx, items)
	require.Error(t, err)
	assert.ErrorContains(t, err, "mssql: update installment")
}

// ---- SoftDeleteByTransaction ----

func TestInstallmentRepository_SoftDeleteByTransaction_Success(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newInstallmentRepo(t)

	result := devkitmocks.NewMockResult(t)
	// ctx + query + deletedAt + updatedAt + transactionId = 5 total
	db.On("ExecContext", anyArgs(5)...).Return(result, nil)

	require.NoError(t, repo.SoftDeleteByTransaction(ctx, vos.NewTransactionID(), time.Now()))
}

func TestInstallmentRepository_SoftDeleteByTransaction_DBError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newInstallmentRepo(t)

	db.On("ExecContext", anyArgs(5)...).Return(nil, errors.New("lock"))

	err := repo.SoftDeleteByTransaction(ctx, vos.NewTransactionID(), time.Now())
	require.Error(t, err)
	assert.ErrorContains(t, err, "mssql: soft delete installments by transaction")
}

// ---- ListByInvoice ----

func TestInstallmentRepository_ListByInvoice_EmptyResult(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newInstallmentRepo(t)

	rows := devkitmocks.NewMockRows(t)
	rows.On("Next").Return(false)
	rows.On("Err").Return(nil)
	rows.On("Close").Return(nil)
	// ctx + query + invoiceId = 3 total
	db.On("QueryContext", anyArgs(3)...).Return(rows, nil)

	items, err := repo.ListByInvoice(ctx, vos.NewInvoiceID())
	require.NoError(t, err)
	assert.Empty(t, items)
}

func TestInstallmentRepository_ListByInvoice_DBError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newInstallmentRepo(t)

	db.On("QueryContext", anyArgs(3)...).Return(nil, errors.New("query fail"))

	_, err := repo.ListByInvoice(ctx, vos.NewInvoiceID())
	require.Error(t, err)
	assert.ErrorContains(t, err, "mssql: list installments by invoice")
}

// ---- ListByTransaction ----

func TestInstallmentRepository_ListByTransaction_EmptyResult(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newInstallmentRepo(t)

	rows := devkitmocks.NewMockRows(t)
	rows.On("Next").Return(false)
	rows.On("Err").Return(nil)
	rows.On("Close").Return(nil)
	// ctx + query + transactionId = 3 total
	db.On("QueryContext", anyArgs(3)...).Return(rows, nil)

	items, err := repo.ListByTransaction(ctx, vos.NewTransactionID())
	require.NoError(t, err)
	assert.Empty(t, items)
}

// ---- HasClosedOrPaidForTransaction ----

func TestInstallmentRepository_HasClosedOrPaidForTransaction_True(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newInstallmentRepo(t)

	row := devkitmocks.NewMockRow(t)
	row.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
		*(args[0].(*int)) = 1
	}).Return(nil)
	// ctx + query + transactionId = 3 total
	db.On("QueryRowContext", anyArgs(3)...).Return(row)

	has, err := repo.HasClosedOrPaidForTransaction(ctx, vos.NewTransactionID())
	require.NoError(t, err)
	assert.True(t, has)
}

func TestInstallmentRepository_HasClosedOrPaidForTransaction_False(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newInstallmentRepo(t)

	row := devkitmocks.NewMockRow(t)
	row.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
		*(args[0].(*int)) = 0
	}).Return(nil)
	db.On("QueryRowContext", anyArgs(3)...).Return(row)

	has, err := repo.HasClosedOrPaidForTransaction(ctx, vos.NewTransactionID())
	require.NoError(t, err)
	assert.False(t, has)
}

func TestInstallmentRepository_HasClosedOrPaidForTransaction_DBError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newInstallmentRepo(t)

	row := devkitmocks.NewMockRow(t)
	row.On("Scan", mock.Anything).Return(errors.New("db err"))
	db.On("QueryRowContext", anyArgs(3)...).Return(row)

	_, err := repo.HasClosedOrPaidForTransaction(ctx, vos.NewTransactionID())
	require.Error(t, err)
}

// ---- SumByMonthCompetence ----

func TestInstallmentRepository_SumByMonthCompetence_Success(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newInstallmentRepo(t)

	period, _ := vos.NewPeriod(2026, 5, nil)
	userID := identityvo.NewUserID()

	row := devkitmocks.NewMockRow(t)
	row.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
		*(args[0].(*string)) = "750.0000"
	}).Return(nil)
	// ctx + query + userId + from + to = 5 total
	db.On("QueryRowContext", anyArgs(5)...).Return(row)

	money, err := repo.SumByMonthCompetence(ctx, userID, period)
	require.NoError(t, err)
	assert.True(t, money.IsPositive())
}

func TestInstallmentRepository_SumByMonthCompetence_DBError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newInstallmentRepo(t)

	period, _ := vos.NewPeriod(2026, 5, nil)
	userID := identityvo.NewUserID()

	row := devkitmocks.NewMockRow(t)
	row.On("Scan", mock.Anything).Return(errors.New("scan fail"))
	db.On("QueryRowContext", anyArgs(5)...).Return(row)

	_, err := repo.SumByMonthCompetence(ctx, userID, period)
	require.Error(t, err)
}

// ---- ListByInvoice with rows (exercises scanInstallmentRows) ----

func setInstallmentScanValues(args mock.Arguments) {
	now := time.Now().UTC()
	*(args[0].(*string)) = validInvID
	*(args[1].(*string)) = validCardID
	*(args[2].(*string)) = validInvID
	*(args[3].(*int)) = 1
	*(args[4].(*int)) = 1
	*(args[5].(*string)) = "100.0000"
	*(args[6].(*string)) = "BRL"
	*(args[7].(*string)) = "scheduled"
	*(args[8].(**string)) = nil
	*(args[9].(*time.Time)) = now
	*(args[10].(*time.Time)) = now
	*(args[11].(**time.Time)) = nil
}

func TestInstallmentRepository_ListByInvoice_WithRows(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newInstallmentRepo(t)

	rows := devkitmocks.NewMockRows(t)
	rows.On("Next").Return(true).Once()
	rows.On("Next").Return(false)
	rows.On("Scan", anyArgs(12)...).Run(setInstallmentScanValues).Return(nil)
	rows.On("Err").Return(nil)
	rows.On("Close").Return(nil)
	db.On("QueryContext", anyArgs(3)...).Return(rows, nil)

	items, err := repo.ListByInvoice(ctx, vos.NewInvoiceID())
	require.NoError(t, err)
	assert.Len(t, items, 1)
}

// ---- ListByTransaction ----

func TestInstallmentRepository_ListByTransaction_DBError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newInstallmentRepo(t)

	db.On("QueryContext", anyArgs(3)...).Return(nil, errors.New("db err"))

	_, err := repo.ListByTransaction(ctx, vos.NewTransactionID())
	require.Error(t, err)
	assert.ErrorContains(t, err, "mssql: list installments by transaction")
}

func TestInstallmentRepository_ListByTransaction_WithRows(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newInstallmentRepo(t)

	rows := devkitmocks.NewMockRows(t)
	rows.On("Next").Return(true).Once()
	rows.On("Next").Return(false)
	rows.On("Scan", anyArgs(12)...).Run(setInstallmentScanValues).Return(nil)
	rows.On("Err").Return(nil)
	rows.On("Close").Return(nil)
	db.On("QueryContext", anyArgs(3)...).Return(rows, nil)

	items, err := repo.ListByTransaction(ctx, vos.NewTransactionID())
	require.NoError(t, err)
	assert.Len(t, items, 1)
}
