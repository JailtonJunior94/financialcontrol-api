package usecase

import (
	"context"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/application/dtos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/ports"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

type monthlySummary struct {
	txRepo   ports.TransactionRepository
	invRepo  ports.InvoiceRepository
	instRepo ports.InstallmentRepository
}

// NewMonthlySummary constructs the MonthlySummary use case.
func NewMonthlySummary(
	txRepo ports.TransactionRepository,
	invRepo ports.InvoiceRepository,
	instRepo ports.InstallmentRepository,
) MonthlySummary {
	return &monthlySummary{
		txRepo:   txRepo,
		invRepo:  invRepo,
		instRepo: instRepo,
	}
}

func (uc *monthlySummary) Execute(ctx context.Context, userID identityvo.UserID, period vos.Period) (dtos.MonthlySummaryResponse, error) {
	aggregates, err := uc.txRepo.SumForSummary(ctx, userID, period)
	if err != nil {
		return dtos.MonthlySummaryResponse{}, err
	}

	creditPurchases, err := uc.instRepo.SumByMonthCompetence(ctx, userID, period)
	if err != nil {
		return dtos.MonthlySummaryResponse{}, err
	}

	invoicesOpen, err := uc.invRepo.SumOpenForUser(ctx, userID, period)
	if err != nil {
		return dtos.MonthlySummaryResponse{}, err
	}

	invoicesPaid, err := uc.invRepo.SumPaidInPeriod(ctx, userID, period)
	if err != nil {
		return dtos.MonthlySummaryResponse{}, err
	}

	// balance = total_income + total_refunds_in − total_expense − total_refunds_out − total_invoices_open
	balance := aggregates.TotalIncome.
		Add(aggregates.TotalRefundsIn).
		Sub(aggregates.TotalExpense).
		Sub(aggregates.TotalRefundsOut).
		Sub(invoicesOpen)

	zero := vos.ZeroMoney()
	return dtos.MonthlySummaryResponse{
		TotalIncome:               stringify(aggregates.TotalIncome, zero),
		TotalExpense:              stringify(aggregates.TotalExpense, zero),
		TotalRefundsIn:            stringify(aggregates.TotalRefundsIn, zero),
		TotalRefundsOut:           stringify(aggregates.TotalRefundsOut, zero),
		TotalCreditPurchasesMonth: stringify(creditPurchases, zero),
		TotalInvoicesOpen:         stringify(invoicesOpen, zero),
		TotalInvoicesPaid:         stringify(invoicesPaid, zero),
		Balance:                   balance.String(),
	}, nil
}

func stringify(m, _ vos.Money) string { return m.String() }
