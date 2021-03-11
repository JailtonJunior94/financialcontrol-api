package entities

import (
	"database/sql"
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/adapters"
)

type Entity struct {
	ID        string       `db:"Id"`
	CreatedAt time.Time    `db:"CreatedAt"`
	UpdatedAt sql.NullTime `db:"UpdatedAt"`
	Active    bool         `db:"Active"`
}

func (e *Entity) NewEntity() {
	e.ID = adapters.NewUuidAdapter().GenerateUuid()
	e.CreatedAt = time.Now().UTC().Local()
	e.Active = true
}

func (e *Entity) ChangeUpdatedAt() {
	e.UpdatedAt.Time = time.Now().UTC().Local()
}

func (e *Entity) ChangeStatus(status bool) {
	e.Active = status
}
