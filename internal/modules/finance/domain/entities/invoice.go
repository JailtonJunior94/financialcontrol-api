package entities

import (
	"time"

	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/projections"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

// Invoice is the aggregate root for a card billing cycle (RF-16/RF-17/RF-18/RF-47).
type Invoice struct {
	id           vos.InvoiceID
	userID       identityvo.UserID
	cardID       vos.CardID
	state        vos.InvoiceState
	cycleStart   time.Time
	cycleEnd     time.Time
	closingDate  time.Time
	dueDate      time.Time
	total        vos.Money
	paidAt       *time.Time
	legacyOrigin *string
	createdAt    time.Time
	updatedAt    time.Time
	deletedAt    *time.Time
}

// NewInvoice creates a new open invoice for the given card and cycle dates.
func NewInvoice(
	card projections.CardView,
	cycleStart time.Time,
	cycleEnd time.Time,
	closingDate time.Time,
	dueDate time.Time,
	now time.Time,
) (*Invoice, error) {
	return &Invoice{
		id:          vos.NewInvoiceID(),
		userID:      card.UserID,
		cardID:      card.ID,
		state:       vos.InvoiceStateOpen,
		cycleStart:  cycleStart.UTC(),
		cycleEnd:    cycleEnd.UTC(),
		closingDate: closingDate.UTC(),
		dueDate:     dueDate.UTC(),
		total:       vos.ZeroMoney(),
		createdAt:   now.UTC(),
		updatedAt:   now.UTC(),
	}, nil
}

// RehydrateInvoice reconstructs an Invoice from persisted data.
func RehydrateInvoice(
	id vos.InvoiceID,
	userID identityvo.UserID,
	cardID vos.CardID,
	state vos.InvoiceState,
	cycleStart time.Time,
	cycleEnd time.Time,
	closingDate time.Time,
	dueDate time.Time,
	total vos.Money,
	paidAt *time.Time,
	legacyOrigin *string,
	createdAt time.Time,
	updatedAt time.Time,
	deletedAt *time.Time,
) *Invoice {
	return &Invoice{
		id:           id,
		userID:       userID,
		cardID:       cardID,
		state:        state,
		cycleStart:   cycleStart,
		cycleEnd:     cycleEnd,
		closingDate:  closingDate,
		dueDate:      dueDate,
		total:        total,
		paidAt:       paidAt,
		legacyOrigin: legacyOrigin,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
		deletedAt:    deletedAt,
	}
}

// CloseIfDue applies RF-17: transitions open → closed when now >= closingDate.
// Returns true if the transition occurred; idempotent when already closed or paid.
func (inv *Invoice) CloseIfDue(now time.Time) bool {
	if inv.state != vos.InvoiceStateOpen {
		return false
	}
	if now.Before(inv.closingDate) {
		return false
	}
	inv.state = vos.InvoiceStateClosed
	inv.updatedAt = now.UTC()
	return true
}

// MarkPaid applies RF-18. Accepts open (implicit close first) or closed; returns
// ErrInvoiceAlreadyPaid for 409 when already paid.
func (inv *Invoice) MarkPaid(at time.Time) error {
	if inv.state == vos.InvoiceStatePaid {
		return domain.ErrInvoiceAlreadyPaid
	}
	if !inv.state.CanTransitionTo(vos.InvoiceStatePaid) {
		return domain.ErrInvoiceCannotPay
	}
	if inv.state == vos.InvoiceStateOpen {
		inv.state = vos.InvoiceStateClosed
	}
	t := at.UTC()
	inv.paidAt = &t
	inv.state = vos.InvoiceStatePaid
	inv.updatedAt = at.UTC()
	return nil
}

// RecalculateTotal updates the invoice total from the given installment slice.
// The caller is responsible for passing the current non-deleted installments.
func (inv *Invoice) RecalculateTotal(items []*Installment) {
	total := vos.ZeroMoney()
	for _, it := range items {
		total = total.Add(it.Amount())
	}
	inv.total = total
}

func (inv *Invoice) ID() vos.InvoiceID         { return inv.id }
func (inv *Invoice) UserID() identityvo.UserID { return inv.userID }
func (inv *Invoice) CardID() vos.CardID        { return inv.cardID }
func (inv *Invoice) State() vos.InvoiceState   { return inv.state }
func (inv *Invoice) CycleStart() time.Time     { return inv.cycleStart }
func (inv *Invoice) CycleEnd() time.Time       { return inv.cycleEnd }
func (inv *Invoice) ClosingDate() time.Time    { return inv.closingDate }
func (inv *Invoice) DueDate() time.Time        { return inv.dueDate }
func (inv *Invoice) Total() vos.Money          { return inv.total }
func (inv *Invoice) PaidAt() *time.Time        { return inv.paidAt }
func (inv *Invoice) LegacyOrigin() *string     { return inv.legacyOrigin }
func (inv *Invoice) CreatedAt() time.Time      { return inv.createdAt }
func (inv *Invoice) UpdatedAt() time.Time      { return inv.updatedAt }
func (inv *Invoice) DeletedAt() *time.Time     { return inv.deletedAt }
func (inv *Invoice) IsDeleted() bool           { return inv.deletedAt != nil }
func (inv *Invoice) IsOpen() bool              { return inv.state == vos.InvoiceStateOpen }
func (inv *Invoice) IsClosed() bool            { return inv.state == vos.InvoiceStateClosed }
func (inv *Invoice) IsPaid() bool              { return inv.state == vos.InvoiceStatePaid }
