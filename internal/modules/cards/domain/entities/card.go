package entities

import (
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

type Card struct {
	id             vos.CardID
	userID         identityvo.UserID
	flagID         vos.FlagID
	flag           *Flag
	name           vos.CardName
	number         vos.CardNumber
	description    string
	closingDay     vos.ClosingDay
	dueDay         vos.DueDay
	billingCycle   vos.BillingCycle
	expirationDate time.Time
	active         bool
	createdAt      time.Time
	updatedAt      time.Time
}

func NewCard(
	userID identityvo.UserID,
	flagID vos.FlagID,
	name string,
	number string,
	description string,
	closingDay int,
	expirationDate time.Time,
) (*Card, error) {
	if expirationDate.IsZero() {
		return nil, vos.ErrInvalidDueDay
	}
	cardName, cardNumber, closing, due, err := validateCardFields(name, number, closingDay, expirationDate.Day())
	if err != nil {
		return nil, err
	}

	cycle, err := compatibleBillingCycle(closing, due)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	return &Card{
		id:             vos.NewCardID(),
		userID:         userID,
		flagID:         flagID,
		name:           cardName,
		number:         cardNumber,
		description:    description,
		closingDay:     closing,
		dueDay:         due,
		billingCycle:   cycle,
		expirationDate: expirationDate,
		active:         true,
		createdAt:      now,
		updatedAt:      now,
	}, nil
}

func RehydrateCard(
	id vos.CardID,
	userID identityvo.UserID,
	flagID vos.FlagID,
	name string,
	number string,
	description string,
	closingDay int,
	expirationDate time.Time,
	createdAt time.Time,
	updatedAt time.Time,
	active bool,
) (*Card, error) {
	cardName, cardNumber, closing, due, err := validateCardFields(name, number, closingDay, expirationDate.Day())
	if err != nil {
		return nil, err
	}

	if expirationDate.IsZero() {
		expirationDate = legacyExpirationDate(due)
	}

	return &Card{
		id:             id,
		userID:         userID,
		flagID:         flagID,
		name:           cardName,
		number:         cardNumber,
		description:    description,
		closingDay:     closing,
		dueDay:         due,
		billingCycle:   vos.RehydrateBillingCycle(closing, due),
		expirationDate: expirationDate,
		active:         active,
		createdAt:      createdAt,
		updatedAt:      updatedAt,
	}, nil
}

func (c *Card) Update(
	flagID vos.FlagID,
	name string,
	number string,
	description string,
	closingDay int,
	expirationDate time.Time,
) error {
	if expirationDate.IsZero() {
		return vos.ErrInvalidDueDay
	}
	cardName, cardNumber, closing, due, err := validateCardFields(name, number, closingDay, expirationDate.Day())
	if err != nil {
		return err
	}

	cycle, err := c.billingCycleForUpdate(closing, due)
	if err != nil {
		return err
	}

	c.flagID = flagID
	c.name = cardName
	c.number = cardNumber
	c.description = description
	c.closingDay = closing
	c.dueDay = due
	c.billingCycle = cycle
	c.expirationDate = expirationDate
	c.updatedAt = time.Now().UTC()
	return nil
}

func (c *Card) Deactivate() {
	c.active = false
	c.updatedAt = time.Now().UTC()
}

func (c *Card) BestPurchaseDay() int {
	return c.billingCycle.BestPurchaseDay()
}

func (c *Card) DueDateFor(purchase time.Time) time.Time {
	return c.billingCycle.DueDateFor(purchase)
}

func (c *Card) ID() vos.CardID             { return c.id }
func (c *Card) UserID() identityvo.UserID  { return c.userID }
func (c *Card) FlagID() vos.FlagID         { return c.flagID }
func (c *Card) Flag() *Flag                { return c.flag }
func (c *Card) Name() vos.CardName         { return c.name }
func (c *Card) Number() vos.CardNumber     { return c.number }
func (c *Card) Description() string        { return c.description }
func (c *Card) ClosingDay() vos.ClosingDay { return c.closingDay }
func (c *Card) DueDay() vos.DueDay         { return c.dueDay }
func (c *Card) ExpirationDate() time.Time  { return c.expirationDate }
func (c *Card) Active() bool               { return c.active }
func (c *Card) CreatedAt() time.Time       { return c.createdAt }
func (c *Card) UpdatedAt() time.Time       { return c.updatedAt }

// AttachFlag attaches catalog data loaded by the persistence layer; it does not
// affect identity, timestamps, or invariants.
func (c *Card) AttachFlag(f *Flag) { c.flag = f }

func (c *Card) billingCycleForUpdate(closing vos.ClosingDay, due vos.DueDay) (vos.BillingCycle, error) {
	return compatibleBillingCycle(closing, due)
}

func compatibleBillingCycle(closing vos.ClosingDay, due vos.DueDay) (vos.BillingCycle, error) {
	if closing.Int() == due.Int() {
		return vos.RehydrateBillingCycle(closing, due), nil
	}
	return vos.NewBillingCycle(closing, due)
}

func validateCardFields(name, number string, closingDay, dueDay int) (vos.CardName, vos.CardNumber, vos.ClosingDay, vos.DueDay, error) {
	cardName, err := vos.NewCardName(name)
	if err != nil {
		return "", "", 0, 0, err
	}

	cardNumber, err := vos.NewCardNumber(number)
	if err != nil {
		return "", "", 0, 0, err
	}

	closing, err := vos.NewClosingDay(closingDay)
	if err != nil {
		return "", "", 0, 0, err
	}

	due, err := vos.NewDueDay(dueDay)
	if err != nil {
		return "", "", 0, 0, err
	}

	return cardName, cardNumber, closing, due, nil
}

func legacyExpirationDate(due vos.DueDay) time.Time {
	return time.Date(1900, time.January, due.Int(), 0, 0, 0, 0, time.UTC)
}
