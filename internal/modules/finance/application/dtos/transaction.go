package dtos

import (
	"time"

	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain"
)

// CreateTransactionRequest carries the payload for POST /finance/transactions (RF-01..04, RF-26).
type CreateTransactionRequest struct {
	Description      string    `json:"description"`
	Amount           string    `json:"amount"`
	OccurredAt       time.Time `json:"occurred_at"`
	TransactionType  string    `json:"transaction_type"`
	PaymentMethod    string    `json:"payment_method"`
	CardID           *string   `json:"card_id,omitempty"`
	CategoryID       string    `json:"category_id"`
	SubcategoryID    *string   `json:"subcategory_id,omitempty"`
	InstallmentCount int       `json:"installment_count,omitempty"`
}

// Validate checks cross-field invariants enforced at the HTTP boundary (G3.a).
func (r CreateTransactionRequest) Validate() error {
	if r.SubcategoryID != nil && *r.SubcategoryID == r.CategoryID {
		return domain.ErrSubcategoryEqualsCategory
	}
	return nil
}

// UpdateTransactionRequest carries the payload for PUT /finance/transactions/{id} (RF-54).
type UpdateTransactionRequest struct {
	Description      string    `json:"description"`
	Amount           string    `json:"amount"`
	OccurredAt       time.Time `json:"occurred_at"`
	TransactionType  string    `json:"transaction_type"`
	PaymentMethod    string    `json:"payment_method"`
	CardID           *string   `json:"card_id,omitempty"`
	CategoryID       string    `json:"category_id"`
	SubcategoryID    *string   `json:"subcategory_id,omitempty"`
	InstallmentCount int       `json:"installment_count,omitempty"`
}

// Validate checks cross-field invariants enforced at the HTTP boundary (G3.a).
func (r UpdateTransactionRequest) Validate() error {
	if r.SubcategoryID != nil && *r.SubcategoryID == r.CategoryID {
		return domain.ErrSubcategoryEqualsCategory
	}
	return nil
}

// RefundTransactionRequest carries the optional overrides for POST /finance/transactions/{id}/refund (RF-53).
type RefundTransactionRequest struct {
	PaymentMethod *string `json:"payment_method,omitempty"`
	Description   *string `json:"description,omitempty"`
}

// InstallmentResponse is the read representation of a single installment.
type InstallmentResponse struct {
	ID            string    `json:"id"`
	TransactionID string    `json:"transaction_id"`
	InvoiceID     string    `json:"invoice_id"`
	Number        int       `json:"number"`
	Total         int       `json:"total"`
	Amount        string    `json:"amount"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// TransactionResponse is the read representation of a transaction aggregate.
type TransactionResponse struct {
	ID                    string                `json:"id"`
	Description           string                `json:"description"`
	Amount                string                `json:"amount"`
	OccurredAt            time.Time             `json:"occurred_at"`
	TransactionType       string                `json:"transaction_type"`
	PaymentMethod         string                `json:"payment_method"`
	CardID                *string               `json:"card_id,omitempty"`
	CategoryID            string                `json:"category_id"`
	SubcategoryID         *string               `json:"subcategory_id,omitempty"`
	OriginalTransactionID *string               `json:"original_transaction_id,omitempty"`
	Installments          []InstallmentResponse `json:"installments,omitempty"`
	LegacyOrigin          *string               `json:"legacy_origin,omitempty"`
	CreatedAt             time.Time             `json:"created_at"`
	UpdatedAt             time.Time             `json:"updated_at"`
}
