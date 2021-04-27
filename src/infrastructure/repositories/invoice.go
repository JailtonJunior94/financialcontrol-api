package repositories

import (
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

func (r *InvoiceRepository) GetInvoiceItemByInvoiceId(invoiceId string) (items []entities.InvoiceItem, err error) {
	connection := r.Db.Connect()
	rows, err := connection.Query(queries.GetInvoiceItemByInvoiceId, sql.Named("invoiceId", invoiceId))
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
		sql.Named("active", p.Active))

	if err := r.Db.ValidateResult(result, err); err != nil {
		return nil, err
	}
	return p, nil
}
