package mssql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	devkitdb "github.com/JailtonJunior94/devkit-go/pkg/database"
	"github.com/shopspring/decimal"

	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/filters"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/ports"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
	pkgdatabase "github.com/jailtonjunior94/financialcontrol-api/pkg/database"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

var _ ports.TransactionRepository = (*TransactionRepository)(nil)

// TransactionRepository implements ports.TransactionRepository against MSSQL.
type TransactionRepository struct {
	db devkitdb.DBTX
}

func NewTransactionRepository(db devkitdb.DBTX) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) Add(ctx context.Context, t *entities.Transaction) error {
	_, err := r.db.ExecContext(ctx, addTransaction,
		sql.Named("id", t.ID().String()),
		sql.Named("userId", t.UserID().String()),
		sql.Named("description", t.Description()),
		sql.Named("amount", t.Amount().Money().Amount().StringFixed(4)),
		sql.Named("currency", "BRL"),
		sql.Named("occurredAt", t.OccurredAt()),
		sql.Named("transactionType", t.TransactionType().String()),
		sql.Named("paymentMethod", t.PaymentMethod().String()),
		sql.Named("cardId", nullableString(cardIDPtr(t.CardID()))),
		sql.Named("categoryId", t.CategoryID().String()),
		sql.Named("subcategoryId", nullableString(categoryIDPtr(t.SubcategoryID()))),
		sql.Named("originalTransactionId", nullableString(transactionIDPtr(t.OriginalTransactionID()))),
		sql.Named("legacyOrigin", t.LegacyOrigin()),
		sql.Named("createdAt", t.CreatedAt()),
		sql.Named("updatedAt", t.UpdatedAt()),
		sql.Named("deletedAt", t.DeletedAt()),
	)
	if err != nil {
		return fmt.Errorf("mssql: add transaction: %w", err)
	}
	return nil
}

func (r *TransactionRepository) Update(ctx context.Context, t *entities.Transaction) error {
	_, err := r.db.ExecContext(ctx, updateTransaction,
		sql.Named("description", t.Description()),
		sql.Named("amount", t.Amount().Money().Amount().StringFixed(4)),
		sql.Named("occurredAt", t.OccurredAt()),
		sql.Named("transactionType", t.TransactionType().String()),
		sql.Named("paymentMethod", t.PaymentMethod().String()),
		sql.Named("cardId", nullableString(cardIDPtr(t.CardID()))),
		sql.Named("categoryId", t.CategoryID().String()),
		sql.Named("subcategoryId", nullableString(categoryIDPtr(t.SubcategoryID()))),
		sql.Named("originalTransactionId", nullableString(transactionIDPtr(t.OriginalTransactionID()))),
		sql.Named("updatedAt", t.UpdatedAt()),
		sql.Named("deletedAt", t.DeletedAt()),
		sql.Named("id", t.ID().String()),
		sql.Named("userId", t.UserID().String()),
	)
	if err != nil {
		return fmt.Errorf("mssql: update transaction: %w", err)
	}
	return nil
}

func (r *TransactionRepository) GetByID(ctx context.Context, userID identityvo.UserID, id vos.TransactionID) (*entities.Transaction, error) {
	var row TransactionRow
	err := r.db.QueryRowContext(ctx, getTransactionByID,
		sql.Named("userId", userID.String()),
		sql.Named("id", id.String()),
	).Scan(
		&row.ID, &row.UserID, &row.Description, &row.Amount, &row.Currency,
		&row.OccurredAt, &row.TransactionType, &row.PaymentMethod,
		&row.CardID, &row.CategoryID, &row.SubcategoryID, &row.OriginalTransactionID,
		&row.LegacyOrigin, &row.CreatedAt, &row.UpdatedAt, &row.DeletedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrTransactionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("mssql: get transaction by id: %w", err)
	}
	tx, err := RowToTransaction(&row)
	if err != nil {
		return nil, fmt.Errorf("mssql: get transaction by id map: %w", err)
	}
	return tx, nil
}

// List returns a paginated list of transactions matching the filter.
// Returns (items, total, error).
func (r *TransactionRepository) List(ctx context.Context, userID identityvo.UserID, f filters.TransactionFilter) ([]entities.Transaction, int64, error) {
	where, args := buildTransactionWhere(userID, f)

	// count query
	countQ := "SELECT COUNT(1) FROM dbo.FinanceTransactions (NOLOCK) WHERE " + where
	var total int64
	if err := r.db.QueryRowContext(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("mssql: list transactions count: %w", err)
	}

	// data query with pagination
	dataQ := transactionListSelect + " WHERE " + where +
		transactionListOrder + transactionListPaging
	args = append(args,
		sql.Named("offset", f.Pagination.Offset()),
		sql.Named("size", f.Pagination.PageSize()),
	)

	rows, err := r.db.QueryContext(ctx, dataQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("mssql: list transactions: %w", err)
	}

	items, err := pkgdatabase.ScanAll[entities.Transaction](rows, func(r devkitdb.Rows) (entities.Transaction, error) {
		var row TransactionRow
		if scanErr := r.Scan(
			&row.ID, &row.UserID, &row.Description, &row.Amount, &row.Currency,
			&row.OccurredAt, &row.TransactionType, &row.PaymentMethod,
			&row.CardID, &row.CategoryID, &row.SubcategoryID, &row.OriginalTransactionID,
			&row.LegacyOrigin, &row.CreatedAt, &row.UpdatedAt, &row.DeletedAt,
		); scanErr != nil {
			return entities.Transaction{}, fmt.Errorf("mssql: list transactions scan: %w", scanErr)
		}
		tx, mapErr := RowToTransaction(&row)
		if mapErr != nil {
			return entities.Transaction{}, fmt.Errorf("mssql: list transactions map: %w", mapErr)
		}
		return *tx, nil
	})
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *TransactionRepository) SoftDelete(ctx context.Context, t *entities.Transaction, at time.Time) error {
	_, err := r.db.ExecContext(ctx, softDeleteTransaction,
		sql.Named("deletedAt", at.UTC()),
		sql.Named("updatedAt", at.UTC()),
		sql.Named("id", t.ID().String()),
		sql.Named("userId", t.UserID().String()),
	)
	if err != nil {
		return fmt.Errorf("mssql: soft delete transaction: %w", err)
	}
	return nil
}

