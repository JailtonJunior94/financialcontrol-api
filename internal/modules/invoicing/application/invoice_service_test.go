package invoicingapp_test

import (
	"context"
	"testing"
	"time"

	appdtos "github.com/jailtonjunior94/financialcontrol-api/internal/application/dtos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/domain/entities"
	invoicingapp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/invoicing/application"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeInvoicePublisher struct {
	published []string
	err       error
}

func (f *fakeInvoicePublisher) PublishInvoiceChanged(_ context.Context, invoiceID string) error {
	f.published = append(f.published, invoiceID)
	return f.err
}

type fakeCardRepository struct {
	card *entities.Card
	err  error
}

func (f *fakeCardRepository) GetCardById(_, _ string) (*entities.Card, error) {
	return f.card, f.err
}

type fakeInvoiceRepository struct {
	items            []*entities.InvoiceItem
	itemByID         *entities.InvoiceItem
	getItemErr       error
	getItemsErr      error
	deleteErr        error
	lastControl      int64
	lastControlErr   error
	existingByDate   *entities.Invoice
	getByDateErr     error
	addInvoiceResult *entities.Invoice
	addInvoiceErr    error
	addManyItemsErr  error
}

func (f *fakeInvoiceRepository) DeleteInvoiceItem(_ int64) error { return f.deleteErr }
func (f *fakeInvoiceRepository) GetInvoiceById(_ string) (*entities.Invoice, error) {
	return nil, nil
}
func (f *fakeInvoiceRepository) UpdateManyInvoices(_ []*entities.Invoice) error { return nil }
func (f *fakeInvoiceRepository) GetLastInvoiceControl() (int64, error) {
	return f.lastControl, f.lastControlErr
}
func (f *fakeInvoiceRepository) AddManyInvoiceItems(items []*entities.InvoiceItem) error {
	f.items = items
	return f.addManyItemsErr
}
func (f *fakeInvoiceRepository) GetInvoiceItemById(_ string) (*entities.InvoiceItem, error) {
	return f.itemByID, f.getItemErr
}
func (f *fakeInvoiceRepository) AddInvoice(_ *entities.Invoice) (*entities.Invoice, error) {
	if f.addInvoiceResult != nil {
		return f.addInvoiceResult, f.addInvoiceErr
	}

	return &entities.Invoice{}, f.addInvoiceErr
}
func (f *fakeInvoiceRepository) UpdateInvoice(_ *entities.Invoice) (*entities.Invoice, error) {
	return nil, nil
}
func (f *fakeInvoiceRepository) GetInvoiceByCardId(_, _ string) ([]entities.Invoice, error) {
	return nil, nil
}
func (f *fakeInvoiceRepository) AddInvoiceItem(_ *entities.InvoiceItem) (*entities.InvoiceItem, error) {
	return nil, nil
}
func (f *fakeInvoiceRepository) GetInvoiceItemByInvoiceControl(_ int64) ([]*entities.InvoiceItem, error) {
	return f.items, f.getItemsErr
}
func (f *fakeInvoiceRepository) GetInvoiceByDate(_, _ time.Time, _ string) (*entities.Invoice, error) {
	return f.existingByDate, f.getByDateErr
}
func (f *fakeInvoiceRepository) GetInvoiceItemByInvoiceId(_, _, _ string) ([]entities.InvoiceItem, error) {
	return nil, nil
}
func (f *fakeInvoiceRepository) GetInvoicesCategories(_, _ time.Time, _ string) ([]entities.InvoiceCategories, error) {
	return nil, nil
}
func (f *fakeInvoiceRepository) FetchInvoiceByCard(_ string) ([]appdtos.InvoiceQuery, error) {
	return nil, nil
}
func (f *fakeInvoiceRepository) GetInvoices(_ time.Time) (*appdtos.InvoiceRead, error) {
	return nil, nil
}

func TestDeleteInvoiceItemPublishesEventsThroughPort(t *testing.T) {
	repository := &fakeInvoiceRepository{
		itemByID: &entities.InvoiceItem{InvoiceControl: 10},
		items: []*entities.InvoiceItem{
			{InvoiceId: "inv-1"},
			{InvoiceId: "inv-2"},
		},
	}

	publisher := &fakeInvoicePublisher{}
	service := invoicingapp.NewInvoiceService(&fakeCardRepository{}, repository, publisher)

	response := service.DeleteInvoiceItem("item-1")

	require.NotNil(t, response)
	assert.Equal(t, 204, response.StatusCode)
	assert.Equal(t, []string{"inv-1", "inv-2"}, publisher.published)
}

func TestDeleteInvoiceItemReturnsServerErrorWhenPublisherFails(t *testing.T) {
	repository := &fakeInvoiceRepository{
		itemByID: &entities.InvoiceItem{InvoiceControl: 10},
		items: []*entities.InvoiceItem{
			{InvoiceId: "inv-1"},
		},
	}

	publisher := &fakeInvoicePublisher{err: assert.AnError}
	service := invoicingapp.NewInvoiceService(&fakeCardRepository{}, repository, publisher)

	response := service.DeleteInvoiceItem("item-1")

	require.NotNil(t, response)
	assert.Equal(t, 500, response.StatusCode)
	assert.Equal(t, []string{"inv-1"}, publisher.published)
}

func TestCreateInvoicePublishesEventsThroughPort(t *testing.T) {
	card := &entities.Card{}
	card.ID = "card-1"
	card.ClosingDay = 10

	createdInvoice := &entities.Invoice{}
	createdInvoice.ID = "invoice-1"

	repository := &fakeInvoiceRepository{
		lastControl:      41,
		addInvoiceResult: createdInvoice,
	}

	publisher := &fakeInvoicePublisher{}
	service := invoicingapp.NewInvoiceService(&fakeCardRepository{card: card}, repository, publisher)
	request := invoicingapp.NewInvoiceRequest("2026-04-12 00:00:00", "100.00", "1", "card-1", "category-1", "mercado", "Compra do mes")

	response := service.CreateInvoice("user-1", request)

	require.NotNil(t, response)
	assert.Equal(t, 201, response.StatusCode)
	assert.Equal(t, []string{"invoice-1"}, publisher.published)
	require.Len(t, repository.items, 1)
	assert.Equal(t, "invoice-1", repository.items[0].InvoiceId)
}
