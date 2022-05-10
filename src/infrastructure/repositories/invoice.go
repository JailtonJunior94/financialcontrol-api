package repositories

import (
	"database/sql"
	"fmt"
	"strings"
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

func (r *InvoiceRepository) AddManyInvoiceItems(invoiceItems []*entities.InvoiceItem) error {
	query := []string{}
	params := []interface{}{}

	for i, item := range invoiceItems {
		query = append(query, fmt.Sprintf(`INSERT INTO dbo.[InvoiceItem] VALUES (@id%d, @invoiceId%d, @categoryId%d, @purchaseDate%d, @description%d, @totalAmount%d, @installment%d, @installmentValue%d, @tags%d,@createdAt%d, @updatedAt%d, @active%d, @invoiceControl%d)`, i, i, i, i, i, i, i, i, i, i, i, i, i))
		params = append(params, sql.Named(fmt.Sprintf("id%d", i), item.ID))
		params = append(params, sql.Named(fmt.Sprintf("invoiceId%d", i), item.InvoiceId))
		params = append(params, sql.Named(fmt.Sprintf("categoryId%d", i), item.CategoryId))
		params = append(params, sql.Named(fmt.Sprintf("purchaseDate%d", i), item.PurchaseDate))
		params = append(params, sql.Named(fmt.Sprintf("description%d", i), item.Description))
		params = append(params, sql.Named(fmt.Sprintf("totalAmount%d", i), item.TotalAmount))
		params = append(params, sql.Named(fmt.Sprintf("installment%d", i), item.Installment))
		params = append(params, sql.Named(fmt.Sprintf("installmentValue%d", i), item.InstallmentValue))
		params = append(params, sql.Named(fmt.Sprintf("tags%d", i), item.Tags))
		params = append(params, sql.Named(fmt.Sprintf("createdAt%d", i), item.CreatedAt))
		params = append(params, sql.Named(fmt.Sprintf("updatedAt%d", i), item.UpdatedAt))
		params = append(params, sql.Named(fmt.Sprintf("active%d", i), item.Active))
		params = append(params, sql.Named(fmt.Sprintf("invoiceControl%d", i), item.InvoiceControl))
	}

	s, err := r.Db.OpenConnectionAndMountStatement(strings.Join(query, " "))
	if err != nil {
		return err
	}
	defer s.Close()

	result, err := s.Exec(params...)
	if err := r.Db.ValidateResult(result, err); err != nil {
		return err
	}
	return nil
}

func (r *InvoiceRepository) GetInvoiceById(id string) (*entities.Invoice, error) {
	var i entities.Invoice
	var ii entities.InvoiceItem
	var invoiceItems = make(map[string][]entities.InvoiceItem)

	connection := r.Db.Connect()
	rows, err := connection.Queryx(queries.GetInvoiceById, sql.Named("id", id))
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		if err := rows.Scan(
			&i.ID,
			&i.Date,
			&i.Total,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Active,
			&i.Card.ID,
			&i.Card.Name,
			&ii.ID,
			&ii.InvoiceId,
			&ii.CategoryId,
			&ii.PurchaseDate,
			&ii.Description,
			&ii.TotalAmount,
			&ii.Installment,
			&ii.InstallmentValue,
			&ii.Tags,
			&ii.InvoiceControl,
			&ii.CreatedAt,
			&ii.UpdatedAt,
			&ii.Active,
			&ii.Category.ID,
			&ii.Category.Name,
			&ii.Category.Active,
		); err != nil {
			return nil, err
		}

		if items, ok := invoiceItems[i.ID]; ok {
			item := entities.InvoiceItem{
				InvoiceId:        ii.InvoiceId,
				CategoryId:       ii.CategoryId,
				Description:      ii.Description,
				Tags:             ii.Tags,
				PurchaseDate:     ii.PurchaseDate,
				TotalAmount:      ii.TotalAmount,
				Installment:      ii.Installment,
				InstallmentValue: ii.InstallmentValue,
				InvoiceControl:   ii.InvoiceControl,
				Entity: entities.Entity{
					ID:        ii.ID,
					CreatedAt: ii.CreatedAt,
					UpdatedAt: ii.UpdatedAt,
					Active:    ii.Active,
				},
				Category: entities.Category{
					Name: ii.Category.Name,
					Entity: entities.Entity{
						ID:     ii.Category.ID,
						Active: ii.Category.Active,
					},
				},
			}
			invoiceItems[i.ID] = append(items, item)
		} else {
			item := entities.InvoiceItem{
				InvoiceId:        ii.InvoiceId,
				CategoryId:       ii.CategoryId,
				Description:      ii.Description,
				Tags:             ii.Tags,
				PurchaseDate:     ii.PurchaseDate,
				TotalAmount:      ii.TotalAmount,
				Installment:      ii.Installment,
				InstallmentValue: ii.InstallmentValue,
				InvoiceControl:   ii.InvoiceControl,
				Entity: entities.Entity{
					ID:        ii.ID,
					CreatedAt: ii.CreatedAt,
					UpdatedAt: ii.UpdatedAt,
					Active:    ii.Active,
				},
				Category: entities.Category{
					Name: ii.Category.Name,
					Entity: entities.Entity{
						ID:     ii.Category.ID,
						Active: ii.Category.Active,
					},
				},
			}
			invoiceItems[i.ID] = []entities.InvoiceItem{item}
		}
	}

	i.AddInvoiceItems(invoiceItems[id])
	return &i, nil
}

func (r *InvoiceRepository) UpdateManyInvoices(invoices []*entities.Invoice) error {
	return nil
}

func (r *InvoiceRepository) GetInvoiceItemByInvoiceControl(invoiceControl int64) ([]*entities.InvoiceItem, error) {
	var items []*entities.InvoiceItem

	connection := r.Db.Connect()
	if err := connection.Select(&items, queries.GetInvoiceItemByInvoiceControl, sql.Named("invoiceControl", invoiceControl)); err != nil {
		return nil, err
	}

	return items, nil
}
