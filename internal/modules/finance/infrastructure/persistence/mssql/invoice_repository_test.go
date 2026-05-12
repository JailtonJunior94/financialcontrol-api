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
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/projections"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/infrastructure/persistence/mssql"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

func newInvoiceRepo(t *testing.T) (*mssql.InvoiceRepository, *devkitmocks.MockDBTX) {
	t.Helper()
	db := devkitmocks.NewMockDBTX(t)
	return mssql.NewInvoiceRepository(db), db
}

func newTestInvoice(t *testing.T) *entities.Invoice {
	t.Helper()
	card := projections.CardView{
		ID:           vos.NewCardID(),
		UserID:       identityvo.NewUserID(),
		ClosingDay:   5,
		DueDay:       15,
		BillingCycle: 30,
		Active:       true,
	}
	now := time.Now().UTC()
	inv, err := entities.NewInvoice(card, now, now.AddDate(0, 1, 0), now.AddDate(0, 1, 0), now.AddDate(0, 1, 10), now)
	require.NoError(t, err)
	return inv
}

// setInvoiceScanValues fills the scan args with valid invoice data.
func setInvoiceScanValues(args mock.Arguments) {
	now := time.Now().UTC()
	*(args[0].(*string)) = validInvID
	*(args[1].(*string)) = validUserID
	*(args[2].(*string)) = validCardID
	*(args[3].(*string)) = "open"
	*(args[4].(*time.Time)) = now.AddDate(0, -1, 0)
	*(args[5].(*time.Time)) = now
	*(args[6].(*time.Time)) = now
	*(args[7].(*time.Time)) = now.AddDate(0, 0, 10)
	*(args[8].(*string)) = "0.0000"
	*(args[9].(*string)) = "BRL"
	*(args[10].(**time.Time)) = nil
	*(args[11].(**string)) = nil
	*(args[12].(*time.Time)) = now
	*(args[13].(*time.Time)) = now
	*(args[14].(**time.Time)) = nil
}

// ---- GetByID ----

func TestInvoiceRepository_GetByID_NotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newInvoiceRepo(t)

	row := devkitmocks.NewMockRow(t)
	// invoice scan has 15 columns
	row.On("Scan", anyArgs(15)...).Return(sql.ErrNoRows)
	// ctx + query + userId + id = 4 total
	db.On("QueryRowContext", anyArgs(4)...).Return(row)

	result, err := repo.GetByID(ctx, identityvo.NewUserID(), vos.NewInvoiceID())
	assert.Nil(t, result)
	require.ErrorIs(t, err, domain.ErrInvoiceNotFound)
}

func TestInvoiceRepository_GetByID_DBError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newInvoiceRepo(t)

	row := devkitmocks.NewMockRow(t)
	row.On("Scan", anyArgs(15)...).Return(errors.New("timeout"))
	db.On("QueryRowContext", anyArgs(4)...).Return(row)

	_, err := repo.GetByID(ctx, identityvo.NewUserID(), vos.NewInvoiceID())
	require.Error(t, err)
	assert.ErrorContains(t, err, "mssql: get invoice by id")
}

func TestInvoiceRepository_GetByID_Success(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newInvoiceRepo(t)

	row := devkitmocks.NewMockRow(t)
	row.On("Scan",
		mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything,
	).Run(setInvoiceScanValues).Return(nil)
	db.On("QueryRowContext", anyArgs(4)...).Return(row)

	inv, err := repo.GetByID(ctx, identityvo.NewUserID(), vos.NewInvoiceID())
	require.NoError(t, err)
	require.NotNil(t, inv)
	assert.True(t, inv.IsOpen())
}

// ---- Update ----

func TestInvoiceRepository_Update_Success(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newInvoiceRepo(t)
	inv := newTestInvoice(t)

	result := devkitmocks.NewMockResult(t)
	// ctx + query + state + total + paidAt + updatedAt + deletedAt + id + userId = 9 total
	db.On("ExecContext", anyArgs(9)...).Return(result, nil)

	require.NoError(t, repo.Update(ctx, inv))
}

func TestInvoiceRepository_Update_DBError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newInvoiceRepo(t)
	inv := newTestInvoice(t)

	db.On("ExecContext", anyArgs(9)...).Return(nil, errors.New("db err"))

	err := repo.Update(ctx, inv)
	require.Error(t, err)
	assert.ErrorContains(t, err, "mssql: update invoice")
}

// ---- List ----

