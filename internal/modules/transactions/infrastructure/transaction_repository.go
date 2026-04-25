package infrastructure

import (
	"database/sql"
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/pkg/persistence"
	"github.com/jailtonjunior94/financialcontrol-api/internal/infrastructure/database"
	transactionsapp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/transactions/application"
	transactionsdomain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/transactions/domain"
)

type TransactionRepository struct {
	db database.ISqlConnection
}

func NewTransactionRepository(db database.ISqlConnection) transactionsapp.TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) AddTransaction(transaction *transactionsdomain.Transaction) (*transactionsdomain.Transaction, error) {
	statement, err := r.db.OpenConnectionAndMountStatement(addTransaction)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = statement.Close()
	}()

	result, err := statement.Exec(
		sql.Named("id", transaction.ID),
		sql.Named("userId", transaction.UserId),
		sql.Named("date", transaction.Date),
		sql.Named("total", transaction.Total),
		sql.Named("income", transaction.Income),
		sql.Named("outcome", transaction.Outcome),
		sql.Named("createdAt", transaction.CreatedAt),
		sql.Named("updatedAt", transaction.UpdatedAt),
		sql.Named("active", transaction.Active),
	)
	if err := r.db.ValidateResult(result, err); err != nil {
		return nil, err
	}

	return transaction, nil
}

func (r *TransactionRepository) AddTransactionItem(item *transactionsdomain.TransactionItem) (*transactionsdomain.TransactionItem, error) {
	statement, err := r.db.OpenConnectionAndMountStatement(addTransactionItem)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = statement.Close()
	}()

	result, err := statement.Exec(
		sql.Named("id", item.ID),
		sql.Named("transactionId", item.TransactionId),
		sql.Named("title", item.Title),
		sql.Named("value", item.Value),
		sql.Named("type", item.Type),
		sql.Named("createdAt", item.CreatedAt),
		sql.Named("updatedAt", item.UpdatedAt),
		sql.Named("isPaid", item.IsPaid),
		sql.Named("active", item.Active),
	)
	if err := r.db.ValidateResult(result, err); err != nil {
		return nil, err
	}

	return item, nil
}

func (r *TransactionRepository) AddRangeTransactionItems(transaction *transactionsdomain.Transaction, items []transactionsdomain.TransactionItem) error {
	if _, err := r.AddTransaction(transaction); err != nil {
		return err
	}

	for index := range items {
		if _, err := r.AddTransactionItem(&items[index]); err != nil {
			return err
		}
	}

	return nil
}

func (r *TransactionRepository) GetItemByTransactionId(transactionID string) ([]transactionsdomain.TransactionItem, error) {
	var items []transactionsdomain.TransactionItem
	if err := r.db.Connect().Select(&items, getItemByTransactionId, sql.Named("transactionId", transactionID)); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *TransactionRepository) GetTransactions(userID string) ([]transactionsdomain.Transaction, error) {
	var transactions []transactionsdomain.Transaction
	if err := r.db.Connect().Select(&transactions, getTransactions, sql.Named("userId", userID)); err != nil {
		return nil, err
	}

	return transactions, nil
}

func (r *TransactionRepository) GetTransactionById(id, userID string) (*transactionsdomain.Transaction, error) {
	row := r.db.Connect().QueryRow(getTransactionById, sql.Named("id", id), sql.Named("userId", userID))

	transaction := new(transactionsdomain.Transaction)
	err := row.Scan(&transaction.ID, &transaction.UserId, &transaction.Date, &transaction.Total, &transaction.Income, &transaction.Outcome, &transaction.CreatedAt, &transaction.UpdatedAt, &transaction.Active)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return transaction, nil
}

func (r *TransactionRepository) GetTransactionByDate(startDate, endDate time.Time, userID string) (*transactionsdomain.Transaction, error) {
	row := r.db.Connect().QueryRow(getTransactionByDate, sql.Named("startDate", startDate), sql.Named("endDate", endDate), sql.Named("userId", userID))

	transaction := new(transactionsdomain.Transaction)
	err := row.Scan(&transaction.ID, &transaction.UserId, &transaction.Date, &transaction.Total, &transaction.Income, &transaction.Outcome, &transaction.CreatedAt, &transaction.UpdatedAt, &transaction.Active)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return transaction, nil
}

func (r *TransactionRepository) UpdateTransaction(transaction *transactionsdomain.Transaction) (*transactionsdomain.Transaction, error) {
	statement, err := r.db.OpenConnectionAndMountStatement(updateTransaction)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = statement.Close()
	}()

	result, err := statement.Exec(
		sql.Named("id", transaction.ID),
		sql.Named("total", transaction.Total),
		sql.Named("income", transaction.Income),
		sql.Named("outcome", transaction.Outcome),
		sql.Named("updatedAt", transaction.UpdatedAt.Time),
	)
	if err := r.db.ValidateResult(result, err); err != nil {
		return nil, err
	}

	return transaction, nil
}

func (r *TransactionRepository) GetTransactionItemsById(transactionID, id string) (*transactionsdomain.TransactionItem, error) {
	row := r.db.Connect().QueryRow(getTransactionItemsById, sql.Named("id", id), sql.Named("transactionId", transactionID))

	item := new(transactionsdomain.TransactionItem)
	err := row.Scan(&item.ID, &item.TransactionId, &item.Title, &item.Value, &item.Type, &item.CreatedAt, &item.UpdatedAt, &item.IsPaid, &item.Active)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return item, nil
}

func (r *TransactionRepository) FetchTransactionByDate(date time.Time, cardDescription string) (*persistence.TransactionQuery, error) {
	query := `SELECT
				CAST(t.Id AS CHAR(36)) [TransactionId],
				CAST(ti.Id AS CHAR(36)) [Id],
				CAST(t.UserId AS CHAR(36)) [UserId]
			FROM
			dbo.[Transaction] t
			INNER JOIN dbo.TransactionItem ti ON ti.TransactionId = t.Id
			WHERE
			t.[Date] = CONVERT(DATETIME, @date)
			AND ti.Title = @cardDescription`

	row := r.db.Connect().QueryRow(query, sql.Named("date", date), sql.Named("cardDescription", cardDescription))

	transaction := new(persistence.TransactionQuery)
	err := row.Scan(&transaction.TransactionID, &transaction.ID, &transaction.UserID)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return transaction, nil
}

func (r *TransactionRepository) UpdateTransactionItem(item *transactionsdomain.TransactionItem) (*transactionsdomain.TransactionItem, error) {
	statement, err := r.db.OpenConnectionAndMountStatement(updateTransactionItem)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = statement.Close()
	}()

	result, err := statement.Exec(
		sql.Named("id", item.ID),
		sql.Named("transactionId", item.TransactionId),
		sql.Named("title", item.Title),
		sql.Named("value", item.Value),
		sql.Named("type", item.Type),
		sql.Named("updatedAt", item.UpdatedAt.Time),
		sql.Named("isPaid", item.IsPaid),
		sql.Named("active", item.Active),
	)
	if err := r.db.ValidateResult(result, err); err != nil {
		return nil, err
	}

	return item, nil
}
