package infrastructure

import (
	"context"
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/planning"
	"github.com/jailtonjunior94/financialcontrol-api/internal/platform/persistence"
)

// invoiceDataSource is the minimal interface the invoicing adapter requires.
// Defined on the consumer side so planning never imports the concrete
// invoicing repository or the invoicing application package.
type invoiceDataSource interface {
	GetInvoices(date time.Time) (*persistence.InvoiceRead, error)
}

// InvoicingReadAdapter implements planning.InvoicingReadPort by querying the
// invoice repository and projecting the result into the planning read model.
//
// The "+1 month" offset applied to referenceDate reflects how invoice cycles
// work: a reference month of April 2026 queries invoices closing in May 2026.
type InvoicingReadAdapter struct {
	repo invoiceDataSource
}

func NewInvoicingReadAdapter(repo invoiceDataSource) *InvoicingReadAdapter {
	return &InvoicingReadAdapter{repo: repo}
}

func (a *InvoicingReadAdapter) GetMonthlyInvoices(_ context.Context, referenceDate time.Time) (*planning.MonthlyInvoicesReadModel, error) {
	query, err := a.repo.GetInvoices(referenceDate.AddDate(0, 1, 0))
	if err != nil {
		return nil, err
	}

	model := &planning.MonthlyInvoicesReadModel{ReferenceDate: referenceDate}
	if query == nil {
		return model, nil
	}

	model.Items = make([]planning.InvoiceReadItem, 0, len(query.Items))
	for _, i := range query.Items {
		model.Items = append(model.Items, planning.InvoiceReadItem{
			ID:          i.ID,
			Description: i.Description,
			Total:       i.InstallmentValue,
			Date:        i.PurchaseDate,
			Tags:        i.Tags,
			Category:    i.Category.Name,
		})
	}

	return model, nil
}
