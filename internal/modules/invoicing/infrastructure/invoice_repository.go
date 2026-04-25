package infrastructure

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	invoicingdomain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/invoicing/domain"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/database"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/persistence"
)

type InvoiceRepository struct {
	db database.ISqlConnection
}

func NewInvoiceRepository(db database.ISqlConnection) *InvoiceRepository {
	return &InvoiceRepository{db: db}
}

func (r *InvoiceRepository) GetInvoiceByCardId(userID, cardID string) ([]invoicingdomain.Invoice, error) {
	connection := r.db.Connect()
	rows, err := connection.Query(getInvoiceByCardID, sql.Named("userId", userID), sql.Named("cardId", cardID))
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	invoices := make([]invoicingdomain.Invoice, 0)
	for rows.Next() {
		var invoice invoicingdomain.Invoice
		if err := rows.Scan(&invoice.ID, &invoice.CardId, &invoice.Date, &invoice.Total, &invoice.CreatedAt, &invoice.UpdatedAt, &invoice.Active); err != nil {
			return nil, err
		}

		invoices = append(invoices, invoice)
	}

	return invoices, nil
}

func (r *InvoiceRepository) GetInvoiceByDate(startDate, endDate time.Time, cardID string) (*invoicingdomain.Invoice, error) {
	row := r.db.Connect().QueryRow(getInvoiceByDate, sql.Named("startDate", startDate), sql.Named("endDate", endDate), sql.Named("cardId", cardID))

	invoice := new(invoicingdomain.Invoice)
	err := row.Scan(&invoice.ID, &invoice.CardId, &invoice.Date, &invoice.Total, &invoice.CreatedAt, &invoice.UpdatedAt, &invoice.Active)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return invoice, nil
}

func (r *InvoiceRepository) AddInvoice(invoice *invoicingdomain.Invoice) (*invoicingdomain.Invoice, error) {
	statement, err := r.db.OpenConnectionAndMountStatement(addInvoice)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = statement.Close()
	}()

	result, err := statement.Exec(
		sql.Named("id", invoice.ID),
		sql.Named("cardId", invoice.CardId),
		sql.Named("date", invoice.Date),
		sql.Named("total", invoice.Total),
		sql.Named("createdAt", invoice.CreatedAt),
		sql.Named("updatedAt", invoice.UpdatedAt),
		sql.Named("active", invoice.Active),
	)
	if err := r.db.ValidateResult(result, err); err != nil {
		return nil, err
	}

	return invoice, nil
}

func (r *InvoiceRepository) UpdateInvoice(invoice *invoicingdomain.Invoice) (*invoicingdomain.Invoice, error) {
	statement, err := r.db.OpenConnectionAndMountStatement(updateInvoice)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = statement.Close()
	}()

	result, err := statement.Exec(sql.Named("total", invoice.Total), sql.Named("id", invoice.ID))
	if err := r.db.ValidateResult(result, err); err != nil {
		return nil, err
	}

	return invoice, nil
}

func (r *InvoiceRepository) GetInvoiceItemByInvoiceId(invoiceID, cardID, userID string) ([]invoicingdomain.InvoiceItem, error) {
	rows, err := r.db.Connect().Query(getInvoiceItemsByInvoiceID, sql.Named("invoiceId", invoiceID), sql.Named("cardId", cardID), sql.Named("userId", userID))
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	items := make([]invoicingdomain.InvoiceItem, 0)
	for rows.Next() {
		var item invoicingdomain.InvoiceItem
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

func (r *InvoiceRepository) GetInvoiceItemById(id string) (*invoicingdomain.InvoiceItem, error) {
	row := r.db.Connect().QueryRow(getInvoiceItemByID, sql.Named("id", id))

	item := new(invoicingdomain.InvoiceItem)
	err := row.Scan(
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
	)
	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}

	return item, nil
}

func (r *InvoiceRepository) AddInvoiceItem(item *invoicingdomain.InvoiceItem) (*invoicingdomain.InvoiceItem, error) {
	statement, err := r.db.OpenConnectionAndMountStatement(addInvoiceItem)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = statement.Close()
	}()

	result, err := statement.Exec(
		sql.Named("id", item.ID),
		sql.Named("invoiceId", item.InvoiceId),
		sql.Named("categoryId", item.CategoryId),
		sql.Named("purchaseDate", item.PurchaseDate),
		sql.Named("description", item.Description),
		sql.Named("totalAmount", item.TotalAmount),
		sql.Named("installment", item.Installment),
		sql.Named("installmentValue", item.InstallmentValue),
		sql.Named("tags", item.Tags),
		sql.Named("createdAt", item.CreatedAt),
		sql.Named("updatedAt", item.UpdatedAt),
		sql.Named("active", item.Active),
		sql.Named("invoiceControl", item.InvoiceControl),
	)
	if err := r.db.ValidateResult(result, err); err != nil {
		return nil, err
	}

	return item, nil
}

