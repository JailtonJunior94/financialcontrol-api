package services

import (
	"testing"
	"time"

	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/ports"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fixedClock returns a constant time.
type fixedClock struct{ t time.Time }

func (c *fixedClock) Now() time.Time { return c.t }

// seqIDs provides sequential IDs for tests.
type seqIDs struct{}

func (s *seqIDs) NewTransactionID() vos.TransactionID { return vos.NewTransactionID() }
func (s *seqIDs) NewInstallmentID() vos.InstallmentID { return vos.NewInstallmentID() }
func (s *seqIDs) NewInvoiceID() vos.InvoiceID         { return vos.NewInvoiceID() }

func makeTx(t *testing.T, txType vos.TransactionType, pm vos.PaymentMethod, cardID *vos.CardID) *entities.Transaction {
	t.Helper()
	now := time.Date(2026, 5, 11, 12, 0, 0, 0, time.UTC)
	clock := &fixedClock{t: now}
	amount, _ := vos.NewMoney("200.00")
	amountVO, _ := vos.NewAmountFromMoney(amount)
	catID := vos.NewCategoryID()
	tx, err := entities.NewTransaction(
		vos.NewTransactionID(),
		identityvo.NewUserID(),
		"Compra original",
		amountVO,
		now,
		txType,
		pm,
		cardID,
		catID,
		nil,
		nil,
		clock,
	)
	require.NoError(t, err)
	return tx
}

func TestRefundFactory_Build(t *testing.T) {
	t.Parallel()
	factory := NewRefundFactory()
	now := time.Date(2026, 5, 11, 12, 0, 0, 0, time.UTC)
	clock := &fixedClock{t: now}
	ids := &seqIDs{}
	cardID := vos.NewCardID()

	t.Run("refund_of_refund_not_allowed", func(t *testing.T) {
		t.Parallel()
		original := makeTx(t, vos.TransactionTypeRefund, vos.PaymentMethodPix, nil)
		_, err := factory.Build(original, false, RefundOverride{}, clock, ids)
		assert.ErrorIs(t, err, domain.ErrRefundOfRefundNotAllowed)
	})

	t.Run("refund_already_exists", func(t *testing.T) {
		t.Parallel()
		original := makeTx(t, vos.TransactionTypeExpense, vos.PaymentMethodPix, nil)
		_, err := factory.Build(original, true, RefundOverride{}, clock, ids)
		assert.ErrorIs(t, err, domain.ErrRefundAlreadyExists)
	})

	t.Run("income_refund_inherits_pix", func(t *testing.T) {
		t.Parallel()
		original := makeTx(t, vos.TransactionTypeIncome, vos.PaymentMethodPix, nil)
		refund, err := factory.Build(original, false, RefundOverride{}, clock, ids)
		require.NoError(t, err)
		assert.Equal(t, vos.TransactionTypeRefund, refund.TransactionType())
		assert.Equal(t, vos.PaymentMethodPix, refund.PaymentMethod())
		assert.Nil(t, refund.CardID())
		require.NotNil(t, refund.OriginalTransactionID())
		assert.Equal(t, original.ID(), *refund.OriginalTransactionID())
	})

	t.Run("credit_purchase_inherits_credit_card_and_card_id", func(t *testing.T) {
		t.Parallel()
		original := makeTx(t, vos.TransactionTypeCreditPurchase, vos.PaymentMethodCreditCard, &cardID)
		refund, err := factory.Build(original, false, RefundOverride{}, clock, ids)
		require.NoError(t, err)
		assert.Equal(t, vos.PaymentMethodCreditCard, refund.PaymentMethod())
		require.NotNil(t, refund.CardID())
		assert.Equal(t, cardID, *refund.CardID())
	})

	t.Run("override_payment_method_drops_card_id_when_non_card", func(t *testing.T) {
		t.Parallel()
		original := makeTx(t, vos.TransactionTypeCreditPurchase, vos.PaymentMethodCreditCard, &cardID)
		pm := vos.PaymentMethodPix
		refund, err := factory.Build(original, false, RefundOverride{PaymentMethod: &pm}, clock, ids)
		require.NoError(t, err)
		assert.Equal(t, vos.PaymentMethodPix, refund.PaymentMethod())
		assert.Nil(t, refund.CardID())
	})

	t.Run("override_description", func(t *testing.T) {
		t.Parallel()
		original := makeTx(t, vos.TransactionTypeExpense, vos.PaymentMethodPix, nil)
		desc := "Estorno manual"
		refund, err := factory.Build(original, false, RefundOverride{Description: &desc}, clock, ids)
		require.NoError(t, err)
		assert.Equal(t, "Estorno manual", refund.Description())
	})

	t.Run("refund_user_id_matches_original", func(t *testing.T) {
		t.Parallel()
		original := makeTx(t, vos.TransactionTypeExpense, vos.PaymentMethodCash, nil)
		refund, err := factory.Build(original, false, RefundOverride{}, clock, ids)
		require.NoError(t, err)
		assert.Equal(t, original.UserID(), refund.UserID())
	})
}

// Ensure seqIDs satisfies ports.IDGenerator at compile time.
var _ ports.IDGenerator = (*seqIDs)(nil)
