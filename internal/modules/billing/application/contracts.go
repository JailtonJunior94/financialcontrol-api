package application

import (
	"time"

	billingdomain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/billing/domain"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/persistence"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/web"
)

type HttpResponse = web.HttpResponse

type BillRepository interface {
	GetBills() ([]billingdomain.Bill, error)
	GetBillById(id string) (*billingdomain.Bill, error)
	GetBillByDate(startDate, endDate time.Time) (*billingdomain.Bill, error)
	AddBill(bill *billingdomain.Bill) (*billingdomain.Bill, error)
	UpdateBill(bill *billingdomain.Bill) (*billingdomain.Bill, error)
	GetBillItemByBillId(billID string) ([]billingdomain.BillItem, error)
	GetBillItemById(id, billID string) (*billingdomain.BillItem, error)
	AddBillItem(item *billingdomain.BillItem) (*billingdomain.BillItem, error)
	UpdateBillItem(item *billingdomain.BillItem) (*billingdomain.BillItem, error)
	Get(date time.Time) (*persistence.BillQuery, error)
}

type BillService interface {
	Bills() *HttpResponse
	BillById(id string) *HttpResponse
	CreateBill(request *BillRequest) *HttpResponse
	BillItemById(id, billID string) *HttpResponse
	CreateBillItem(request *BillItemRequest, billID string) *HttpResponse
	UpdateBillItem(billID, id string, request *BillItemRequest) *HttpResponse
	RemoveBillItem(billID, id string) *HttpResponse
}