func (r *InvoiceRepository) DeleteInvoiceItem(invoiceControl int64) error {
	statement, err := r.db.OpenConnectionAndMountStatement("DELETE FROM dbo.InvoiceItem WHERE InvoiceControl = @id")
	if err != nil {
		return err
	}
	defer func() {
		_ = statement.Close()
	}()

	result, err := statement.Exec(sql.Named("id", invoiceControl))
	return r.db.ValidateResult(result, err)
}

func (r *InvoiceRepository) GetLastInvoiceControl() (int64, error) {
	row := r.db.Connect().QueryRow(getLastControl)

	var control int64
	err := row.Scan(&control)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	return control, nil
}

func (r *InvoiceRepository) GetInvoicesCategories(startDate, endDate time.Time, cardID string) ([]invoicingdomain.InvoiceCategories, error) {
	var categories []invoicingdomain.InvoiceCategories
	if err := r.db.Connect().Select(&categories, getInvoicesCategories, sql.Named("startDate", startDate), sql.Named("endDate", endDate), sql.Named("cardId", cardID)); err != nil {
		return nil, err
	}

	return categories, nil
}

func (r *InvoiceRepository) AddManyInvoiceItems(items []*invoicingdomain.InvoiceItem) error {
	query := make([]string, 0, len(items))
	params := make([]any, 0, len(items)*13)

	for i, item := range items {
		query = append(query, fmt.Sprintf(`INSERT INTO dbo.[InvoiceItem] VALUES (@id%d, @invoiceId%d, @categoryId%d, @purchaseDate%d, @description%d, @totalAmount%d, @installment%d, @installmentValue%d, @tags%d, @createdAt%d, @updatedAt%d, @active%d, @invoiceControl%d)`, i, i, i, i, i, i, i, i, i, i, i, i, i))
		params = append(params,
			sql.Named(fmt.Sprintf("id%d", i), item.ID),
			sql.Named(fmt.Sprintf("invoiceId%d", i), item.InvoiceId),
			sql.Named(fmt.Sprintf("categoryId%d", i), item.CategoryId),
			sql.Named(fmt.Sprintf("purchaseDate%d", i), item.PurchaseDate),
			sql.Named(fmt.Sprintf("description%d", i), item.Description),
			sql.Named(fmt.Sprintf("totalAmount%d", i), item.TotalAmount),
			sql.Named(fmt.Sprintf("installment%d", i), item.Installment),
			sql.Named(fmt.Sprintf("installmentValue%d", i), item.InstallmentValue),
			sql.Named(fmt.Sprintf("tags%d", i), item.Tags),
			sql.Named(fmt.Sprintf("createdAt%d", i), item.CreatedAt),
			sql.Named(fmt.Sprintf("updatedAt%d", i), item.UpdatedAt),
			sql.Named(fmt.Sprintf("active%d", i), item.Active),
			sql.Named(fmt.Sprintf("invoiceControl%d", i), item.InvoiceControl),
		)
	}

	statement, err := r.db.OpenConnectionAndMountStatement(strings.Join(query, " "))
	if err != nil {
		return err
	}
	defer func() {
		_ = statement.Close()
	}()

	result, err := statement.Exec(params...)
	return r.db.ValidateResult(result, err)
}

