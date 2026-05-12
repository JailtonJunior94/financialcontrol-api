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
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/projections"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
	pkgdatabase "github.com/jailtonjunior94/financialcontrol-api/pkg/database"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

var _ ports.InvoiceRepository = (*InvoiceRepository)(nil)

// InvoiceRepository implements ports.InvoiceRepository against MSSQL.
type InvoiceRepository struct {
	db devkitdb.DBTX
}

func NewInvoiceRepository(db devkitdb.DBTX) *InvoiceRepository {
	return &InvoiceRepository{db: db}
}

// AssignOrCreateOpen atomically finds or creates the open invoice for the given
// card and billing cycle derived from occurredAtLocal (RF-16, UPDLOCK guard).
func (r *InvoiceRepository) AssignOrCreateOpen(
	ctx context.Context,
	userID identityvo.UserID,
	cardID vos.CardID,
	occurredAtLocal time.Time,
	card projections.CardView,
	clock ports.Clock,
) (*entities.Invoice, error) {
	saoPaulo, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		return nil, fmt.Errorf("mssql: assign or create open: load tz: %w", err)
	}
	targetClosing := computeTargetClosingDate(occurredAtLocal.In(saoPaulo), card.ClosingDay, saoPaulo)

	// 1. try to find an existing open invoice (UPDLOCK + HOLDLOCK)
	existing, err := r.scanInvoice(r.db.QueryRowContext(ctx, getOpenInvoiceForCardClosing,
		sql.Named("userId", userID.String()),
		sql.Named("cardId", cardID.String()),
		sql.Named("closingDate", targetClosing.UTC()),
	))
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("mssql: assign or create open: %w", err)
	}
	if existing != nil {
		return existing, nil
	}

	// 2. create a new open invoice for this cycle
	cycleStart, cycleEnd, closingDate, dueDate := computeInvoiceDates(targetClosing, card, saoPaulo)
	inv, err := entities.NewInvoice(card, cycleStart, cycleEnd, closingDate, dueDate, clock.Now())
	if err != nil {
		return nil, fmt.Errorf("mssql: assign or create open: new invoice: %w", err)
	}
	if addErr := r.insertInvoice(ctx, inv); addErr != nil {
		return nil, fmt.Errorf("mssql: assign or create open: insert: %w", addErr)
	}
	return inv, nil
}

func (r *InvoiceRepository) GetByID(ctx context.Context, userID identityvo.UserID, id vos.InvoiceID) (*entities.Invoice, error) {
	inv, err := r.scanInvoice(r.db.QueryRowContext(ctx, getInvoiceByID,
		sql.Named("userId", userID.String()),
		sql.Named("id", id.String()),
	))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrInvoiceNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("mssql: get invoice by id: %w", err)
	}
	return inv, nil
}

func (r *InvoiceRepository) List(ctx context.Context, userID identityvo.UserID, f filters.InvoiceFilter) ([]entities.Invoice, int64, error) {
	where, args := buildInvoiceWhere(userID, f)

	var total int64
	if err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(1) FROM dbo.FinanceInvoices (NOLOCK) WHERE "+where, args...,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("mssql: list invoices count: %w", err)
	}

	dataQ := invoiceListSelect + " WHERE " + where + invoiceListOrder + invoiceListPaging
	args = append(args,
		sql.Named("offset", f.Pagination.Offset()),
		sql.Named("size", f.Pagination.PageSize()),
	)

	rows, err := r.db.QueryContext(ctx, dataQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("mssql: list invoices: %w", err)
	}

	items, err := pkgdatabase.ScanAll[entities.Invoice](rows, func(row devkitdb.Rows) (entities.Invoice, error) {
		var ir InvoiceRow
		if scanErr := row.Scan(
			&ir.ID, &ir.UserID, &ir.CardID, &ir.State,
			&ir.CycleStart, &ir.CycleEnd, &ir.ClosingDate, &ir.DueDate,
			&ir.Total, &ir.Currency, &ir.PaidAt, &ir.LegacyOrigin,
			&ir.CreatedAt, &ir.UpdatedAt, &ir.DeletedAt,
		); scanErr != nil {
			return entities.Invoice{}, fmt.Errorf("mssql: list invoices scan: %w", scanErr)
		}
		inv, mapErr := RowToInvoice(&ir)
		if mapErr != nil {
			return entities.Invoice{}, fmt.Errorf("mssql: list invoices map: %w", mapErr)
		}
		return *inv, nil
	})
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *InvoiceRepository) Update(ctx context.Context, inv *entities.Invoice) error {
	_, err := r.db.ExecContext(ctx, updateInvoice,
		sql.Named("state", inv.State().String()),
		sql.Named("total", inv.Total().Amount().StringFixed(4)),
		sql.Named("paidAt", inv.PaidAt()),
		sql.Named("updatedAt", inv.UpdatedAt()),
		sql.Named("deletedAt", inv.DeletedAt()),
		sql.Named("id", inv.ID().String()),
		sql.Named("userId", inv.UserID().String()),
	)
	if err != nil {
		return fmt.Errorf("mssql: update invoice: %w", err)
	}
	return nil
}

