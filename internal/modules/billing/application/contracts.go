package application

import (
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/internal/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/platform/persistence"
	"github.com/jailtonjunior94/financialcontrol-api/internal/platform/web"
)

type HttpResponse = web.HttpResponse

type BillRepository interface {
	GetBills() ([]entities.Bill, error)
	GetBillById(id string) (*entities.Bill, error)
	GetBillByDate(startDate, endDate time.Time) (*entities.Bill, error)
	AddBill(bill *entities.Bill) (*entities.Bill, error)
	UpdateBill(bill *entities.Bill) (*entities.Bill, error)
	GetBillItemByBillId(billID string) ([]entities.BillItem, error)
	GetBillItemById(id, billID string) (*entities.BillItem, error)
	AddBillItem(item *entities.BillItem) (*entities.BillItem, error)
	UpdateBillItem(item *entities.BillItem) (*entities.BillItem, error)
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
