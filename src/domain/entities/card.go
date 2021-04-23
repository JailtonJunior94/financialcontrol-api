package entities

import "time"

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

func (p *Card) NewCard(userId, flagId, name, description, number string, closingDay int, expirationDate time.Time) {
	p.Entity.NewEntity()
	p.UserId = userId
	p.FlagId = flagId
	p.Name = name
	p.Description = description
	p.Number = number
	p.ClosingDay = closingDay
	p.ExpirationDate = expirationDate
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