func (r *InvoiceRepository) NextOpenFor(ctx context.Context, userID identityvo.UserID, cardID vos.CardID, _ ports.Clock) (*entities.Invoice, error) {
	inv, err := r.scanInvoice(r.db.QueryRowContext(ctx, nextOpenInvoiceForCard,
		sql.Named("userId", userID.String()),
		sql.Named("cardId", cardID.String()),
	))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("mssql: next open invoice for card: %w", err)
	}
	return inv, nil
}

func (r *InvoiceRepository) SumPaidInPeriod(ctx context.Context, userID identityvo.UserID, period vos.Period) (vos.Money, error) {
	from, to := period.Bounds()
	var raw string
	if err := r.db.QueryRowContext(ctx, sumPaidInPeriod,
		sql.Named("userId", userID.String()),
		sql.Named("from", from),
		sql.Named("to", to),
	).Scan(&raw); err != nil {
		return vos.ZeroMoney(), fmt.Errorf("mssql: sum paid in period: %w", err)
	}
	d, err := decimal.NewFromString(raw)
	if err != nil {
		return vos.ZeroMoney(), fmt.Errorf("mssql: sum paid in period parse: %w", err)
	}
	return vos.NewMoneyFromDecimal(d), nil
}

func (r *InvoiceRepository) SumOpenForUser(ctx context.Context, userID identityvo.UserID, period vos.Period) (vos.Money, error) {
	from, to := period.Bounds()
	var raw string
	if err := r.db.QueryRowContext(ctx, sumOpenForUser,
		sql.Named("userId", userID.String()),
		sql.Named("from", from),
		sql.Named("to", to),
	).Scan(&raw); err != nil {
		return vos.ZeroMoney(), fmt.Errorf("mssql: sum open for user: %w", err)
	}
	d, err := decimal.NewFromString(raw)
	if err != nil {
		return vos.ZeroMoney(), fmt.Errorf("mssql: sum open for user parse: %w", err)
	}
	return vos.NewMoneyFromDecimal(d), nil
}

// scanInvoice scans a single Row into an InvoiceRow and maps it to an Invoice entity.
func (r *InvoiceRepository) scanInvoice(row devkitdb.Row) (*entities.Invoice, error) {
	var ir InvoiceRow
	if err := row.Scan(
		&ir.ID, &ir.UserID, &ir.CardID, &ir.State,
		&ir.CycleStart, &ir.CycleEnd, &ir.ClosingDate, &ir.DueDate,
		&ir.Total, &ir.Currency, &ir.PaidAt, &ir.LegacyOrigin,
		&ir.CreatedAt, &ir.UpdatedAt, &ir.DeletedAt,
	); err != nil {
		return nil, err
	}
	return RowToInvoice(&ir)
}

// insertInvoice persists a new Invoice record.
func (r *InvoiceRepository) insertInvoice(ctx context.Context, inv *entities.Invoice) error {
	_, err := r.db.ExecContext(ctx, addInvoice,
		sql.Named("id", inv.ID().String()),
		sql.Named("userId", inv.UserID().String()),
		sql.Named("cardId", inv.CardID().String()),
		sql.Named("state", inv.State().String()),
		sql.Named("cycleStart", inv.CycleStart()),
		sql.Named("cycleEnd", inv.CycleEnd()),
		sql.Named("closingDate", inv.ClosingDate()),
		sql.Named("dueDate", inv.DueDate()),
		sql.Named("total", inv.Total().Amount().StringFixed(4)),
		sql.Named("currency", "BRL"),
		sql.Named("paidAt", inv.PaidAt()),
		sql.Named("legacyOrigin", inv.LegacyOrigin()),
		sql.Named("createdAt", inv.CreatedAt()),
		sql.Named("updatedAt", inv.UpdatedAt()),
		sql.Named("deletedAt", inv.DeletedAt()),
	)
	return err
}

