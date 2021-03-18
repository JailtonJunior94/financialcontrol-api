package repositories

import (
	"database/sql"

	"github.com/jailtonjunior94/financialcontrol-api/src/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/interfaces"
	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/database"
	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/queries"
)

type TransactionRepository struct {
	Db database.ISqlConnection
}

func NewTransactionRepository(db database.ISqlConnection) interfaces.ITransactionRepository {
	return &TransactionRepository{Db: db}
}

func (r *TransactionRepository) AddTransaction(t *entities.Transaction) (transaction *entities.Transaction, err error) {
	s, err := r.Db.OpenConnectionAndMountStatement(queries.AddTransaction)
	if err != nil {
		return nil, err
	}
	defer s.Close()

	result, err := s.Exec(
		sql.Named("id", t.ID),
		sql.Named("userId", t.UserId),
		sql.Named("date", t.Date),
		sql.Named("total", t.Total),
		sql.Named("income", t.Income),
		sql.Named("outcome", t.Outcome),
		sql.Named("createdAt", t.CreatedAt),
		sql.Named("updatedAt", t.UpdatedAt),
		sql.Named("active", t.Active))

	if err := r.Db.ValidateResult(result, err); err != nil {
		return nil, err
	}
	return t, nil
}

func (r *TransactionRepository) AddTransactionItem(t *entities.TransactionItem) (transactionItem *entities.TransactionItem, err error) {
	s, err := r.Db.OpenConnectionAndMountStatement(queries.AddTransactionItem)
	if err != nil {
		return nil, err
	}
	defer s.Close()

	result, err := s.Exec(
		sql.Named("id", t.ID),
		sql.Named("transactionId", t.TransactionId),
		sql.Named("title", t.Title),
		sql.Named("value", t.Value),
		sql.Named("type", t.Type),
		sql.Named("createdAt", t.CreatedAt),
		sql.Named("updatedAt", t.UpdatedAt),
		sql.Named("active", t.Active))

	if err := r.Db.ValidateResult(result, err); err != nil {
		return nil, err
	}
	return t, nil
}

func (r *TransactionRepository) GetItemByTransactionId(transactionId string) (items []entities.TransactionItem, err error) {
	connection := r.Db.Connect()

	if err := connection.Select(&items, queries.GetItemByTransactionId, sql.Named("transactionId", transactionId)); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *TransactionRepository) GetTransactions(userId string) (transactions []entities.Transaction, err error) {
	connection := r.Db.Connect()
	if err := connection.Select(&transactions, queries.GetTransactions, sql.Named("userId", userId)); err != nil {
		return nil, err
	}
	return transactions, nil
}

func (r *TransactionRepository) GetTransactionById(id string, userId string) (transaction *entities.Transaction, err error) {
	connection := r.Db.Connect()
	row := connection.QueryRow(queries.GetTransactionById, sql.Named("id", id), sql.Named("userId", userId))

	t := new(entities.Transaction)
	if err := row.Scan(&t.ID, &t.UserId, &t.Date, &t.Total, &t.Income, &t.Outcome, &t.CreatedAt, &t.UpdatedAt, &t.Active); err != nil {
		return nil, err
	}
	return t, nil
}