func TestInvoiceRepository_List_EmptyResult(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newInvoiceRepo(t)

	userID := identityvo.NewUserID()
	pg, _ := vos.NewPagination(1, 20)
	f, _ := filters.NewInvoiceFilter(userID, nil, nil, nil, nil, pg)

	row := devkitmocks.NewMockRow(t)
	row.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
		*(args[0].(*int64)) = 0
	}).Return(nil)
	// ctx + query + userId = 3 total (minimal filter)
	db.On("QueryRowContext", anyArgs(3)...).Return(row)

	rows := devkitmocks.NewMockRows(t)
	rows.On("Next").Return(false)
	rows.On("Err").Return(nil)
	rows.On("Close").Return(nil)
	// ctx + query + userId + offset + size = 5 total
	db.On("QueryContext", anyArgs(5)...).Return(rows, nil)

	items, total, err := repo.List(ctx, userID, f)
	require.NoError(t, err)
	assert.Zero(t, total)
	assert.Empty(t, items)
}

func TestInvoiceRepository_List_DBError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newInvoiceRepo(t)

	userID := identityvo.NewUserID()
	pg, _ := vos.NewPagination(1, 20)
	f, _ := filters.NewInvoiceFilter(userID, nil, nil, nil, nil, pg)

	row := devkitmocks.NewMockRow(t)
	row.On("Scan", mock.Anything).Return(errors.New("count fail"))
	db.On("QueryRowContext", anyArgs(3)...).Return(row)

	_, _, err := repo.List(ctx, userID, f)
	require.Error(t, err)
}

// ---- NextOpenFor ----

func TestInvoiceRepository_NextOpenFor_NotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newInvoiceRepo(t)

	row := devkitmocks.NewMockRow(t)
	row.On("Scan", anyArgs(15)...).Return(sql.ErrNoRows)
	// ctx + query + userId + cardId = 4 total
	db.On("QueryRowContext", anyArgs(4)...).Return(row)

	inv, err := repo.NextOpenFor(ctx, identityvo.NewUserID(), vos.NewCardID(), &fixedClock{time.Now().UTC()})
	require.NoError(t, err)
	assert.Nil(t, inv)
}

func TestInvoiceRepository_NextOpenFor_DBError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newInvoiceRepo(t)

	row := devkitmocks.NewMockRow(t)
	row.On("Scan", anyArgs(15)...).Return(errors.New("conn fail"))
	db.On("QueryRowContext", anyArgs(4)...).Return(row)

	_, err := repo.NextOpenFor(ctx, identityvo.NewUserID(), vos.NewCardID(), &fixedClock{time.Now().UTC()})
	require.Error(t, err)
}

// ---- SumPaidInPeriod ----

func TestInvoiceRepository_SumPaidInPeriod_Success(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newInvoiceRepo(t)

	period, _ := vos.NewPeriod(2026, 5, nil)

	row := devkitmocks.NewMockRow(t)
	row.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
		*(args[0].(*string)) = "1500.0000"
	}).Return(nil)
	// ctx + query + userId + from + to = 5 total
	db.On("QueryRowContext", anyArgs(5)...).Return(row)

	money, err := repo.SumPaidInPeriod(ctx, identityvo.NewUserID(), period)
	require.NoError(t, err)
	assert.True(t, money.IsPositive())
}

func TestInvoiceRepository_SumPaidInPeriod_DBError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newInvoiceRepo(t)

	period, _ := vos.NewPeriod(2026, 5, nil)

	row := devkitmocks.NewMockRow(t)
	row.On("Scan", mock.Anything).Return(errors.New("scan err"))
	db.On("QueryRowContext", anyArgs(5)...).Return(row)

	_, err := repo.SumPaidInPeriod(ctx, identityvo.NewUserID(), period)
	require.Error(t, err)
}

// ---- SumOpenForUser ----

func TestInvoiceRepository_SumOpenForUser_Success(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newInvoiceRepo(t)

	period, _ := vos.NewPeriod(2026, 5, nil)

	row := devkitmocks.NewMockRow(t)
	row.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
		*(args[0].(*string)) = "2000.0000"
	}).Return(nil)
	// ctx + query + userId + from + to = 5 total
	db.On("QueryRowContext", anyArgs(5)...).Return(row)

	money, err := repo.SumOpenForUser(ctx, identityvo.NewUserID(), period)
	require.NoError(t, err)
	assert.True(t, money.IsPositive())
}

// ---- AssignOrCreateOpen: finds existing ----

func TestInvoiceRepository_AssignOrCreateOpen_FindsExisting(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newInvoiceRepo(t)

	userID := identityvo.NewUserID()
	cardID := vos.NewCardID()
	card := projections.CardView{
		ID:           cardID,
		UserID:       userID,
		ClosingDay:   5,
		DueDay:       15,
		BillingCycle: 30,
		Active:       true,
	}
	// occurred on May 3, 2026 (day 3 < closing day 5 → target closing = May 5, 2026)
	sp, _ := time.LoadLocation("America/Sao_Paulo")
	occurred := time.Date(2026, 5, 3, 10, 0, 0, 0, sp)

	// QueryRowContext call for getOpenInvoiceForCardClosing succeeds
	row := devkitmocks.NewMockRow(t)
	row.On("Scan",
		mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything,
	).Run(setInvoiceScanValues).Return(nil)
	// ctx + query + userId + cardId + closingDate = 5 total
	db.On("QueryRowContext", anyArgs(5)...).Return(row)

	inv, err := repo.AssignOrCreateOpen(ctx, userID, cardID, occurred, card, &fixedClock{time.Now().UTC()})
	require.NoError(t, err)
	require.NotNil(t, inv)
}

