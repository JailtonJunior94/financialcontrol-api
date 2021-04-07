package interfaces

import (
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/src/domain/entities"
)

type IBillRepository interface {
	GetBills() (bills []entities.Bill, err error)
	GetBillById(id string) (bill *entities.Bill, err error)
	GetBillByDate(startDate, endDate time.Time) (bill *entities.Bill, err error)
	AddBill(p *entities.Bill) (bill *entities.Bill, err error)
	UpdateBill(p *entities.Bill) (bill *entities.Bill, err error)
	GetBillItemByBillId(billId string) (billItems []entities.BillItem, err error)
	AddBillItem(p *entities.BillItem) (billItem *entities.BillItem, err error)
}
