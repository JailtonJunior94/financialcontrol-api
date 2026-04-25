package domain

import (
	"math"
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/pkg/shared"
)

type Bill struct {
	Date         time.Time `db:"Date"`
	Total        float64   `db:"Total"`
	SixtyPercent float64   `db:"SixtyPercent"`
	FortyPercent float64   `db:"FortyPercent"`
	Entity
	BillItems []BillItem
}

func (b *Bill) NewBill(date time.Time) {
	b.NewEntity()
	b.Date = shared.NewTime(shared.Time{Date: date}).FormatDate()
}

func (b *Bill) AddBillItems(billItems []BillItem) {
	b.BillItems = billItems
}

func (b *Bill) UpdatingValues() {
	b.sumTotal()
	b.sumSixtyPercent()
	b.sumFortyPercent()
	b.ChangeUpdatedAt()
}

func (b *Bill) sumTotal() {
	var total float64

	if len(b.BillItems) == 0 {
		b.Total = total
	}

	for _, bill := range b.BillItems {
		total += bill.Value
	}

	b.Total = total
}

func (b *Bill) sumSixtyPercent() {
	b.SixtyPercent = math.Round((60.0 / 100.0) * b.Total)
}

func (b *Bill) sumFortyPercent() {
	b.FortyPercent = math.Round((40.0 / 100.0) * b.Total)
}

type BillItem struct {
	BillId string  `db:"BillId"`
	Title  string  `db:"Title"`
	Value  float64 `db:"Value"`
	Entity
	Bill Bill
}

func (b *BillItem) NewBillItem(billId, title string, value float64) {
	b.NewEntity()
	b.BillId = billId
	b.Title = title
	b.Value = value
}

func (b *BillItem) Update(title string, value float64) {
	b.Title = title
	b.Value = value
	b.ChangeUpdatedAt()
}

func (b *BillItem) UpdateStatus(status bool) {
	b.ChangeUpdatedAt()
	b.ChangeStatus(status)
}
