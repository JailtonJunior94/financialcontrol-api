package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/src/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/interfaces"
	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/database"
	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/queries"
)

type InvoiceRepository struct {
	Db database.ISqlConnection
}

func NewInvoiceRepository(db database.ISqlConnection) interfaces.IInvoiceRepository {
	return &InvoiceRepository{Db: db}
}

func (r *InvoiceRepository) GetInvoiceByCardId(userId, cardId string) (invoices []entities.Invoice, err error) {
	connection := r.Db.Connect()
	rows, err := connection.Query(queries.GetInvoiceByCardId, sql.Named("userId", userId), sql.Named("cardId", cardId))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var invoice entities.Invoice
		if err := rows.Scan(&invoice.ID,
			&invoice.CardId,
			&invoice.Date,
			&invoice.Total,
			&invoice.CreatedAt,
			&invoice.UpdatedAt,
			&invoice.Active,
		); err != nil {
			return nil, err
		}

		invoices = append(invoices, invoice)
	}

	return invoices, nil
}

func (r *InvoiceRepository) GetInvoiceByDate(startDate, endDate time.Time, cardId string) (invoice *entities.Invoice, err error) {
	connection := r.Db.Connect()
	row := connection.QueryRow(queries.GetInvoiceByDate, sql.Named("startDate", startDate), sql.Named("endDate", endDate), sql.Named("cardId", cardId))

	i := new(entities.Invoice)
	err = row.Scan(&i.ID, &i.CardId, &i.Date, &i.Total, &i.CreatedAt, &i.UpdatedAt, &i.Active)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return i, nil
}

func (r *InvoiceRepository) AddInvoice(p *entities.Invoice) (invoice *entities.Invoice, err error) {
	s, err := r.Db.OpenConnectionAndMountStatement(queries.AddInvoice)
	if err != nil {
		return nil, err
	}
	defer s.Close()

	result, err := s.Exec(
		sql.Named("id", p.ID),
		sql.Named("cardId", p.CardId),
		sql.Named("date", p.Date),
		sql.Named("total", p.Total),
		sql.Named("createdAt", p.CreatedAt),
		sql.Named("updatedAt", p.UpdatedAt),
		sql.Named("active", p.Active))

	if err := r.Db.ValidateResult(result, err); err != nil {
		return nil, err
	}
	return p, nil
}

func (r *InvoiceRepository) UpdateInvoice(p *entities.Invoice) (invoice *entities.Invoice, err error) {
	s, err := r.Db.OpenConnectionAndMountStatement(queries.UpdateInvoice)
	if err != nil {
		return nil, err
	}
	defer s.Close()

	result, err := s.Exec(sql.Named("total", p.Total), sql.Named("id", p.ID))
	if err := r.Db.ValidateResult(result, err); err != nil {
		return nil, err
	}

	return p, nil
}

func (r *InvoiceRepository) GetInvoiceItemByInvoiceId(invoiceId, cardId, userId string) (items []entities.InvoiceItem, err error) {
	connection := r.Db.Connect()
	rows, err := connection.Query(queries.GetInvoiceItemByInvoiceId, sql.Named("invoiceId", invoiceId), sql.Named("cardId", cardId), sql.Named("userId", userId))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item entities.InvoiceItem
		if err := rows.Scan(
			&item.ID,
			&item.InvoiceId,
			&item.CategoryId,
			&item.PurchaseDate,
			&item.Description,
			&item.TotalAmount,
			&item.Installment,
			&item.InstallmentValue,
			&item.Tags,
			&item.InvoiceControl,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.Active,
			&item.Category.ID,
			&item.Category.Name,
			&item.Category.Active,
		); err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, nil
}

func (r *InvoiceRepository) GetInvoiceItemById(id string) (item *entities.InvoiceItem, err error) {
	connection := r.Db.Connect()
	row := connection.QueryRow(queries.GetInvoiceItemById, sql.Named("id", id))

	i := new(entities.InvoiceItem)
	err = row.Scan(&i.ID,
		&i.InvoiceId,
		&i.CategoryId,
		&i.PurchaseDate,
		&i.Description,
		&i.TotalAmount,
		&i.Installment,
		&i.InstallmentValue,
		&i.Tags,
		&i.InvoiceControl,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Active,
	)

	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows
	}

	if err != nil {
		return nil, err
	}

	return i, nil
}