// TestInvoiceRepository_AssignOrCreateOpen_CreatesNew verifies a new invoice is inserted
// when no open invoice exists for the cycle.
func TestInvoiceRepository_AssignOrCreateOpen_CreatesNew(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newInvoiceRepo(t)

	userID := identityvo.NewUserID()
	cardID := vos.NewCardID()
	card := projections.CardView{
		ID:           cardID,
		UserID:       userID,
		ClosingDay:   5,
		DueDay:       15,
		BillingCycle: 30,
		Active:       true,
	}
	sp, _ := time.LoadLocation("America/Sao_Paulo")
	occurred := time.Date(2026, 5, 3, 10, 0, 0, 0, sp)

	// SELECT returns no rows → no existing invoice
	row := devkitmocks.NewMockRow(t)
	row.On("Scan", anyArgs(15)...).Return(sql.ErrNoRows)
	db.On("QueryRowContext", anyArgs(5)...).Return(row)

	// INSERT for the new invoice: ctx + query + 15 named args = 17 total
	result := devkitmocks.NewMockResult(t)
	db.On("ExecContext", anyArgs(17)...).Return(result, nil)

	inv, err := repo.AssignOrCreateOpen(ctx, userID, cardID, occurred, card, &fixedClock{time.Now().UTC()})
	require.NoError(t, err)
	require.NotNil(t, inv)
	assert.True(t, inv.IsOpen())
}

// ---- SumOpenForUser DBError ----

func TestInvoiceRepository_SumOpenForUser_DBError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newInvoiceRepo(t)

	period, _ := vos.NewPeriod(2026, 5, nil)

	row := devkitmocks.NewMockRow(t)
	row.On("Scan", mock.Anything).Return(errors.New("scan err"))
	db.On("QueryRowContext", anyArgs(5)...).Return(row)

	_, err := repo.SumOpenForUser(ctx, identityvo.NewUserID(), period)
	require.Error(t, err)
}

// ---- List with all invoice filters (exercises every buildInvoiceWhere branch) ----

func TestInvoiceRepository_List_AllFilters(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newInvoiceRepo(t)

	userID := identityvo.NewUserID()
	pg, _ := vos.NewPagination(1, 20)
	cardID := vos.NewCardID()
	state := vos.InvoiceStateOpen
	from := time.Now().UTC().Add(-24 * time.Hour)
	to := time.Now().UTC()
	f, _ := filters.NewInvoiceFilter(userID, &cardID, &state, &from, &to, pg)

	// ctx + query + userId + cardId + state + from + to = 7
	row := devkitmocks.NewMockRow(t)
	row.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
		*(args[0].(*int64)) = 0
	}).Return(nil)
	db.On("QueryRowContext", anyArgs(7)...).Return(row)

	rows := devkitmocks.NewMockRows(t)
	rows.On("Next").Return(false)
	rows.On("Err").Return(nil)
	rows.On("Close").Return(nil)
	// ctx + query + 5 filter args + offset + size = 9
	db.On("QueryContext", anyArgs(9)...).Return(rows, nil)

	items, total, err := repo.List(ctx, userID, f)
	require.NoError(t, err)
	assert.Zero(t, total)
	assert.Empty(t, items)
}

// ---- AssignOrCreateOpen with day >= closingDay (exercises second branch of computeTargetClosingDate) ----

func TestInvoiceRepository_AssignOrCreateOpen_DayAfterClosing(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newInvoiceRepo(t)

	userID := identityvo.NewUserID()
	cardID := vos.NewCardID()
	// DueDay < ClosingDay → triggers else branch in computeInvoiceDates
	card := projections.CardView{
		ID:           cardID,
		UserID:       userID,
		ClosingDay:   20,
		DueDay:       5,
		BillingCycle: 30,
		Active:       true,
	}
	sp, _ := time.LoadLocation("America/Sao_Paulo")
	// day 25 >= closingDay 20 → target closing = next month (June 20)
	occurred := time.Date(2026, 5, 25, 10, 0, 0, 0, sp)

	row := devkitmocks.NewMockRow(t)
	row.On("Scan", anyArgs(15)...).Return(sql.ErrNoRows)
	db.On("QueryRowContext", anyArgs(5)...).Return(row)

	result := devkitmocks.NewMockResult(t)
	db.On("ExecContext", anyArgs(17)...).Return(result, nil)

	inv, err := repo.AssignOrCreateOpen(ctx, userID, cardID, occurred, card, &fixedClock{time.Now().UTC()})
	require.NoError(t, err)
	require.NotNil(t, inv)
}