// buildInvoiceWhere constructs the WHERE clause and named args for invoice listing.
func buildInvoiceWhere(userID identityvo.UserID, f filters.InvoiceFilter) (string, []any) {
	var sb strings.Builder
	args := make([]any, 0, 6)

	sb.WriteString("[UserId] = @userId AND [DeletedAt] IS NULL")
	args = append(args, sql.Named("userId", userID.String()))

	if f.CardID != nil {
		sb.WriteString(" AND [CardId] = @cardId")
		args = append(args, sql.Named("cardId", f.CardID.String()))
	}
	if f.State != nil {
		sb.WriteString(" AND [State] = @state")
		args = append(args, sql.Named("state", f.State.String()))
	}
	if f.From != nil {
		sb.WriteString(" AND [CycleEnd] >= @from")
		args = append(args, sql.Named("from", *f.From))
	}
	if f.To != nil {
		sb.WriteString(" AND [CycleEnd] < @to")
		args = append(args, sql.Named("to", *f.To))
	}
	return sb.String(), args
}

// computeTargetClosingDate returns the target closing date in saoPaulo location
// for the billing cycle that spTime belongs to.
func computeTargetClosingDate(spTime time.Time, closingDay int, loc *time.Location) time.Time {
	day := spTime.Day()
	year := spTime.Year()
	month := spTime.Month()

	if day >= closingDay {
		month++
		if month > 12 {
			month = 1
			year++
		}
	}
	maxDay := daysInMonthLoc(year, month, loc)
	if closingDay > maxDay {
		closingDay = maxDay
	}
	return time.Date(year, month, closingDay, 0, 0, 0, 0, loc)
}

// computeInvoiceDates derives cycleStart, cycleEnd, closingDate and dueDate from
// the target closing date and the card's billing configuration.
func computeInvoiceDates(targetClosing time.Time, card projections.CardView, loc *time.Location) (cycleStart, cycleEnd, closingDate, dueDate time.Time) {
	closingDate = targetClosing.UTC()
	cycleEnd = closingDate

	// cycleStart: previous month's closing day in SP, then converted to UTC.
	spClosing := targetClosing.In(loc)
	prevYear, prevMonth := spClosing.Year(), spClosing.Month()-1
	if prevMonth < 1 {
		prevMonth = 12
		prevYear--
	}
	prevMaxDay := daysInMonthLoc(prevYear, prevMonth, loc)
	prevClosingDay := card.ClosingDay
	if prevClosingDay > prevMaxDay {
		prevClosingDay = prevMaxDay
	}
	cycleStart = time.Date(prevYear, prevMonth, prevClosingDay, 0, 0, 0, 0, loc).UTC()

	// dueDate: DueDay in same month as closingDate when DueDay > ClosingDay, else next month.
	closingYear, closingMonth := spClosing.Year(), spClosing.Month()
	if card.DueDay > card.ClosingDay {
		maxDue := daysInMonthLoc(closingYear, closingMonth, loc)
		dueDay := card.DueDay
		if dueDay > maxDue {
			dueDay = maxDue
		}
		dueDate = time.Date(closingYear, closingMonth, dueDay, 0, 0, 0, 0, loc).UTC()
	} else {
		nextMonth := closingMonth + 1
		nextYear := closingYear
		if nextMonth > 12 {
			nextMonth = 1
			nextYear++
		}
		maxDue := daysInMonthLoc(nextYear, nextMonth, loc)
		dueDay := card.DueDay
		if dueDay > maxDue {
			dueDay = maxDue
		}
		dueDate = time.Date(nextYear, nextMonth, dueDay, 0, 0, 0, 0, loc).UTC()
	}
	return
}

func daysInMonthLoc(year int, month time.Month, loc *time.Location) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, loc).Day()
}