func (r *InvoiceRepository) AddInvoiceItem(p *entities.InvoiceItem) (invoiceItem *entities.InvoiceItem, err error) {
	s, err := r.Db.OpenConnectionAndMountStatement(queries.AddInvoiceItem)
	if err != nil {
		return nil, err
	}
	defer s.Close()

	result, err := s.Exec(
		sql.Named("id", p.ID),
		sql.Named("invoiceId", p.InvoiceId),
		sql.Named("categoryId", p.CategoryId),
		sql.Named("purchaseDate", p.PurchaseDate),
		sql.Named("description", p.Description),
		sql.Named("totalAmount", p.TotalAmount),
		sql.Named("installment", p.Installment),
		sql.Named("installmentValue", p.InstallmentValue),
		sql.Named("tags", p.Tags),
		sql.Named("createdAt", p.CreatedAt),
		sql.Named("updatedAt", p.UpdatedAt),
		sql.Named("active", p.Active),
		sql.Named("invoiceControl", p.InvoiceControl))

	if err := r.Db.ValidateResult(result, err); err != nil {
		return nil, err
	}
	return p, nil
}

func (r *InvoiceRepository) DeleteInvoiceItem(invoiceControl int64) error {
	s, err := r.Db.OpenConnectionAndMountStatement("DELETE FROM dbo.InvoiceItem WHERE InvoiceControl = @id")
	if err != nil {
		return err
	}
	defer s.Close()

	result, err := s.Exec(sql.Named("id", invoiceControl))
	if err := r.Db.ValidateResult(result, err); err != nil {
		return err
	}

	return nil
}

func (r *InvoiceRepository) GetLastInvoiceControl() (invoiceControl int64, err error) {
	connection := r.Db.Connect()
	row := connection.QueryRow(queries.GetLastInvoiceControl)

	err = row.Scan(&invoiceControl)
	if err == sql.ErrNoRows {
		return 0, nil
	}

	if err != nil {
		return 0, err
	}

	return invoiceControl, nil
}

func (r *InvoiceRepository) GetInvoicesCategories(startDate, endDate time.Time, cardId string) (invoiceCategories []entities.InvoiceCategories, err error) {
	connection := r.Db.Connect()
	if err := connection.Select(&invoiceCategories, queries.GetInvoicesCategories, sql.Named("startDate", startDate), sql.Named("endDate", endDate), sql.Named("cardId", cardId)); err != nil {
		return nil, err
	}
	return invoiceCategories, nil
}

func (r *InvoiceRepository) DeleteAndAddInvoiceItem(invoiceControl int64) error {
	ctx := context.Background()

	tx, err := r.Db.Connect().BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	result, err := tx.ExecContext(ctx, "DELETE FROM dbo.InvoiceItem WHERE InvoiceControl = @invoiceControl", sql.Named("invoiceControl", invoiceControl))
	if err != nil {
		tx.Rollback()
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		tx.Rollback()
		return err
	}

	if rows <= 0 {
		tx.Rollback()
		return err
	}

	ii := entities.NewInvoiceItem("EADA1FB3-D3DC-4D25-903A-D7D385D5DAB1", "4A41F0DB-1F76-44CF-9139-9C52ECED3C3A", "Drogasil (Editado)", "-", time.Now(), 150)

	resultInsert, err := tx.ExecContext(ctx, queries.AddInvoiceItem,
		sql.Named("id", ii.ID),
		sql.Named("invoiceId", ii.InvoiceId),
		sql.Named("categoryId", ii.CategoryId),
		sql.Named("purchaseDate", ii.PurchaseDate),
		sql.Named("description", ii.Description),
		sql.Named("totalAmount", ii.TotalAmount),
		sql.Named("installment", ii.Installment),
		sql.Named("installmentValue", ii.InstallmentValue),
		sql.Named("tags", ii.Tags),
		sql.Named("createdAt", ii.CreatedAt),
		sql.Named("updatedAt", ii.UpdatedAt),
		sql.Named("active", ii.Active),
		sql.Named("invoiceControl", ii.InvoiceControl))

	if err != nil {
		tx.Rollback()
		return err
	}

	rowsInsert, err := resultInsert.RowsAffected()
	if err != nil {
		tx.Rollback()
		return err
	}

	if rowsInsert <= 0 {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
