package application_test

import (
	"testing"
	"time"

	appdtos "github.com/jailtonjunior94/financialcontrol-api/internal/application/dtos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/domain/entities"
	billingapp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/billing/application"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeBillRepository struct {
	billByDate      *entities.Bill
	billByDateErr   error
	addedBill       *entities.Bill
	addBillErr      error
	itemByID        *entities.BillItem
	itemByIDErr     error
	updatedItem     *entities.BillItem
	updateItemErr   error
	billByID        *entities.Bill
	billByIDErr     error
	billItems       []entities.BillItem
	billItemsErr    error
	updatedBill     *entities.Bill
	updateBillErr   error
	lastUpdatedItem *entities.BillItem
}

func (f *fakeBillRepository) GetBills() ([]entities.Bill, error) { return nil, nil }
func (f *fakeBillRepository) GetBillById(_ string) (*entities.Bill, error) {
	return f.billByID, f.billByIDErr
}
func (f *fakeBillRepository) GetBillByDate(_, _ time.Time) (*entities.Bill, error) {
	return f.billByDate, f.billByDateErr
}
func (f *fakeBillRepository) AddBill(_ *entities.Bill) (*entities.Bill, error) {
	if f.addedBill != nil {
		return f.addedBill, f.addBillErr
	}

	return &entities.Bill{}, f.addBillErr
}
func (f *fakeBillRepository) UpdateBill(bill *entities.Bill) (*entities.Bill, error) {
	f.updatedBill = bill
	return bill, f.updateBillErr
}
func (f *fakeBillRepository) GetBillItemByBillId(_ string) ([]entities.BillItem, error) {
	return f.billItems, f.billItemsErr
}
func (f *fakeBillRepository) GetBillItemById(_, _ string) (*entities.BillItem, error) {
	return f.itemByID, f.itemByIDErr
}
func (f *fakeBillRepository) AddBillItem(item *entities.BillItem) (*entities.BillItem, error) {
	return item, nil
}
func (f *fakeBillRepository) UpdateBillItem(item *entities.BillItem) (*entities.BillItem, error) {
	f.lastUpdatedItem = item
	if f.updatedItem != nil {
		return f.updatedItem, f.updateItemErr
	}

	return item, f.updateItemErr
}
func (f *fakeBillRepository) Get(_ time.Time) (*appdtos.BillQuery, error) { return nil, nil }

func TestCreateBillRejectsExistingReferenceMonth(t *testing.T) {
	service := billingapp.NewBillService(&fakeBillRepository{
		billByDate: &entities.Bill{},
	})

	response := service.CreateBill(&billingapp.BillRequest{Date: time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC)})

	require.NotNil(t, response)
	assert.Equal(t, 400, response.StatusCode)
}

func TestRemoveBillItemRecalculatesBillTotals(t *testing.T) {
	bill := &entities.Bill{}
	bill.NewBill(time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC))
	bill.ID = "bill-1"

	item := &entities.BillItem{}
	item.NewBillItem("bill-1", "energia", 100)
	item.ID = "item-1"

	repository := &fakeBillRepository{
		itemByID:  item,
		billByID:  bill,
		billItems: []entities.BillItem{},
	}

	service := billingapp.NewBillService(repository)
	response := service.RemoveBillItem("bill-1", "item-1")

	require.NotNil(t, response)
	assert.Equal(t, 204, response.StatusCode)
	require.NotNil(t, repository.lastUpdatedItem)
	assert.False(t, repository.lastUpdatedItem.Active)
	require.NotNil(t, repository.updatedBill)
	assert.Equal(t, 0.0, repository.updatedBill.Total)
}
