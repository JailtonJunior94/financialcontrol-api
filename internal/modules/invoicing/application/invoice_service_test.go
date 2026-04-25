package invoicingapp_test

import (
	"context"
	"testing"
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/internal/application/dtos"
	invoicingapp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/invoicing/application"
	invoicingdomain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/invoicing/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeInvoicePublisher struct {
	published []invoicingdomain.InvoiceChangedPayload
	err       error
}

func (f *fakeInvoicePublisher) PublishInvoiceChanged(_ context.Context, payload invoicingdomain.InvoiceChangedPayload) error {
	f.published = append(f.published, payload)
	return f.err
}

type fakeCardRepository struct {
	card *invoicingapp.CardView
	err  error
}

func (f *fakeCardRepository) GetCardById(_, _ string) (*invoicingapp.CardView, error) {
	return f.card, f.err
}

type fakeInvoiceRepository struct {
	items            []*invoicingdomain.InvoiceItem
	itemByID         *invoicingdomain.InvoiceItem
	getItemErr       error
	getItemsErr      error
	deleteErr        error
	lastControl      int64
	lastControlErr   error
	existingByDate   *invoicingdomain.Invoice
	getByDateErr     error
	addInvoiceResult *invoicingdomain.Invoice
	addInvoiceErr    error
	addManyItemsErr  error
	getByIDResult    *invoicingdomain.Invoice
	getByIDErr       error
	updateErr        error
}

func (f *fakeInvoiceRepository) DeleteInvoiceItem(_ int64) error { return f.deleteErr }
func (f *fakeInvoiceRepository) GetInvoiceById(_ string) (*invoicingdomain.Invoice, error) {
	return f.getByIDResult, f.getByIDErr
}
func (f *fakeInvoiceRepository) UpdateManyInvoices(_ []*invoicingdomain.Invoice) error { return nil }
func (f *fakeInvoiceRepository) GetLastInvoiceControl() (int64, error) {
	return f.lastControl, f.lastControlErr
}
func (f *fakeInvoiceRepository) AddManyInvoiceItems(items []*invoicingdomain.InvoiceItem) error {
	f.items = items
	return f.addManyItemsErr
}
func (f *fakeInvoiceRepository) GetInvoiceItemById(_ string) (*invoicingdomain.InvoiceItem, error) {
	return f.itemByID, f.getItemErr
}
func (f *fakeInvoiceRepository) AddInvoice(_ *invoicingdomain.Invoice) (*invoicingdomain.Invoice, error) {
	if f.addInvoiceResult != nil {
		return f.addInvoiceResult, f.addInvoiceErr
	}
	return &invoicingdomain.Invoice{}, f.addInvoiceErr
}
func (f *fakeInvoiceRepository) UpdateInvoice(inv *invoicingdomain.Invoice) (*invoicingdomain.Invoice, error) {
	return inv, f.updateErr
}
func (f *fakeInvoiceRepository) GetInvoiceByCardId(_, _ string) ([]invoicingdomain.Invoice, error) {
	return nil, nil
}
func (f *fakeInvoiceRepository) AddInvoiceItem(_ *invoicingdomain.InvoiceItem) (*invoicingdomain.InvoiceItem, error) {
	return nil, nil
}
func (f *fakeInvoiceRepository) GetInvoiceItemByInvoiceControl(_ int64) ([]*invoicingdomain.InvoiceItem, error) {
	return f.items, f.getItemsErr
}
func (f *fakeInvoiceRepository) GetInvoiceByDate(_, _ time.Time, _ string) (*invoicingdomain.Invoice, error) {
	return f.existingByDate, f.getByDateErr
}
func (f *fakeInvoiceRepository) GetInvoiceItemByInvoiceId(_, _, _ string) ([]invoicingdomain.InvoiceItem, error) {
	return nil, nil
}
func (f *fakeInvoiceRepository) GetInvoicesCategories(_, _ time.Time, _ string) ([]invoicingdomain.InvoiceCategories, error) {
	return nil, nil
}
func (f *fakeInvoiceRepository) FetchInvoiceByCard(_ string) ([]dtos.InvoiceQuery, error) {
	return nil, nil
}
func (f *fakeInvoiceRepository) GetInvoices(_ time.Time) (*dtos.InvoiceRead, error) { return nil, nil }

// flaggedInvoice returns a minimal invoice that triggers the publish path.
func flaggedInvoice(id string) *invoicingdomain.Invoice {
	inv := &invoicingdomain.Invoice{MarkImportTransactions: true}
	inv.ID = id
	inv.Card.Description = "Visa"
	inv.Card.UserId = "user-1"
	return inv
}

func TestDeleteInvoiceItemPublishesEventsThroughPort(t *testing.T) {
	repository := &fakeInvoiceRepository{
		itemByID: &invoicingdomain.InvoiceItem{InvoiceControl: 10},
		items: []*invoicingdomain.InvoiceItem{
			{InvoiceId: "inv-1"},
			{InvoiceId: "inv-2"},
		},
		getByIDResult: flaggedInvoice(""),
	}

	publisher := &fakeInvoicePublisher{}
	service := invoicingapp.NewInvoiceService(&fakeCardRepository{}, repository, publisher)

	response := service.DeleteInvoiceItem("item-1")

	require.NotNil(t, response)
	assert.Equal(t, 204, response.StatusCode)
	require.Len(t, publisher.published, 2)
	assert.Equal(t, "inv-1", publisher.published[0].InvoiceID)
	assert.Equal(t, "inv-2", publisher.published[1].InvoiceID)
}

func TestDeleteInvoiceItemReturnsServerErrorWhenPublisherFails(t *testing.T) {
	repository := &fakeInvoiceRepository{
		itemByID: &invoicingdomain.InvoiceItem{InvoiceControl: 10},
		items: []*invoicingdomain.InvoiceItem{
			{InvoiceId: "inv-1"},
		},
		getByIDResult: flaggedInvoice("inv-1"),
	}

	publisher := &fakeInvoicePublisher{err: assert.AnError}
	service := invoicingapp.NewInvoiceService(&fakeCardRepository{}, repository, publisher)

	response := service.DeleteInvoiceItem("item-1")

	require.NotNil(t, response)
	assert.Equal(t, 500, response.StatusCode)
	require.Len(t, publisher.published, 1)
	assert.Equal(t, "inv-1", publisher.published[0].InvoiceID)
}

func TestCreateInvoicePublishesEventsThroughPort(t *testing.T) {
	card := &invoicingapp.CardView{ID: "card-1", ClosingDay: 10}

	createdInvoice := &invoicingdomain.Invoice{}
	createdInvoice.ID = "invoice-1"

	repository := &fakeInvoiceRepository{
		lastControl:      41,
		addInvoiceResult: createdInvoice,
		getByIDResult:    flaggedInvoice("invoice-1"),
	}

	publisher := &fakeInvoicePublisher{}
	service := invoicingapp.NewInvoiceService(&fakeCardRepository{card: card}, repository, publisher)
	request := invoicingapp.NewInvoiceRequest("2026-04-12 00:00:00", "100.00", "1", "card-1", "category-1", "mercado", "Compra do mes")

	response := service.CreateInvoice("user-1", request)

	require.NotNil(t, response)
	assert.Equal(t, 201, response.StatusCode)
	require.Len(t, publisher.published, 1)
	assert.Equal(t, "invoice-1", publisher.published[0].InvoiceID)
	require.Len(t, repository.items, 1)
	assert.Equal(t, "invoice-1", repository.items[0].InvoiceId)
}
