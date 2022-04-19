package entities

import (
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/src/shared"
)

type Card struct {
	UserId         string    `db:"UserId"`
	FlagId         string    `db:"FlagId"`
	Name           string    `db:"Name"`
	Number         string    `db:"Number"`
	Description    string    `db:"Description"`
	ClosingDay     int       `db:"ClosingDay"`
	ExpirationDate time.Time `db:"ExpirationDate"`

	Entity
	User User
	Flag Flag
}

func NewCard(userId, flagId, name, description, number string, closingDay int, expirationDate time.Time) *Card {
	card := &Card{
		UserId:         userId,
		FlagId:         flagId,
		Name:           name,
		Description:    description,
		Number:         number,
		ClosingDay:     closingDay,
		ExpirationDate: expirationDate,
	}
	card.Entity.NewEntity()

	return card
}

func (p *Card) Update(flagId, name, description, number string, closingDay int, expirationDate time.Time) {
	p.ChangeUpdatedAt()
	p.FlagId = flagId
	p.Name = name
	p.Description = description
	p.Number = number
	p.ClosingDay = closingDay
	p.ExpirationDate = expirationDate
}

func (p *Card) UpdateStatus(status bool) {
	p.ChangeUpdatedAt()
	p.ChangeStatus(status)
}

func (p *Card) BestDayToBuy() int {
	time := shared.NewTime()
	lastDay := time.LastDayOfMonth()

	bestDay := lastDay.AddDate(0, 0, -7)
	return bestDay.Day()
}
