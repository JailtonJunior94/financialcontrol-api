package entities

type TransactionItem struct {
	TransactionId string  `db:"TransactionId"`
	Title         string  `db:"Title"`
	Value         float64 `db:"Value"`
	Type          string  `db:"Type"`
	IsPaid        bool    `db:"IsPaid"`

	Entity
	Transaction Transaction
}

func NewTransactionItem(transactionId, title, typ string, value float64) *TransactionItem {
	transactionItem := &TransactionItem{
		TransactionId: transactionId,
		Title:         title,
		Value:         value,
		Type:          typ,
		IsPaid:        false,
	}
	transactionItem.Entity.NewEntity()

	return transactionItem
}

func (u *TransactionItem) UpdateTransactionItem(title, typ string, value float64) {
	u.Title = title
	u.Type = typ
	u.Value = value
	u.ChangeUpdatedAt()
}

func (u *TransactionItem) UpdateStatus(status bool) {
	u.ChangeUpdatedAt()
	u.ChangeStatus(status)
}

func (u *TransactionItem) MarkAsPaid(mark bool) {
	u.IsPaid = mark
	u.ChangeUpdatedAt()
}
