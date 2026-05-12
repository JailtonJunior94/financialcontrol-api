package mssql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	devkitdb "github.com/JailtonJunior94/devkit-go/pkg/database"
	"github.com/shopspring/decimal"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/ports"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
	pkgdatabase "github.com/jailtonjunior94/financialcontrol-api/pkg/database"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

var _ ports.InstallmentRepository = (*InstallmentRepository)(nil)

// InstallmentRepository implements ports.InstallmentRepository against MSSQL.
type InstallmentRepository struct {
	db devkitdb.DBTX
}

func NewInstallmentRepository(db devkitdb.DBTX) *InstallmentRepository {
	return &InstallmentRepository{db: db}
}

// AddBatch inserts N installments in a single batch (one INSERT per item — RF-34).
func (r *InstallmentRepository) AddBatch(ctx context.Context, items []*entities.Installment) error {
	for _, item := range items {
		if err := r.addOne(ctx, item); err != nil {
			return err
		}
	}
	return nil
}

func (r *InstallmentRepository) addOne(ctx context.Context, item *entities.Installment) error {
	_, err := r.db.ExecContext(ctx, addInstallment,
		sql.Named("id", item.ID().String()),
		sql.Named("transactionId", item.TransactionID().String()),
		sql.Named("invoiceId", item.InvoiceID().String()),
		sql.Named("number", item.Number().Value()),
		sql.Named("total", item.Total().Value()),
		sql.Named("amount", item.Amount().Amount().StringFixed(4)),
		sql.Named("currency", "BRL"),
		sql.Named("status", item.Status().String()),
		sql.Named("legacyOrigin", item.LegacyOrigin()),
		sql.Named("createdAt", item.CreatedAt()),
		sql.Named("updatedAt", item.UpdatedAt()),
		sql.Named("deletedAt", item.DeletedAt()),
	)
	if err != nil {
		return fmt.Errorf("mssql: add installment: %w", err)
	}
	return nil
}

// UpdateBatch updates N installments (one UPDATE per item).
func (r *InstallmentRepository) UpdateBatch(ctx context.Context, items []*entities.Installment) error {
	for _, item := range items {
		if err := r.updateOne(ctx, item); err != nil {
			return err
		}
	}
	return nil
}

func (r *InstallmentRepository) updateOne(ctx context.Context, item *entities.Installment) error {
	_, err := r.db.ExecContext(ctx, updateInstallment,
		sql.Named("invoiceId", item.InvoiceID().String()),
		sql.Named("status", item.Status().String()),
		sql.Named("updatedAt", item.UpdatedAt()),
		sql.Named("deletedAt", item.DeletedAt()),
		sql.Named("id", item.ID().String()),
	)
	if err != nil {
		return fmt.Errorf("mssql: update installment: %w", err)
	}
	return nil
}

func (r *InstallmentRepository) SoftDeleteByTransaction(ctx context.Context, userID identityvo.UserID, transactionID vos.TransactionID, at time.Time) error {
	_, err := r.db.ExecContext(ctx, softDeleteInstallmentsByTransaction,
		sql.Named("deletedAt", at.UTC()),
		sql.Named("updatedAt", at.UTC()),
		sql.Named("transactionId", transactionID.String()),
		sql.Named("userId", userID.String()),
	)
	if err != nil {
		return fmt.Errorf("mssql: soft delete installments by transaction: %w", err)
	}
	return nil
}

func (r *InstallmentRepository) ListByInvoice(ctx context.Context, invoiceID vos.InvoiceID) ([]entities.Installment, error) {
	rows, err := r.db.QueryContext(ctx, listInstallmentsByInvoice,
		sql.Named("invoiceId", invoiceID.String()),
	)
	if err != nil {
		return nil, fmt.Errorf("mssql: list installments by invoice: %w", err)
	}
	return scanInstallmentRows(rows)
}

func (r *InstallmentRepository) ListByTransaction(ctx context.Context, transactionID vos.TransactionID) ([]entities.Installment, error) {
	rows, err := r.db.QueryContext(ctx, listInstallmentsByTransaction,
		sql.Named("transactionId", transactionID.String()),
	)
	if err != nil {
		return nil, fmt.Errorf("mssql: list installments by transaction: %w", err)
	}
	return scanInstallmentRows(rows)
}

func (r *InstallmentRepository) HasClosedOrPaidForTransaction(ctx context.Context, transactionID vos.TransactionID) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, hasClosedOrPaidForTransaction,
		sql.Named("transactionId", transactionID.String()),
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("mssql: has closed or paid for transaction: %w", err)
	}
	return count > 0, nil
}

func (r *InstallmentRepository) SumByMonthCompetence(ctx context.Context, userID identityvo.UserID, period vos.Period) (vos.Money, error) {
	from, to := period.Bounds()
	var raw string
	err := r.db.QueryRowContext(ctx, sumByMonthCompetence,
		sql.Named("userId", userID.String()),
		sql.Named("from", from),
		sql.Named("to", to),
	).Scan(&raw)
	if err != nil {
		return vos.ZeroMoney(), fmt.Errorf("mssql: sum by month competence: %w", err)
	}
	d, err := decimal.NewFromString(raw)
	if err != nil {
		return vos.ZeroMoney(), fmt.Errorf("mssql: sum by month competence parse: %w", err)
	}
	return vos.NewMoneyFromDecimal(d), nil
}

func scanInstallmentRows(rows devkitdb.Rows) ([]entities.Installment, error) {
	return pkgdatabase.ScanAll[entities.Installment](rows, func(r devkitdb.Rows) (entities.Installment, error) {
		var ir InstallmentRow
		if err := r.Scan(
			&ir.ID, &ir.TransactionID, &ir.InvoiceID,
			&ir.Number, &ir.Total, &ir.Amount, &ir.Currency,
			&ir.Status, &ir.LegacyOrigin,
			&ir.CreatedAt, &ir.UpdatedAt, &ir.DeletedAt,
		); err != nil {
			return entities.Installment{}, fmt.Errorf("mssql: list installments scan: %w", err)
		}
		inst, err := RowToInstallment(&ir)
		if err != nil {
			return entities.Installment{}, fmt.Errorf("mssql: list installments map: %w", err)
		}
		return *inst, nil
	})
}
