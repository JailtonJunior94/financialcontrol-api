package domain

type TransactionItem struct {
	TransactionId string  `db:"TransactionId"`
	Title         string  `db:"Title"`
	Value         float64 `db:"Value"`
	Type          string  `db:"Type"`
	IsPaid        bool    `db:"IsPaid"`

	Entity
}

func NewTransactionItem(transactionID, title, transactionType string, value float64) *TransactionItem {
	item := &TransactionItem{
		TransactionId: transactionID,
		Title:         title,
		Value:         value,
		Type:          transactionType,
		IsPaid:        false,
	}
	item.NewEntity()

	return item
}

func (t *TransactionItem) UpdateTransactionItem(title, transactionType string, value float64) {
	t.Title = title
	t.Type = transactionType
	t.Value = value
	t.ChangeUpdatedAt()
}

func (t *TransactionItem) UpdateStatus(status bool) {
	t.ChangeUpdatedAt()
	t.ChangeStatus(status)
}

func (t *TransactionItem) MarkAsPaid(markAsPaid bool) {
	t.IsPaid = markAsPaid
	t.ChangeUpdatedAt()
}