func (r *InvoiceRepository) GetInvoiceById(id string) (*invoicingdomain.Invoice, error) {
	var invoice invoicingdomain.Invoice
	var invoiceItem invoicingdomain.InvoiceItem
	itemMap := make(map[string][]invoicingdomain.InvoiceItem)

	rows, err := r.db.Connect().Queryx(getInvoiceByID, sql.Named("id", id))
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	for rows.Next() {
		if err := rows.Scan(
			&invoice.ID,
			&invoice.Date,
			&invoice.Total,
			&invoice.CreatedAt,
			&invoice.UpdatedAt,
			&invoice.Active,
			&invoice.Card.ID,
			&invoice.Card.Name,
			&invoice.Card.Description,
			&invoice.Card.UserId,
			&invoice.MarkImportTransactions,
			&invoiceItem.ID,
			&invoiceItem.InvoiceId,
			&invoiceItem.CategoryId,
			&invoiceItem.PurchaseDate,
			&invoiceItem.Description,
			&invoiceItem.TotalAmount,
			&invoiceItem.Installment,
			&invoiceItem.InstallmentValue,
			&invoiceItem.Tags,
			&invoiceItem.InvoiceControl,
			&invoiceItem.CreatedAt,
			&invoiceItem.UpdatedAt,
			&invoiceItem.Active,
			&invoiceItem.Category.ID,
			&invoiceItem.Category.Name,
			&invoiceItem.Category.Active,
		); err != nil {
			return nil, err
		}

		item := invoicingdomain.InvoiceItem{
			InvoiceId:        invoiceItem.InvoiceId,
			CategoryId:       invoiceItem.CategoryId,
			Description:      invoiceItem.Description,
			Tags:             invoiceItem.Tags,
			PurchaseDate:     invoiceItem.PurchaseDate,
			TotalAmount:      invoiceItem.TotalAmount,
			Installment:      invoiceItem.Installment,
			InstallmentValue: invoiceItem.InstallmentValue,
			InvoiceControl:   invoiceItem.InvoiceControl,
			Entity: invoicingdomain.Entity{
				ID:        invoiceItem.ID,
				CreatedAt: invoiceItem.CreatedAt,
				UpdatedAt: invoiceItem.UpdatedAt,
				Active:    invoiceItem.Active,
			},
			Category: invoicingdomain.Category{
				Name: invoiceItem.Category.Name,
				Entity: invoicingdomain.Entity{
					ID:     invoiceItem.Category.ID,
					Active: invoiceItem.Category.Active,
				},
			},
		}

		itemMap[invoice.ID] = append(itemMap[invoice.ID], item)
	}

	invoice.AddInvoiceItems(itemMap[id])
	return &invoice, nil
}

func (r *InvoiceRepository) UpdateManyInvoices(_ []*invoicingdomain.Invoice) error {
	return nil
}

func (r *InvoiceRepository) GetInvoiceItemByInvoiceControl(invoiceControl int64) ([]*invoicingdomain.InvoiceItem, error) {
	var items []*invoicingdomain.InvoiceItem
	if err := r.db.Connect().Select(&items, getInvoiceItemsByControl, sql.Named("invoiceControl", invoiceControl)); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *InvoiceRepository) FetchInvoiceByCard(cardID string) ([]persistence.InvoiceQuery, error) {
	var items []persistence.InvoiceQuery

	query := `SELECT
				i.[Date],
				c.Description,
				i.Total
			FROM
				dbo.Invoice i
				INNER JOIN dbo.Card c ON c.Id = i.CardId
			WHERE c.Id = @cardID
			ORDER BY
				i.[Date]`

	if err := r.db.Connect().Select(&items, query, sql.Named("cardID", cardID)); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *InvoiceRepository) GetInvoices(date time.Time) (*persistence.InvoiceRead, error) {
	query := `SELECT
				CAST(i.Id AS CHAR(36)) [InvoiceID],
				i.[Date],
				i.Total,
				CAST(ii.Id AS CHAR(36)) [InvoiceItemID],
				ii.PurchaseDate,
				ii.Description,
				ii.TotalAmount,
				ii.Installment,
				ii.InstallmentValue,
				ii.Tags,
				CAST(c2.Id AS CHAR(36)) [CategoryID],
				c2.Name
			FROM
				Invoice i
				inner join InvoiceItem ii on ii.InvoiceId = i.Id
				inner join Category c2 on c2.Id = ii.CategoryId
			WHERE
				i.Date = @date
				AND ii.Tags != ''
			ORDER BY ii.PurchaseDate`

	rows, err := r.db.Connect().Query(query, sql.Named("date", date))
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var invoice persistence.InvoiceRead
	var item persistence.InvoiceItemRead
	itemMap := make(map[string][]persistence.InvoiceItemRead)

	for rows.Next() {
		if err := rows.Scan(
			&invoice.ID,
			&invoice.Date,
			&invoice.Total,
			&item.ID,
			&item.PurchaseDate,
			&item.Description,
			&item.TotalAmount,
			&item.Installment,
			&item.InstallmentValue,
			&item.Tags,
			&item.Category.ID,
			&item.Category.Name,
		); err != nil {
			return nil, err
		}

		itemMap[invoice.ID] = append(itemMap[invoice.ID], item)
	}

	invoice.Items = itemMap[invoice.ID]
	return &invoice, nil
}