func (r *TransactionRepository) HasActiveRefundFor(ctx context.Context, userID identityvo.UserID, original vos.TransactionID) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, hasActiveRefundFor,
		sql.Named("userId", userID.String()),
		sql.Named("originalId", original.String()),
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("mssql: has active refund for: %w", err)
	}
	return count > 0, nil
}

func (r *TransactionRepository) SumForSummary(ctx context.Context, userID identityvo.UserID, period vos.Period) (ports.SummaryAggregates, error) {
	from, to := period.Bounds()
	var incomeStr, expenseStr, refundsInStr, refundsOutStr string
	err := r.db.QueryRowContext(ctx, sumForSummary,
		sql.Named("userId", userID.String()),
		sql.Named("from", from),
		sql.Named("to", to),
	).Scan(&incomeStr, &expenseStr, &refundsInStr, &refundsOutStr)
	if err != nil {
		return ports.SummaryAggregates{}, fmt.Errorf("mssql: sum for summary: %w", err)
	}
	incomeDec, err := decimal.NewFromString(incomeStr)
	if err != nil {
		return ports.SummaryAggregates{}, fmt.Errorf("mssql: sum for summary income parse: %w", err)
	}
	expenseDec, err := decimal.NewFromString(expenseStr)
	if err != nil {
		return ports.SummaryAggregates{}, fmt.Errorf("mssql: sum for summary expense parse: %w", err)
	}
	refundsInDec, err := decimal.NewFromString(refundsInStr)
	if err != nil {
		return ports.SummaryAggregates{}, fmt.Errorf("mssql: sum for summary refunds_in parse: %w", err)
	}
	refundsOutDec, err := decimal.NewFromString(refundsOutStr)
	if err != nil {
		return ports.SummaryAggregates{}, fmt.Errorf("mssql: sum for summary refunds_out parse: %w", err)
	}
	return ports.SummaryAggregates{
		TotalIncome:     vos.NewMoneyFromDecimal(incomeDec),
		TotalExpense:    vos.NewMoneyFromDecimal(expenseDec),
		TotalRefundsIn:  vos.NewMoneyFromDecimal(refundsInDec),
		TotalRefundsOut: vos.NewMoneyFromDecimal(refundsOutDec),
	}, nil
}

// buildTransactionWhere constructs the WHERE clause and named args for the filter.
// Every query always filters by userId and deleted_at IS NULL (R-SEC-001).
func buildTransactionWhere(userID identityvo.UserID, f filters.TransactionFilter) (string, []any) {
	var sb strings.Builder
	args := make([]any, 0, 10)

	sb.WriteString("[UserId] = @userId AND [DeletedAt] IS NULL")
	args = append(args, sql.Named("userId", userID.String()))

	if f.From != nil {
		sb.WriteString(" AND [OccurredAt] >= @from")
		args = append(args, sql.Named("from", *f.From))
	}
	if f.To != nil {
		sb.WriteString(" AND [OccurredAt] < @to")
		args = append(args, sql.Named("to", *f.To))
	}
	if f.TransactionType != nil {
		sb.WriteString(" AND [TransactionType] = @txType")
		args = append(args, sql.Named("txType", f.TransactionType.String()))
	}
	if f.PaymentMethod != nil {
		sb.WriteString(" AND [PaymentMethod] = @paymentMethod")
		args = append(args, sql.Named("paymentMethod", f.PaymentMethod.String()))
	}
	if f.CardID != nil {
		sb.WriteString(" AND [CardId] = @cardId")
		args = append(args, sql.Named("cardId", f.CardID.String()))
	}
	if f.CategoryID != nil {
		sb.WriteString(" AND [CategoryId] = @categoryId")
		args = append(args, sql.Named("categoryId", f.CategoryID.String()))
	}
	if f.DescriptionContains != "" {
		sb.WriteString(" AND [Description] LIKE @term")
		args = append(args, sql.Named("term", "%"+f.DescriptionContains+"%"))
	}
	if f.InvoiceStatus.IsActive() {
		sb.WriteString(" AND EXISTS (SELECT 1 FROM dbo.FinanceInstallments inst (NOLOCK)" +
			" INNER JOIN dbo.FinanceInvoices inv (NOLOCK) ON inv.[Id] = inst.[InvoiceId]" +
			" WHERE inst.[TransactionId] = [Id]" +
			" AND inst.[DeletedAt] IS NULL AND inv.[DeletedAt] IS NULL" +
			" AND inv.[State] = @invoiceStatus)")
		args = append(args, sql.Named("invoiceStatus", f.InvoiceStatus.String()))
	}

	return sb.String(), args
}

// nullableString converts a *string to interface{} (nil when pointer is nil).
func nullableString(s *string) interface{} {
	if s == nil {
		return nil
	}
	return *s
}

func cardIDPtr(c *vos.CardID) *string {
	if c == nil {
		return nil
	}
	s := c.String()
	return &s
}

func categoryIDPtr(c *vos.CategoryID) *string {
	if c == nil {
		return nil
	}
	s := c.String()
	return &s
}

func transactionIDPtr(t *vos.TransactionID) *string {
	if t == nil {
		return nil
	}
	s := t.String()
	return &s
}
