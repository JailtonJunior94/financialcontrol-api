package vos

import domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain"

// PaymentMethod enumerates the supported payment methods.
type PaymentMethod string

const (
	PaymentMethodPix        PaymentMethod = "pix"
	PaymentMethodBoleto     PaymentMethod = "boleto"
	PaymentMethodTED        PaymentMethod = "ted"
	PaymentMethodDebitCard  PaymentMethod = "debit_card"
	PaymentMethodCreditCard PaymentMethod = "credit_card"
	PaymentMethodCash       PaymentMethod = "cash"
)

var validPaymentMethods = map[PaymentMethod]struct{}{
	PaymentMethodPix:        {},
	PaymentMethodBoleto:     {},
	PaymentMethodTED:        {},
	PaymentMethodDebitCard:  {},
	PaymentMethodCreditCard: {},
	PaymentMethodCash:       {},
}

// ParsePaymentMethod parses a raw string into a PaymentMethod.
func ParsePaymentMethod(s string) (PaymentMethod, error) {
	pm := PaymentMethod(s)
	if _, ok := validPaymentMethods[pm]; !ok {
		return "", domain.ErrInvalidPaymentMethod
	}
	return pm, nil
}

func (pm PaymentMethod) String() string { return string(pm) }

// RequiresCard reports whether the payment method requires a card reference.
func (pm PaymentMethod) RequiresCard() bool {
	return pm == PaymentMethodDebitCard || pm == PaymentMethodCreditCard
}
