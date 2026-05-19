package usecase

import (
	"context"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/application/dtos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/filters"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

type CreateTransaction interface {
	Execute(ctx context.Context, userID identityvo.UserID, key vos.IdempotencyKey, req dtos.CreateTransactionRequest) (dtos.TransactionResponse, error)
}

type ListTransactions interface {
	Execute(ctx context.Context, userID identityvo.UserID, f filters.TransactionFilter) (dtos.PaginatedResponse[dtos.TransactionResponse], error)
}

type GetTransaction interface {
	Execute(ctx context.Context, userID identityvo.UserID, txID vos.TransactionID) (dtos.TransactionResponse, error)
}

type UpdateTransaction interface {
	Execute(ctx context.Context, userID identityvo.UserID, txID vos.TransactionID, req dtos.UpdateTransactionRequest) (dtos.TransactionResponse, error)
}

type DeleteTransaction interface {
	Execute(ctx context.Context, userID identityvo.UserID, txID vos.TransactionID) error
}

type RefundTransaction interface {
	Execute(ctx context.Context, userID identityvo.UserID, txID vos.TransactionID, req dtos.RefundTransactionRequest) (dtos.TransactionResponse, error)
}

type ListInvoices interface {
	Execute(ctx context.Context, userID identityvo.UserID, f filters.InvoiceFilter) (dtos.PaginatedResponse[dtos.InvoiceResponse], error)
}

type GetInvoice interface {
	Execute(ctx context.Context, userID identityvo.UserID, invoiceID vos.InvoiceID) (dtos.InvoiceDetailResponse, error)
}

type PayInvoice interface {
	Execute(ctx context.Context, userID identityvo.UserID, invoiceID vos.InvoiceID, req dtos.PayInvoiceRequest) (dtos.InvoiceResponse, error)
}

type AnticipateInstallment interface {
	Execute(ctx context.Context, userID identityvo.UserID, txID vos.TransactionID, installmentID vos.InstallmentID) (dtos.TransactionResponse, error)
}

type MonthlySummary interface {
	Execute(ctx context.Context, userID identityvo.UserID, period vos.Period) (dtos.MonthlySummaryResponse, error)
}
