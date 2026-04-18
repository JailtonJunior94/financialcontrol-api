package dtos

type TransactionQuery struct {
	ID            string `db:"Id"`
	TransactionID string `db:"TransactionId"`
	UserID        string `db:"UserId"`
}
