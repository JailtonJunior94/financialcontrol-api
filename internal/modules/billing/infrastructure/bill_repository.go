package infrastructure

import (
	"database/sql"
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/internal/platform/persistence"
	"github.com/jailtonjunior94/financialcontrol-api/internal/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/infrastructure/database"
)

type BillRepository struct {
	db database.ISqlConnection
}

func NewBillRepository(db database.ISqlConnection) *BillRepository {
	return &BillRepository{db: db}
}

func (r *BillRepository) GetBills() ([]entities.Bill, error) {
	var bills []entities.Bill
	if err := r.db.Connect().Select(&bills, getBills); err != nil {
		return nil, err
	}

	return bills, nil
}

func (r *BillRepository) GetBillById(id string) (*entities.Bill, error) {
	row := r.db.Connect().QueryRow(getBillByID, sql.Named("id", id))

	bill := new(entities.Bill)
	err := row.Scan(&bill.ID, &bill.Date, &bill.Total, &bill.SixtyPercent, &bill.FortyPercent, &bill.CreatedAt, &bill.UpdatedAt, &bill.Active)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return bill, nil
}

func (r *BillRepository) GetBillByDate(startDate, endDate time.Time) (*entities.Bill, error) {
	row := r.db.Connect().QueryRow(getBillByDate, sql.Named("startDate", startDate), sql.Named("endDate", endDate))

	bill := new(entities.Bill)
	err := row.Scan(&bill.ID, &bill.Date, &bill.Total, &bill.SixtyPercent, &bill.FortyPercent, &bill.CreatedAt, &bill.UpdatedAt, &bill.Active)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return bill, nil
}

func (r *BillRepository) AddBill(bill *entities.Bill) (*entities.Bill, error) {
	statement, err := r.db.OpenConnectionAndMountStatement(addBill)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = statement.Close()
	}()

	result, err := statement.Exec(
		sql.Named("id", bill.ID),
		sql.Named("date", bill.Date),
		sql.Named("total", bill.Total),
		sql.Named("sixtyPercent", bill.SixtyPercent),
		sql.Named("fortyPercent", bill.FortyPercent),
		sql.Named("createdAt", bill.CreatedAt),
		sql.Named("updatedAt", bill.UpdatedAt),
		sql.Named("active", bill.Active),
	)
	if err := r.db.ValidateResult(result, err); err != nil {
		return nil, err
	}

	return bill, nil
}

func (r *BillRepository) UpdateBill(bill *entities.Bill) (*entities.Bill, error) {
	statement, err := r.db.OpenConnectionAndMountStatement(updateBill)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = statement.Close()
	}()

	result, err := statement.Exec(
		sql.Named("id", bill.ID),
		sql.Named("total", bill.Total),
		sql.Named("sixtyPercent", bill.SixtyPercent),
		sql.Named("fortyPercent", bill.FortyPercent),
		sql.Named("updatedAt", bill.UpdatedAt.Time),
	)
	if err := r.db.ValidateResult(result, err); err != nil {
		return nil, err
	}

	return bill, nil
}

func (r *BillRepository) GetBillItemByBillId(billID string) ([]entities.BillItem, error) {
	var items []entities.BillItem
	if err := r.db.Connect().Select(&items, getBillItemByBillID, sql.Named("billId", billID)); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *BillRepository) GetBillItemById(id, billID string) (*entities.BillItem, error) {
	row := r.db.Connect().QueryRow(getBillItemByID, sql.Named("id", id), sql.Named("billId", billID))

	item := new(entities.BillItem)
	err := row.Scan(&item.ID, &item.BillId, &item.Title, &item.Value, &item.CreatedAt, &item.UpdatedAt, &item.Active)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return item, nil
}

func (r *BillRepository) AddBillItem(item *entities.BillItem) (*entities.BillItem, error) {
	statement, err := r.db.OpenConnectionAndMountStatement(addBillItem)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = statement.Close()
	}()

	result, err := statement.Exec(
		sql.Named("id", item.ID),
		sql.Named("billId", item.BillId),
		sql.Named("title", item.Title),
		sql.Named("value", item.Value),
		sql.Named("createdAt", item.CreatedAt),
		sql.Named("updatedAt", item.UpdatedAt),
		sql.Named("active", item.Active),
	)
	if err := r.db.ValidateResult(result, err); err != nil {
		return nil, err
	}

	return item, nil
}

func (r *BillRepository) UpdateBillItem(item *entities.BillItem) (*entities.BillItem, error) {
	statement, err := r.db.OpenConnectionAndMountStatement(updateBillItem)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = statement.Close()
	}()

	result, err := statement.Exec(
		sql.Named("id", item.ID),
		sql.Named("billId", item.BillId),
		sql.Named("title", item.Title),
		sql.Named("value", item.Value),
		sql.Named("updatedAt", item.UpdatedAt.Time),
		sql.Named("active", item.Active),
	)
	if err := r.db.ValidateResult(result, err); err != nil {
		return nil, err
	}

	return item, nil
}

func (r *BillRepository) Get(date time.Time) (*persistence.BillQuery, error) {
	query := `SELECT
				CAST(b.Id AS CHAR(36)) [BillID],
				b.Date,
				CAST(bi.Id AS CHAR(36)) [BillItemID],
				bi.Title,
				bi.Value
			  FROM
				Bill b
			  INNER JOIN BillItem bi ON bi.BillId = b.Id
			  WHERE
				b.[Date] = @date
			  AND b.Active = 1`

	rows, err := r.db.Connect().Query(query, sql.Named("date", date))
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var bill persistence.BillQuery
	var item persistence.BillItemQuery
	itemMap := make(map[string][]persistence.BillItemQuery)

	for rows.Next() {
		if err := rows.Scan(&bill.ID, &bill.Date, &item.ID, &item.Description, &item.Total); err != nil {
			return nil, err
		}

		itemMap[bill.ID] = append(itemMap[bill.ID], item)
	}

	bill.Items = itemMap[bill.ID]
	return &bill, nil
}
