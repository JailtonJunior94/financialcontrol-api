package dtos

import (
	"time"

	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain"
)

// PayInvoiceRequest carries the payload for PATCH /finance/invoices/{id}/pay (RF-18).
type PayInvoiceRequest struct {
	PaymentMethod                 *string `json:"payment_method,omitempty"`
	GenerateSettlementTransaction bool    `json:"generate_settlement_transaction"`
	SettlementCategoryID          *string `json:"settlement_category_id,omitempty"`
}

// Validate checks that payment_method is present when generate_settlement_transaction is true.
func (r PayInvoiceRequest) Validate() error {
	if r.GenerateSettlementTransaction && r.PaymentMethod == nil {
		return domain.ErrInvalidPaymentMethod
	}
	return nil
}

// InvoiceResponse is the read representation of an invoice aggregate.
type InvoiceResponse struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	CardID      string     `json:"card_id"`
	State       string     `json:"state"`
	CycleStart  time.Time  `json:"cycle_start"`
	CycleEnd    time.Time  `json:"cycle_end"`
	ClosingDate time.Time  `json:"closing_date"`
	DueDate     time.Time  `json:"due_date"`
	Total       string     `json:"total"`
	PaidAt      *time.Time `json:"paid_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// InvoiceItemResponse represents a single installment line within an invoice detail.
type InvoiceItemResponse struct {
	InstallmentID string `json:"installment_id"`
	TransactionID string `json:"transaction_id"`
	Number        int    `json:"number"`
	Total         int    `json:"total"`
	Amount        string `json:"amount"`
	Status        string `json:"status"`
}

// InvoiceDetailResponse extends InvoiceResponse with the full installment list (RF-21).
type InvoiceDetailResponse struct {
	InvoiceResponse
	Items []InvoiceItemResponse `json:"items"`
}
