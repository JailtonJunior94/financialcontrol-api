package mssql_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	devkitmocks "github.com/JailtonJunior94/devkit-go/pkg/database/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/filters"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/infrastructure/persistence/mssql"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

// ---- shared test helpers ----

type fixedClock struct{ t time.Time }

func (c *fixedClock) Now() time.Time { return c.t }

// anyArgs returns n mock.Anything matchers — required because devkit-go mocks spread
// variadic args individually, so testify needs one matcher per argument position.
func anyArgs(n int) []interface{} {
	args := make([]interface{}, n)
	for i := range args {
		args[i] = mock.Anything
	}
	return args
}

func newTestTransaction(t *testing.T) *entities.Transaction {
	t.Helper()
	id := vos.NewTransactionID()
	userID := identityvo.NewUserID()
	catID := vos.NewCategoryID()
	amount, _ := vos.NewAmount("100.00")
	clock := &fixedClock{time.Now().UTC()}
	tx, err := entities.NewTransaction(
		id, userID, "test description", amount,
		time.Now().UTC(),
		vos.TransactionTypeExpense,
		vos.PaymentMethodPix,
		nil, catID, nil, nil, clock,
	)
	require.NoError(t, err)
	return tx
}

func newTransactionRepo(t *testing.T) (*mssql.TransactionRepository, *devkitmocks.MockDBTX) {
	t.Helper()
	db := devkitmocks.NewMockDBTX(t)
	return mssql.NewTransactionRepository(db), db
}

// ---- Add ----

func TestTransactionRepository_Add_Success(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newTransactionRepo(t)
	tx := newTestTransaction(t)

	result := devkitmocks.NewMockResult(t)
	// ctx + query + 16 named args = 18 total
	db.On("ExecContext", anyArgs(18)...).Return(result, nil)

	require.NoError(t, repo.Add(ctx, tx))
}

func TestTransactionRepository_Add_DBError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newTransactionRepo(t)
	tx := newTestTransaction(t)

	db.On("ExecContext", anyArgs(18)...).Return(nil, errors.New("db error"))

	err := repo.Add(ctx, tx)
	require.Error(t, err)
	assert.ErrorContains(t, err, "mssql: add transaction")
}

// ---- Update ----

func TestTransactionRepository_Update_Success(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newTransactionRepo(t)
	tx := newTestTransaction(t)

	result := devkitmocks.NewMockResult(t)
	// ctx + query + 13 named args = 15 total
	db.On("ExecContext", anyArgs(15)...).Return(result, nil)

	require.NoError(t, repo.Update(ctx, tx))
}

func TestTransactionRepository_Update_DBError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newTransactionRepo(t)
	tx := newTestTransaction(t)

	db.On("ExecContext", anyArgs(15)...).Return(nil, errors.New("timeout"))

	err := repo.Update(ctx, tx)
	require.Error(t, err)
	assert.ErrorContains(t, err, "mssql: update transaction")
}

// ---- GetByID ----

func TestTransactionRepository_GetByID_NotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newTransactionRepo(t)

	row := devkitmocks.NewMockRow(t)
	// transaction scan has 16 columns
	row.On("Scan", anyArgs(16)...).Return(sql.ErrNoRows)
	// ctx + query + userId + id = 4 total
	db.On("QueryRowContext", anyArgs(4)...).Return(row)

	result, err := repo.GetByID(ctx, identityvo.NewUserID(), vos.NewTransactionID())
	assert.Nil(t, result)
	require.ErrorIs(t, err, domain.ErrTransactionNotFound)
}

func TestTransactionRepository_GetByID_DBError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newTransactionRepo(t)

	row := devkitmocks.NewMockRow(t)
	row.On("Scan", anyArgs(16)...).Return(errors.New("conn error"))
	db.On("QueryRowContext", anyArgs(4)...).Return(row)

	result, err := repo.GetByID(ctx, identityvo.NewUserID(), vos.NewTransactionID())
	assert.Nil(t, result)
	require.Error(t, err)
	assert.ErrorContains(t, err, "mssql: get transaction by id")
}

// ---- SoftDelete ----

func TestTransactionRepository_SoftDelete_Success(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newTransactionRepo(t)
	tx := newTestTransaction(t)

	result := devkitmocks.NewMockResult(t)
	// ctx + query + deletedAt + updatedAt + id + userId = 6 total
	db.On("ExecContext", anyArgs(6)...).Return(result, nil)

	require.NoError(t, repo.SoftDelete(ctx, tx, time.Now()))
}

func TestTransactionRepository_SoftDelete_DBError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newTransactionRepo(t)
	tx := newTestTransaction(t)

	db.On("ExecContext", anyArgs(6)...).Return(nil, errors.New("lock"))

	err := repo.SoftDelete(ctx, tx, time.Now())
	require.Error(t, err)
	assert.ErrorContains(t, err, "mssql: soft delete transaction")
}

// ---- HasActiveRefundFor ----

