package entities

type BillItem struct {
	BillId string  `db:"BillId"`
	Title  string  `db:"Title"`
	Value  float64 `db:"Value"`
	Entity
	Bill Bill
}

func (b *BillItem) NewBillItem(billId, title string, value float64) {
	b.Entity.NewEntity()
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
