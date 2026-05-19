package usecase_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/application/usecase"
	portmocks "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/ports"
	mocks "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/ports/mocks"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

func newMonthlySummaryUC(
	t *testing.T,
	txRepo *mocks.TransactionRepository,
	invRepo *mocks.InvoiceRepository,
	instRepo *mocks.InstallmentRepository,
) usecase.MonthlySummary {
	t.Helper()
	return usecase.NewMonthlySummary(txRepo, invRepo, instRepo)
}

func mustPeriod(t *testing.T) vos.Period {
	t.Helper()
	p, err := vos.NewPeriod(2026, 5, nil)
	if err != nil {
		t.Fatalf("mustPeriod: %v", err)
	}
	return p
}

func mustMoney(t *testing.T, s string) vos.Money {
	t.Helper()
	m, err := vos.NewMoney(s)
	if err != nil {
		t.Fatalf("mustMoney(%s): %v", s, err)
	}
	return m
}

func TestMonthlySummary_GoldenPath(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	period := mustPeriod(t)

	txRepo := mocks.NewTransactionRepository(t)
	invRepo := mocks.NewInvoiceRepository(t)
	instRepo := mocks.NewInstallmentRepository(t)

	agg := portmocks.SummaryAggregates{
		TotalIncome:     mustMoney(t, "1000.00"),
		TotalExpense:    mustMoney(t, "400.00"),
		TotalRefundsIn:  mustMoney(t, "50.00"),
		TotalRefundsOut: mustMoney(t, "20.00"),
	}

	txRepo.EXPECT().SumForSummary(mock.Anything, userID, period).Return(agg, nil)
	instRepo.EXPECT().SumByMonthCompetence(mock.Anything, userID, period).Return(mustMoney(t, "200.00"), nil)
	invRepo.EXPECT().SumOpenForUser(mock.Anything, userID, period).Return(mustMoney(t, "300.00"), nil)
	invRepo.EXPECT().SumPaidInPeriod(mock.Anything, userID, period).Return(mustMoney(t, "150.00"), nil)

	uc := newMonthlySummaryUC(t, txRepo, invRepo, instRepo)
	resp, err := uc.Execute(ctx, userID, period)

	require.NoError(t, err)
	assert.Equal(t, "1000.00", resp.TotalIncome)
	assert.Equal(t, "400.00", resp.TotalExpense)
	assert.Equal(t, "200.00", resp.TotalCreditPurchasesMonth)
	// balance = 1000 + 50 - 400 - 20 - 300 = 330
	assert.Equal(t, "330.00", resp.Balance)
}

func TestMonthlySummary_ErrorOnInstallmentSum(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	period := mustPeriod(t)

	txRepo := mocks.NewTransactionRepository(t)
	invRepo := mocks.NewInvoiceRepository(t)
	instRepo := mocks.NewInstallmentRepository(t)

	zero := vos.ZeroMoney()
	agg := portmocks.SummaryAggregates{
		TotalIncome:     zero,
		TotalExpense:    zero,
		TotalRefundsIn:  zero,
		TotalRefundsOut: zero,
	}

	sumErr := errTest("installment sum failed")
	txRepo.EXPECT().SumForSummary(mock.Anything, userID, period).Return(agg, nil)
	instRepo.EXPECT().SumByMonthCompetence(mock.Anything, userID, period).Return(zero, sumErr)

	uc := newMonthlySummaryUC(t, txRepo, invRepo, instRepo)
	_, err := uc.Execute(ctx, userID, period)

	assert.ErrorIs(t, err, sumErr)
}

func TestMonthlySummary_ErrorOnSumForSummary(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	period := mustPeriod(t)

	txRepo := mocks.NewTransactionRepository(t)
	invRepo := mocks.NewInvoiceRepository(t)
	instRepo := mocks.NewInstallmentRepository(t)

	zero := vos.ZeroMoney()
	sumErr := errTest("sum for summary failed")
	txRepo.EXPECT().SumForSummary(mock.Anything, userID, period).Return(portmocks.SummaryAggregates{}, sumErr)

	uc := newMonthlySummaryUC(t, txRepo, invRepo, instRepo)
	_, err := uc.Execute(ctx, userID, period)

	assert.ErrorIs(t, err, sumErr)
	instRepo.AssertNotCalled(t, "SumByMonthCompetence")
	_ = zero
}

func TestMonthlySummary_ZeroValues(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	period := mustPeriod(t)

	txRepo := mocks.NewTransactionRepository(t)
	invRepo := mocks.NewInvoiceRepository(t)
	instRepo := mocks.NewInstallmentRepository(t)

	zero := vos.ZeroMoney()
	agg := portmocks.SummaryAggregates{
		TotalIncome:     zero,
		TotalExpense:    zero,
		TotalRefundsIn:  zero,
		TotalRefundsOut: zero,
	}

	txRepo.EXPECT().SumForSummary(mock.Anything, userID, period).Return(agg, nil)
	instRepo.EXPECT().SumByMonthCompetence(mock.Anything, userID, period).Return(zero, nil)
	invRepo.EXPECT().SumOpenForUser(mock.Anything, userID, period).Return(zero, nil)
	invRepo.EXPECT().SumPaidInPeriod(mock.Anything, userID, period).Return(zero, nil)

	uc := newMonthlySummaryUC(t, txRepo, invRepo, instRepo)
	resp, err := uc.Execute(ctx, userID, period)

	require.NoError(t, err)
	assert.Equal(t, "0.00", resp.Balance)
}
