package interfaces

import (
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/src/domain/entities"
)

type IBillRepository interface {
	GetBills() (bills []entities.Bill, err error)
	GetBillByDate(startDate, endDate time.Time) (bill *entities.Bill, err error)
	AddBill(p *entities.Bill) (bill *entities.Bill, err error)
}
