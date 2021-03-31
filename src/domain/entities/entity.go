package entities

import (
	"database/sql"
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/adapters"
	"github.com/jailtonjunior94/financialcontrol-api/src/shared"
)

type Entity struct {
	ID        string       `db:"Id"`
	CreatedAt time.Time    `db:"CreatedAt"`
	UpdatedAt sql.NullTime `db:"UpdatedAt"`
	Active    bool         `db:"Active"`
}

func (e *Entity) NewEntity() {
	timer := shared.NewTime()
	e.ID = adapters.NewUuidAdapter().GenerateUuid()
	e.CreatedAt = timer.Now
	e.Active = true
}

func (e *Entity) ChangeUpdatedAt() {
	timer := shared.NewTime()
	e.UpdatedAt.Time = timer.Now
}

func (e *Entity) ChangeStatus(status bool) {
	e.Active = status
}
