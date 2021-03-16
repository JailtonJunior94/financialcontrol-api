package entities

type TransactionItem struct {
	TransactionId string  `db:"TransactionId"`
	Title         string  `db:"Title"`
	Value         float64 `db:"Value"`
	Type          string  `db:"Type"`
	Entity
	Transaction Transaction
}

func (u *TransactionItem) NewTransactionItem(transactionId, title, typ string, value float64) {
	u.Entity.NewEntity()
	u.TransactionId = transactionId
	u.Title = title
	u.Value = value
	u.Type = typ
}
