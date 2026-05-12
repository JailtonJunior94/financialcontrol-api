package entities

import (
	"strings"
	"time"

	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

// Clock abstracts wall-clock access for deterministic testing.
// ports.Clock satisfies this interface via Go structural typing.
type Clock interface {
	Now() time.Time
}

const (
	descriptionMinLen = 1
	descriptionMaxLen = 255
)

// Splitter splits a Money total into installment parts. Minimal interface for Replace.
type Splitter interface {
	Split(total vos.Money, count vos.InstallmentCount) ([]vos.Money, []vos.InstallmentID)
}

// Assigner resolves the target invoice ID for each installment number (1-indexed).
type Assigner interface {
	InvoiceForNumber(number int) (vos.InvoiceID, error)
}

// ReplaceInput carries all fields needed to replace a transaction's mutable state.
type ReplaceInput struct {
	Description      string
	Amount           vos.Amount
	OccurredAt       time.Time
	TransactionType  vos.TransactionType
	PaymentMethod    vos.PaymentMethod
	CardID           *vos.CardID
	CategoryID       vos.CategoryID
	SubcategoryID    *vos.CategoryID
	InstallmentCount int
	Now              time.Time
	// HasClosedOrPaidInvoice signals that the caller observed at least one installment
	// bound to an invoice in state 'closed' or 'paid' (RF-11). Required because the
	// installment.status alone cannot detect 'closed-but-unpaid' invoices (RF-46).
	HasClosedOrPaidInvoice bool
}

// Transaction is the aggregate root for all financial transactions (RF-01..RF-08).
type Transaction struct {
	id                    vos.TransactionID
	userID                identityvo.UserID
	description           string
	amount                vos.Amount
	occurredAt            time.Time
	transactionType       vos.TransactionType
	paymentMethod         vos.PaymentMethod
	cardID                *vos.CardID
	categoryID            vos.CategoryID
	subcategoryID         *vos.CategoryID
	originalTransactionID *vos.TransactionID
	installments          []*Installment
	legacyOrigin          *string
	createdAt             time.Time
	updatedAt             time.Time
	deletedAt             *time.Time
}

// NewTransaction constructs a Transaction, validating RF-05/RF-07/RF-08 structural invariants.
// The caller is responsible for generating the ID (via vos.NewTransactionID or IDGenerator).
func NewTransaction(
	id vos.TransactionID,
	userID identityvo.UserID,
	description string,
	amount vos.Amount,
	occurredAt time.Time,
	txType vos.TransactionType,
	paymentMethod vos.PaymentMethod,
	cardID *vos.CardID,
	categoryID vos.CategoryID,
	subcategoryID *vos.CategoryID,
	originalTransactionID *vos.TransactionID,
	clock Clock,
) (*Transaction, error) {
	desc := strings.TrimSpace(description)
	if len(desc) < descriptionMinLen || len(desc) > descriptionMaxLen {
		return nil, domain.ErrDescriptionRequired
	}
	if paymentMethod.RequiresCard() && cardID == nil {
		return nil, domain.ErrCardRequiredForPaymentMethod
	}
	saoPaulo, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		return nil, err
	}
	now := clock.Now()
	limitSP := now.In(saoPaulo).AddDate(0, 0, 365)
	if occurredAt.In(saoPaulo).After(limitSP) {
		return nil, domain.ErrOccurredAtTooFarInFuture
	}
	return &Transaction{
		id:                    id,
		userID:                userID,
		description:           desc,
		amount:                amount,
		occurredAt:            occurredAt.UTC(),
		transactionType:       txType,
		paymentMethod:         paymentMethod,
		cardID:                cardID,
		categoryID:            categoryID,
		subcategoryID:         subcategoryID,
		originalTransactionID: originalTransactionID,
		createdAt:             now.UTC(),
		updatedAt:             now.UTC(),
	}, nil
}

// RehydrateTransaction reconstructs a Transaction from persisted data.
func RehydrateTransaction(
	id vos.TransactionID,
	userID identityvo.UserID,
	description string,
	amount vos.Amount,
	occurredAt time.Time,
	txType vos.TransactionType,
	paymentMethod vos.PaymentMethod,
	cardID *vos.CardID,
	categoryID vos.CategoryID,
	subcategoryID *vos.CategoryID,
	originalTransactionID *vos.TransactionID,
	legacyOrigin *string,
	createdAt time.Time,
	updatedAt time.Time,
	deletedAt *time.Time,
) *Transaction {
	return &Transaction{
		id:                    id,
		userID:                userID,
		description:           description,
		amount:                amount,
		occurredAt:            occurredAt,
		transactionType:       txType,
		paymentMethod:         paymentMethod,
		cardID:                cardID,
		categoryID:            categoryID,
		subcategoryID:         subcategoryID,
		originalTransactionID: originalTransactionID,
		legacyOrigin:          legacyOrigin,
		createdAt:             createdAt,
		updatedAt:             updatedAt,
		deletedAt:             deletedAt,
	}
}

// SetInstallments attaches installments loaded by the use case or repository.
func (t *Transaction) SetInstallments(items []*Installment) {
	t.installments = items
}

