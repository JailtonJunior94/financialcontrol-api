package repositories

import (
	"database/sql"
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/src/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/interfaces"
	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/database"
	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/queries"
)

type BillRepository struct {
	Db database.ISqlConnection
}

func NewBillRepository(db database.ISqlConnection) interfaces.IBillRepository {
	return &BillRepository{Db: db}
}

func (r *BillRepository) GetBills() (bills []entities.Bill, err error) {
	connection := r.Db.Connect()
	if err := connection.Select(&bills, queries.GetBills); err != nil {
		return nil, err
	}
	return bills, nil
}

func (r *BillRepository) GetBillByDate(startDate, endDate time.Time) (bill *entities.Bill, err error) {
	connection := r.Db.Connect()
	row := connection.QueryRow(queries.GetBillByDate, sql.Named("startDate", startDate), sql.Named("endDate", endDate))

	b := new(entities.Bill)
	err = row.Scan(&b.ID, &b.Date, &b.Total, &b.SixtyPercent, &b.FortyPercent, &b.CreatedAt, &b.UpdatedAt, &b.Active)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return b, nil
}

func (r *BillRepository) AddBill(p *entities.Bill) (bill *entities.Bill, err error) {
	s, err := r.Db.OpenConnectionAndMountStatement(queries.AddBill)
	if err != nil {
		return nil, err
	}
	defer s.Close()

	result, err := s.Exec(
		sql.Named("id", p.ID),
		sql.Named("date", p.Date),
		sql.Named("total", p.Total),
		sql.Named("sixtyPercent", p.SixtyPercent),
		sql.Named("fortyPercent", p.FortyPercent),
		sql.Named("createdAt", p.CreatedAt),
		sql.Named("updatedAt", p.UpdatedAt),
		sql.Named("active", p.Active))

	if err := r.Db.ValidateResult(result, err); err != nil {
		return nil, err
	}
	return p, nil
}
