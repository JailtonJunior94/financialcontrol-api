package mssql

import (
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

// CardRow holds the DB columns for the Card + joined Flag scan.
type CardRow struct {
	ID             string    `db:"Id"`
	UserID         string    `db:"UserId"`
	FlagID         string    `db:"FlagId"`
	Name           string    `db:"Name"`
	Number         string    `db:"Number"`
	Description    string    `db:"Description"`
	ClosingDay     int       `db:"ClosingDay"`
	ExpirationDate time.Time `db:"ExpirationDate"`
	CreatedAt      time.Time `db:"CreatedAt"`
	UpdatedAt      time.Time `db:"UpdatedAt"`
	Active         bool      `db:"Active"`
	FlagEntityID   string    `db:"FlagEntityId"`
	FlagName       string    `db:"FlagName"`
	FlagActive     bool      `db:"FlagActive"`
}

// FlagRow holds the DB columns for a Flag scan.
type FlagRow struct {
	ID     string `db:"Id"`
	Name   string `db:"Name"`
	Active bool   `db:"Active"`
}

// RowToCard maps a CardRow into a domain Card entity (no IO).
func RowToCard(r *CardRow) (*entities.Card, error) {
	userID, err := identityvo.ParseUserID(r.UserID)
	if err != nil {
		return nil, err
	}

	flagID, err := vos.ParseFlagID(r.FlagID)
	if err != nil {
		return nil, err
	}

	id, err := vos.ParseCardID(r.ID)
	if err != nil {
		return nil, err
	}

	card, err := entities.RehydrateCard(
		id,
		userID,
		flagID,
		r.Name,
		r.Number,
		r.Description,
		r.ClosingDay,
		r.ExpirationDate,
		r.CreatedAt,
		r.UpdatedAt,
		r.Active,
	)
	if err != nil {
		return nil, err
	}

	flagEntityID, err := vos.ParseFlagID(r.FlagEntityID)
	if err != nil {
		return nil, err
	}
	flag := entities.NewFlag(flagEntityID, r.FlagName, r.FlagActive)
	card.AttachFlag(&flag)

	return card, nil
}

// RowToFlag maps a FlagRow into a domain Flag entity (no IO).
func RowToFlag(r *FlagRow) (entities.Flag, error) {
	id, err := vos.ParseFlagID(r.ID)
	if err != nil {
		return entities.Flag{}, err
	}
	return entities.NewFlag(id, r.Name, r.Active), nil
}

func cardToExpirationDate(card *entities.Card) time.Time {
	return card.ExpirationDate()
}