// Replace updates the transaction's mutable fields (RF-11/RF-12).
// For installment_purchase it re-splits installments via splitter/assigner.
// Returns ErrInstallmentInClosedOrPaidInvoice if any current installment is in a
// closed/paid invoice (either signalled via input.HasClosedOrPaidInvoice — primary
// guard reading the parent invoice state — or detected by the installment.status
// defense-in-depth loop).
func (t *Transaction) Replace(input ReplaceInput, splitter Splitter, assigner Assigner) error {
	if input.HasClosedOrPaidInvoice {
		return domain.ErrInstallmentInClosedOrPaidInvoice
	}
	for _, inst := range t.installments {
		if inst.IsClosedOrPaid() {
			return domain.ErrInstallmentInClosedOrPaidInvoice
		}
	}
	desc := strings.TrimSpace(input.Description)
	if len(desc) < descriptionMinLen || len(desc) > descriptionMaxLen {
		return domain.ErrDescriptionRequired
	}
	t.description = desc
	t.amount = input.Amount
	t.occurredAt = input.OccurredAt.UTC()
	t.transactionType = input.TransactionType
	t.paymentMethod = input.PaymentMethod
	t.cardID = input.CardID
	t.categoryID = input.CategoryID
	t.subcategoryID = input.SubcategoryID
	t.updatedAt = input.Now.UTC()

	if input.TransactionType != vos.TransactionTypeInstallmentPurchase {
		t.softDeleteActiveInstallments(input.Now.UTC())
		return nil
	}
	if splitter == nil || assigner == nil {
		return domain.ErrSplitterRequired
	}
	count, err := vos.NewInstallmentCount(input.InstallmentCount)
	if err != nil {
		return err
	}
	t.softDeleteActiveInstallments(input.Now.UTC())
	amounts, ids := splitter.Split(input.Amount.Money(), count)
	for i, amount := range amounts {
		number, _ := vos.NewInstallmentNumber(i + 1)
		invoiceID, invErr := assigner.InvoiceForNumber(i + 1)
		if invErr != nil {
			return invErr
		}
		t.installments = append(t.installments, newInstallment(ids[i], t.id, invoiceID, number, count, amount, input.Now))
	}
	return nil
}

// softDeleteActiveInstallments marks every non-deleted installment as refunded and
// sets its deleted_at to now. Used by Replace when discarding old installments —
// either to recreate them for installment_purchase (RF-12) or to release them when
// the transaction type transitions away from card-based (RF-11/RF-54).
func (t *Transaction) softDeleteActiveInstallments(now time.Time) {
	for _, inst := range t.installments {
		if inst.IsDeleted() {
			continue
		}
		inst.status = vos.InstallmentStatusRefunded
		inst.deletedAt = &now
		inst.updatedAt = now
	}
}

// SoftDelete applies RF-44/RF-55 guards before marking the transaction deleted.
func (t *Transaction) SoftDelete(at time.Time, hasClosedOrPaidInstallment bool, hasDependentRefund bool) error {
	if hasClosedOrPaidInstallment {
		return domain.ErrInstallmentInClosedOrPaidInvoice
	}
	if hasDependentRefund {
		return domain.ErrTransactionHasDependentRefund
	}
	d := at.UTC()
	t.deletedAt = &d
	t.updatedAt = at.UTC()
	return nil
}

// Anticipate moves an installment to a different open invoice (RF-14).
func (t *Transaction) Anticipate(installmentID vos.InstallmentID, targetInvoice *Invoice, now time.Time) error {
	if !targetInvoice.IsOpen() {
		return domain.ErrInstallmentInClosedOrPaidInvoice
	}
	inst := t.findInstallment(installmentID)
	if inst == nil {
		return domain.ErrInstallmentNotFound
	}
	if !inst.status.CanTransitionTo(vos.InstallmentStatusAnticipated) {
		return domain.ErrInstallmentInClosedOrPaidInvoice
	}
	inst.changeInvoice(targetInvoice.ID(), vos.InstallmentStatusAnticipated, now)
	t.updatedAt = now.UTC()
	return nil
}

// IsRefund reports whether this transaction is of type refund.
func (t *Transaction) IsRefund() bool { return t.transactionType.IsRefund() }

// IsCardPurchase reports whether this transaction is a card-based purchase.
func (t *Transaction) IsCardPurchase() bool { return t.transactionType.IsCardBased() }

func (t *Transaction) findInstallment(id vos.InstallmentID) *Installment {
	for _, inst := range t.installments {
		if inst.ID() == id {
			return inst
		}
	}
	return nil
}

func (t *Transaction) ID() vos.TransactionID                     { return t.id }
func (t *Transaction) UserID() identityvo.UserID                 { return t.userID }
func (t *Transaction) Description() string                       { return t.description }
func (t *Transaction) Amount() vos.Amount                        { return t.amount }
func (t *Transaction) OccurredAt() time.Time                     { return t.occurredAt }
func (t *Transaction) TransactionType() vos.TransactionType      { return t.transactionType }
func (t *Transaction) PaymentMethod() vos.PaymentMethod          { return t.paymentMethod }
func (t *Transaction) CardID() *vos.CardID                       { return t.cardID }
func (t *Transaction) CategoryID() vos.CategoryID                { return t.categoryID }
func (t *Transaction) SubcategoryID() *vos.CategoryID            { return t.subcategoryID }
func (t *Transaction) OriginalTransactionID() *vos.TransactionID { return t.originalTransactionID }
func (t *Transaction) Installments() []*Installment              { return t.installments }
func (t *Transaction) LegacyOrigin() *string                     { return t.legacyOrigin }
func (t *Transaction) CreatedAt() time.Time                      { return t.createdAt }
func (t *Transaction) UpdatedAt() time.Time                      { return t.updatedAt }
func (t *Transaction) DeletedAt() *time.Time                     { return t.deletedAt }
func (t *Transaction) IsDeleted() bool                           { return t.deletedAt != nil }