func TestTransactionRepository_HasActiveRefundFor_True(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newTransactionRepo(t)

	row := devkitmocks.NewMockRow(t)
	row.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
		*(args[0].(*int)) = 1
	}).Return(nil)
	// ctx + query + userId + originalId = 4 total
	db.On("QueryRowContext", anyArgs(4)...).Return(row)

	has, err := repo.HasActiveRefundFor(ctx, identityvo.NewUserID(), vos.NewTransactionID())
	require.NoError(t, err)
	assert.True(t, has)
}

func TestTransactionRepository_HasActiveRefundFor_False(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newTransactionRepo(t)

	row := devkitmocks.NewMockRow(t)
	row.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
		*(args[0].(*int)) = 0
	}).Return(nil)
	db.On("QueryRowContext", anyArgs(4)...).Return(row)

	has, err := repo.HasActiveRefundFor(ctx, identityvo.NewUserID(), vos.NewTransactionID())
	require.NoError(t, err)
	assert.False(t, has)
}

func TestTransactionRepository_HasActiveRefundFor_DBError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newTransactionRepo(t)

	row := devkitmocks.NewMockRow(t)
	row.On("Scan", mock.Anything).Return(errors.New("scan err"))
	db.On("QueryRowContext", anyArgs(4)...).Return(row)

	_, err := repo.HasActiveRefundFor(ctx, identityvo.NewUserID(), vos.NewTransactionID())
	require.Error(t, err)
}

// ---- List ----

func TestTransactionRepository_List_CountDBError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newTransactionRepo(t)

	userID := identityvo.NewUserID()
	pg, _ := vos.NewPagination(1, 20)
	f, _ := filters.NewTransactionFilter(userID, nil, nil, nil, nil,
		vos.InvoiceStatusFilterNone, "", nil, nil, pg)

	row := devkitmocks.NewMockRow(t)
	row.On("Scan", mock.Anything).Return(errors.New("count fail"))
	// ctx + query + userId = 3 total (minimal filter: just userId)
	db.On("QueryRowContext", anyArgs(3)...).Return(row)

	items, total, err := repo.List(ctx, userID, f)
	assert.Nil(t, items)
	assert.Zero(t, total)
	require.Error(t, err)
}

func TestTransactionRepository_List_QueryDBError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newTransactionRepo(t)

	userID := identityvo.NewUserID()
	pg, _ := vos.NewPagination(1, 20)
	f, _ := filters.NewTransactionFilter(userID, nil, nil, nil, nil,
		vos.InvoiceStatusFilterNone, "", nil, nil, pg)

	row := devkitmocks.NewMockRow(t)
	row.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
		*(args[0].(*int64)) = 5
	}).Return(nil)
	db.On("QueryRowContext", anyArgs(3)...).Return(row)
	// ctx + query + userId + offset + size = 5 total
	db.On("QueryContext", anyArgs(5)...).Return(nil, errors.New("timeout"))

	_, _, err := repo.List(ctx, userID, f)
	require.Error(t, err)
}

func TestTransactionRepository_List_EmptyResult(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newTransactionRepo(t)

	userID := identityvo.NewUserID()
	pg, _ := vos.NewPagination(1, 20)
	f, _ := filters.NewTransactionFilter(userID, nil, nil, nil, nil,
		vos.InvoiceStatusFilterNone, "", nil, nil, pg)

	row := devkitmocks.NewMockRow(t)
	row.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
		*(args[0].(*int64)) = 0
	}).Return(nil)
	db.On("QueryRowContext", anyArgs(3)...).Return(row)

	rows := devkitmocks.NewMockRows(t)
	rows.On("Next").Return(false)
	rows.On("Err").Return(nil)
	rows.On("Close").Return(nil)
	db.On("QueryContext", anyArgs(5)...).Return(rows, nil)

	items, total, err := repo.List(ctx, userID, f)
	require.NoError(t, err)
	assert.Zero(t, total)
	assert.Empty(t, items)
}

func TestTransactionRepository_List_WithDescriptionFilter(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newTransactionRepo(t)

	userID := identityvo.NewUserID()
	pg, _ := vos.NewPagination(1, 20)
	// description_contains does NOT duplicate COLLATE (RF-19 + C3.a)
	f, _ := filters.NewTransactionFilter(userID, nil, nil, nil, nil,
		vos.InvoiceStatusFilterNone, "café", nil, nil, pg)

	row := devkitmocks.NewMockRow(t)
	row.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
		*(args[0].(*int64)) = 0
	}).Return(nil)
	// ctx + query + userId + term = 4 total (description filter adds @term)
	db.On("QueryRowContext", anyArgs(4)...).Return(row)

	rows := devkitmocks.NewMockRows(t)
	rows.On("Next").Return(false)
	rows.On("Err").Return(nil)
	rows.On("Close").Return(nil)
	// ctx + query + userId + term + offset + size = 6 total
	db.On("QueryContext", anyArgs(6)...).Return(rows, nil)

	items, _, err := repo.List(ctx, userID, f)
	require.NoError(t, err)
	assert.Empty(t, items)
}

// ---- SumForSummary ----

