package repositories

import (
	"database/sql"
	"time"

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
		sql.Named("isPaid", t.IsPaid),
		sql.Named("active", t.Active))

	if err := r.Db.ValidateResult(result, err); err != nil {
		return nil, err
	}
	return t, nil
}

func (r *TransactionRepository) AddRangeTransactionItems(t *entities.Transaction, ti []entities.TransactionItem) error {
	_, err := r.AddTransaction(t)
	if err != nil {
		return err
	}

	ch := make(chan error)
	var errorsCount []error

	go func() {
		for _, item := range ti {
			_, err := r.AddTransactionItem(&item)
			if err != nil {
				ch <- err
				errorsCount = append(errorsCount, err)
			}
		}
	}()

	if len(errorsCount) > 0 {
		return <-ch
	}

	return nil
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
	err = row.Scan(&t.ID, &t.UserId, &t.Date, &t.Total, &t.Income, &t.Outcome, &t.CreatedAt, &t.UpdatedAt, &t.Active)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return t, nil
}

func (r *TransactionRepository) GetTransactionByDate(startDate, endDate time.Time, userId string) (transaction *entities.Transaction, err error) {
	connection := r.Db.Connect()
	row := connection.QueryRow(queries.GetTransactionByDate, sql.Named("startDate", startDate), sql.Named("endDate", endDate), sql.Named("userId", userId))

	t := new(entities.Transaction)
	err = row.Scan(&t.ID, &t.UserId, &t.Date, &t.Total, &t.Income, &t.Outcome, &t.CreatedAt, &t.UpdatedAt, &t.Active)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return t, nil
}

func (r *TransactionRepository) UpdateTransaction(t *entities.Transaction) (transaction *entities.Transaction, err error) {
	s, err := r.Db.OpenConnectionAndMountStatement(queries.UpdateTransaction)
	if err != nil {
		return nil, err
	}
	defer s.Close()

	result, err := s.Exec(
		sql.Named("id", t.ID),
		sql.Named("total", t.Total),
		sql.Named("income", t.Income),
		sql.Named("outcome", t.Outcome),
		sql.Named("updatedAt", t.UpdatedAt.Time))

	if err := r.Db.ValidateResult(result, err); err != nil {
		return nil, err
	}
	return t, nil
}

func (r *TransactionRepository) GetTransactionItemsById(transactionId, id string) (transactionItem *entities.TransactionItem, err error) {
	connection := r.Db.Connect()
	row := connection.QueryRow(queries.GetTransactionItemsById, sql.Named("id", id), sql.Named("transactionId", transactionId))

	t := new(entities.TransactionItem)
	err = row.Scan(&t.ID, &t.TransactionId, &t.Title, &t.Value, &t.Type, &t.CreatedAt, &t.UpdatedAt, &t.IsPaid, &t.Active)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return t, nil
}

func (r *TransactionRepository) UpdateTransactionItem(t *entities.TransactionItem) (transactionItem *entities.TransactionItem, err error) {
	s, err := r.Db.OpenConnectionAndMountStatement(queries.UpdateTransactionItem)
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
		sql.Named("updatedAt", t.UpdatedAt.Time),
		sql.Named("isPaid", t.IsPaid),
		sql.Named("active", t.Active))

	if err := r.Db.ValidateResult(result, err); err != nil {
		return nil, err
	}
	return t, nil
}
