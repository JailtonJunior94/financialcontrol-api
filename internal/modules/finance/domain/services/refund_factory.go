package services

import (
	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/ports"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
)

// RefundOverride allows the caller to override select fields on the generated refund.
type RefundOverride struct {
	PaymentMethod *vos.PaymentMethod
	Description   *string
}

// RefundFactory builds a refund Transaction from an original (RF-13/RF-53, ADR-007).
// Refunds are autonomous: they inherit payment_method, copy card_id when applicable,
// and do not create installments or mutate the original transaction.
type RefundFactory struct{}

func NewRefundFactory() *RefundFactory { return &RefundFactory{} }

// Build constructs the refund Transaction.
//
//   - Returns ErrRefundOfRefundNotAllowed when original.IsRefund().
//   - Returns ErrRefundAlreadyExists when hasActiveRefund is true.
//   - Inherits paymentMethod from original; override replaces it.
//   - Copies card_id from original when the effective paymentMethod requires a card.
//   - When override switches to a non-card method, card_id is discarded.
func (f *RefundFactory) Build(
	original *entities.Transaction,
	hasActiveRefund bool,
	override RefundOverride,
	clock ports.Clock,
	ids ports.IDGenerator,
) (*entities.Transaction, error) {
	if original.IsRefund() {
		return nil, domain.ErrRefundOfRefundNotAllowed
	}
	if hasActiveRefund {
		return nil, domain.ErrRefundAlreadyExists
	}
	pm := original.PaymentMethod()
	if override.PaymentMethod != nil {
		pm = *override.PaymentMethod
	}
	var cardID *vos.CardID
	if pm.RequiresCard() && original.CardID() != nil {
		cid := *original.CardID()
		cardID = &cid
	}
	desc := original.Description()
	if override.Description != nil {
		desc = *override.Description
	}
	originalID := original.ID()
	return entities.NewTransaction(
		ids.NewTransactionID(),
		original.UserID(),
		desc,
		original.Amount(),
		clock.Now(),
		vos.TransactionTypeRefund,
		pm,
		cardID,
		original.CategoryID(),
		original.SubcategoryID(),
		&originalID,
		clock,
	)
}