func TestTransactionRepository_SumForSummary_Success(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newTransactionRepo(t)

	period, _ := vos.NewPeriod(2026, 5, nil)

	row := devkitmocks.NewMockRow(t)
	row.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		*(args[0].(*string)) = "500.0000"
		*(args[1].(*string)) = "300.0000"
		*(args[2].(*string)) = "50.0000"
		*(args[3].(*string)) = "20.0000"
	}).Return(nil)
	// ctx + query + userId + from + to = 5 total
	db.On("QueryRowContext", anyArgs(5)...).Return(row)

	agg, err := repo.SumForSummary(ctx, identityvo.NewUserID(), period)
	require.NoError(t, err)
	assert.True(t, agg.TotalIncome.IsPositive())
	assert.True(t, agg.TotalExpense.IsPositive())
}

func TestTransactionRepository_SumForSummary_DBError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newTransactionRepo(t)

	period, _ := vos.NewPeriod(2026, 5, nil)

	row := devkitmocks.NewMockRow(t)
	row.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("db err"))
	db.On("QueryRowContext", anyArgs(5)...).Return(row)

	_, err := repo.SumForSummary(ctx, identityvo.NewUserID(), period)
	require.Error(t, err)
}

// ---- GetByID success (exercises RowToTransaction) ----

func setTransactionScanValues(args mock.Arguments) {
	now := time.Now().UTC()
	*(args[0].(*string)) = validInvID
	*(args[1].(*string)) = validUserID
	*(args[2].(*string)) = "test description"
	*(args[3].(*string)) = "100.0000"
	*(args[4].(*string)) = "BRL"
	*(args[5].(*time.Time)) = now
	*(args[6].(*string)) = "expense"
	*(args[7].(*string)) = "pix"
	*(args[8].(**string)) = nil
	*(args[9].(*string)) = validCardID
	*(args[10].(**string)) = nil
	*(args[11].(**string)) = nil
	*(args[12].(**string)) = nil
	*(args[13].(*time.Time)) = now
	*(args[14].(*time.Time)) = now
	*(args[15].(**time.Time)) = nil
}

func TestTransactionRepository_GetByID_Success(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newTransactionRepo(t)

	row := devkitmocks.NewMockRow(t)
	row.On("Scan", anyArgs(16)...).Run(setTransactionScanValues).Return(nil)
	db.On("QueryRowContext", anyArgs(4)...).Return(row)

	tx, err := repo.GetByID(ctx, identityvo.NewUserID(), vos.NewTransactionID())
	require.NoError(t, err)
	require.NotNil(t, tx)
}

// ---- List with rows (exercises scan path) ----

func TestTransactionRepository_List_WithRows(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newTransactionRepo(t)

	userID := identityvo.NewUserID()
	pg, _ := vos.NewPagination(1, 20)
	f, _ := filters.NewTransactionFilter(userID, nil, nil, nil, nil,
		vos.InvoiceStatusFilterNone, "", nil, nil, pg)

	row := devkitmocks.NewMockRow(t)
	row.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
		*(args[0].(*int64)) = 1
	}).Return(nil)
	db.On("QueryRowContext", anyArgs(3)...).Return(row)

	rows := devkitmocks.NewMockRows(t)
	rows.On("Next").Return(true).Once()
	rows.On("Next").Return(false)
	rows.On("Scan", anyArgs(16)...).Run(setTransactionScanValues).Return(nil)
	rows.On("Err").Return(nil)
	rows.On("Close").Return(nil)
	db.On("QueryContext", anyArgs(5)...).Return(rows, nil)

	items, total, err := repo.List(ctx, userID, f)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, items, 1)
}

// ---- List with all filters (exercises every buildTransactionWhere branch) ----

func TestTransactionRepository_List_AllFilters(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newTransactionRepo(t)

	userID := identityvo.NewUserID()
	pg, _ := vos.NewPagination(1, 20)
	txType := vos.TransactionTypeExpense
	pm := vos.PaymentMethodPix
	cardID := vos.NewCardID()
	catID := vos.NewCategoryID()
	from := time.Now().UTC().Add(-24 * time.Hour)
	to := time.Now().UTC()
	f, _ := filters.NewTransactionFilter(userID, &txType, &pm, &cardID, &catID,
		vos.InvoiceStatusFilterOpen, "", &from, &to, pg)

	// ctx + query + userId + from + to + txType + pm + cardId + catId + invoiceStatus = 10
	row := devkitmocks.NewMockRow(t)
	row.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
		*(args[0].(*int64)) = 0
	}).Return(nil)
	db.On("QueryRowContext", anyArgs(10)...).Return(row)

	rows := devkitmocks.NewMockRows(t)
	rows.On("Next").Return(false)
	rows.On("Err").Return(nil)
	rows.On("Close").Return(nil)
	// ctx + query + 8 filter args + offset + size = 12
	db.On("QueryContext", anyArgs(12)...).Return(rows, nil)

	items, total, err := repo.List(ctx, userID, f)
	require.NoError(t, err)
	assert.Zero(t, total)
	assert.Empty(t, items)
}
