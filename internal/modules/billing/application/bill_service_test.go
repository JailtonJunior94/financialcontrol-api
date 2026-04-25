package application_test

import (
	"testing"
	"time"

	pkgpersistence "github.com/jailtonjunior94/financialcontrol-api/pkg/persistence"
	billingdomain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/billing/domain"
	billingapp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/billing/application"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeBillRepository struct {
	billByDate      *billingdomain.Bill
	billByDateErr   error
	addedBill       *billingdomain.Bill
	addBillErr      error
	itemByID        *billingdomain.BillItem
	itemByIDErr     error
	updatedItem     *billingdomain.BillItem
	updateItemErr   error
	billByID        *billingdomain.Bill
	billByIDErr     error
	billItems       []billingdomain.BillItem
	billItemsErr    error
	updatedBill     *billingdomain.Bill
	updateBillErr   error
	lastUpdatedItem *billingdomain.BillItem
}

func (f *fakeBillRepository) GetBills() ([]billingdomain.Bill, error) { return nil, nil }
func (f *fakeBillRepository) GetBillById(_ string) (*billingdomain.Bill, error) {
	return f.billByID, f.billByIDErr
}
func (f *fakeBillRepository) GetBillByDate(_, _ time.Time) (*billingdomain.Bill, error) {
	return f.billByDate, f.billByDateErr
}
func (f *fakeBillRepository) AddBill(_ *billingdomain.Bill) (*billingdomain.Bill, error) {
	if f.addedBill != nil {
		return f.addedBill, f.addBillErr
	}

	return &billingdomain.Bill{}, f.addBillErr
}
func (f *fakeBillRepository) UpdateBill(bill *billingdomain.Bill) (*billingdomain.Bill, error) {
	f.updatedBill = bill
	return bill, f.updateBillErr
}
func (f *fakeBillRepository) GetBillItemByBillId(_ string) ([]billingdomain.BillItem, error) {
	return f.billItems, f.billItemsErr
}
func (f *fakeBillRepository) GetBillItemById(_, _ string) (*billingdomain.BillItem, error) {
	return f.itemByID, f.itemByIDErr
}
func (f *fakeBillRepository) AddBillItem(item *billingdomain.BillItem) (*billingdomain.BillItem, error) {
	return item, nil
}
func (f *fakeBillRepository) UpdateBillItem(item *billingdomain.BillItem) (*billingdomain.BillItem, error) {
	f.lastUpdatedItem = item
	if f.updatedItem != nil {
		return f.updatedItem, f.updateItemErr
	}

	return item, f.updateItemErr
}
func (f *fakeBillRepository) Get(_ time.Time) (*pkgpersistence.BillQuery, error) { return nil, nil }

func TestCreateBillRejectsExistingReferenceMonth(t *testing.T) {
	service := billingapp.NewBillService(&fakeBillRepository{
		billByDate: &billingdomain.Bill{},
	})

	response := service.CreateBill(&billingapp.BillRequest{Date: time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC)})

	require.NotNil(t, response)
	assert.Equal(t, 400, response.StatusCode)
}

func TestRemoveBillItemRecalculatesBillTotals(t *testing.T) {
	bill := &billingdomain.Bill{}
	bill.NewBill(time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC))
	bill.ID = "bill-1"

	item := &billingdomain.BillItem{}
	item.NewBillItem("bill-1", "energia", 100)
	item.ID = "item-1"

	repository := &fakeBillRepository{
		itemByID:  item,
		billByID:  bill,
		billItems: []billingdomain.BillItem{},
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

func TestUpdatingValuesCalculatesCorrectPercentages(t *testing.T) {
	tests := []struct {
		name         string
		items        []billingdomain.BillItem
		wantTotal    float64
		wantSixty    float64
		wantForty    float64
	}{
		{
			name:      "empty items",
			items:     []billingdomain.BillItem{},
			wantTotal: 0,
			wantSixty: 0,
			wantForty: 0,
		},
		{
			name: "single item 100",
			items: func() []billingdomain.BillItem {
				item := &billingdomain.BillItem{}
				item.NewBillItem("bill-1", "energia", 100)
				return []billingdomain.BillItem{*item}
			}(),
			wantTotal: 100,
			wantSixty: 60,
			wantForty: 40,
		},
		{
			name: "two items 50 and 50",
			items: func() []billingdomain.BillItem {
				a := &billingdomain.BillItem{}
				a.NewBillItem("bill-1", "agua", 50)
				b := &billingdomain.BillItem{}
				b.NewBillItem("bill-1", "luz", 50)
				return []billingdomain.BillItem{*a, *b}
			}(),
			wantTotal: 100,
			wantSixty: 60,
			wantForty: 40,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bill := &billingdomain.Bill{}
			bill.NewBill(time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC))
			bill.AddBillItems(tt.items)
			bill.UpdatingValues()

			assert.Equal(t, tt.wantTotal, bill.Total)
			assert.Equal(t, tt.wantSixty, bill.SixtyPercent)
			assert.Equal(t, tt.wantForty, bill.FortyPercent)
		})
	}
}
