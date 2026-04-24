package planning_test

import (
	"context"
	"testing"
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/planning"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeBillingRead is a simple in-memory implementation of BillingReadPort
// used to verify the port contract is satisfiable without concrete repos.
type fakeBillingRead struct {
	model *planning.MonthlyBillsReadModel
	err   error
}

func (f *fakeBillingRead) GetMonthlyBills(_ context.Context, referenceDate time.Time) (*planning.MonthlyBillsReadModel, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.model, nil
}

// fakeInvoicingRead is a simple in-memory implementation of InvoicingReadPort.
type fakeInvoicingRead struct {
	model *planning.MonthlyInvoicesReadModel
	err   error
}

func (f *fakeInvoicingRead) GetMonthlyInvoices(_ context.Context, referenceDate time.Time) (*planning.MonthlyInvoicesReadModel, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.model, nil
}

func TestBillingReadPortContractReturnsModel(t *testing.T) {
	refDate := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	expected := &planning.MonthlyBillsReadModel{
		ReferenceDate: refDate,
		Items: []planning.BillReadItem{
			{ID: "b1", Description: "Rent", Total: 1500},
		},
	}
	var port planning.BillingReadPort = &fakeBillingRead{model: expected}

	result, err := port.GetMonthlyBills(context.Background(), refDate)

	require.NoError(t, err)
	assert.Equal(t, expected.ReferenceDate, result.ReferenceDate)
	require.Len(t, result.Items, 1)
	assert.Equal(t, "Rent", result.Items[0].Description)
}

func TestInvoicingReadPortContractReturnsModel(t *testing.T) {
	refDate := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	expected := &planning.MonthlyInvoicesReadModel{
		ReferenceDate: refDate,
		Items: []planning.InvoiceReadItem{
			{ID: "i1", Description: "Card X", Total: 800, Date: refDate, Tags: "Prazeres", Category: "Restaurante"},
		},
	}
	var port planning.InvoicingReadPort = &fakeInvoicingRead{model: expected}

	result, err := port.GetMonthlyInvoices(context.Background(), refDate)

	require.NoError(t, err)
	assert.Equal(t, expected.ReferenceDate, result.ReferenceDate)
	require.Len(t, result.Items, 1)
	assert.Equal(t, float64(800), result.Items[0].Total)
	assert.Equal(t, "Prazeres", result.Items[0].Tags)
	assert.Equal(t, "Restaurante", result.Items[0].Category)
}

func TestBillingReadPortContractPropagatesError(t *testing.T) {
	var port planning.BillingReadPort = &fakeBillingRead{err: context.DeadlineExceeded}
	_, err := port.GetMonthlyBills(context.Background(), time.Now())
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestInvoicingReadPortContractPropagatesError(t *testing.T) {
	var port planning.InvoicingReadPort = &fakeInvoicingRead{err: context.DeadlineExceeded}
	_, err := port.GetMonthlyInvoices(context.Background(), time.Now())
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}
