package vos

import domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain"

// TransactionType enumerates the supported transaction types.
type TransactionType string

const (
	TransactionTypeIncome              TransactionType = "income"
	TransactionTypeExpense             TransactionType = "expense"
	TransactionTypeCreditPurchase      TransactionType = "credit_purchase"
	TransactionTypeInstallmentPurchase TransactionType = "installment_purchase"
	TransactionTypeRefund              TransactionType = "refund"
)

var validTransactionTypes = map[TransactionType]struct{}{
	TransactionTypeIncome:              {},
	TransactionTypeExpense:             {},
	TransactionTypeCreditPurchase:      {},
	TransactionTypeInstallmentPurchase: {},
	TransactionTypeRefund:              {},
}

// ParseTransactionType parses a raw string into a TransactionType.
func ParseTransactionType(s string) (TransactionType, error) {
	t := TransactionType(s)
	if _, ok := validTransactionTypes[t]; !ok {
		return "", domain.ErrInvalidTransactionType
	}
	return t, nil
}

func (t TransactionType) String() string { return string(t) }

// IsCardBased reports whether the transaction type requires a card.
func (t TransactionType) IsCardBased() bool {
	return t == TransactionTypeCreditPurchase || t == TransactionTypeInstallmentPurchase
}

// IsRefund reports whether this type represents a refund.
func (t TransactionType) IsRefund() bool { return t == TransactionTypeRefund }
