package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/planning"
	planningapp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/planning/application"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- fakes ---

type fakeBillingPort struct {
	model *planning.MonthlyBillsReadModel
	err   error
}

func (f *fakeBillingPort) GetMonthlyBills(_ context.Context, referenceDate time.Time) (*planning.MonthlyBillsReadModel, error) {
	if f.err != nil {
		return nil, f.err
	}
	if f.model != nil {
		return f.model, nil
	}
	return &planning.MonthlyBillsReadModel{ReferenceDate: referenceDate}, nil
}

type fakeInvoicingPort struct {
	model *planning.MonthlyInvoicesReadModel
	err   error
}

func (f *fakeInvoicingPort) GetMonthlyInvoices(_ context.Context, referenceDate time.Time) (*planning.MonthlyInvoicesReadModel, error) {
	if f.err != nil {
		return nil, f.err
	}
	if f.model != nil {
		return f.model, nil
	}
	return &planning.MonthlyInvoicesReadModel{ReferenceDate: referenceDate}, nil
}

type fakeSyncPort struct {
	called bool
	err    error
}

func (f *fakeSyncPort) Sync() error {
	f.called = true
	return f.err
}

// --- helpers ---

func newRunner(billing planning.BillingReadPort, invoicing planning.InvoicingReadPort, sync planning.SyncPort) *planningapp.PlanningRunner {
	return planningapp.NewPlanningRunner(billing, invoicing, sync)
}

func defaultBills(date time.Time) *planning.MonthlyBillsReadModel {
	return &planning.MonthlyBillsReadModel{
		ReferenceDate: date,
		Items: []planning.BillReadItem{
			{ID: "b1", Description: "Faxina casa", Total: 200},
			{ID: "b2", Description: "Internet", Total: 100},
			{ID: "b3", Description: "Energia", Total: 150},
		},
	}
}

func defaultInvoices(date time.Time) *planning.MonthlyInvoicesReadModel {
	return &planning.MonthlyInvoicesReadModel{
		ReferenceDate: date,
		Items: []planning.InvoiceReadItem{
			{ID: "i1", Description: "Restaurante", Total: 300, Tags: "Prazeres", Date: date},
			{ID: "i2", Description: "Livro Go", Total: 80, Tags: "Conhecimento", Date: date},
			{ID: "i3", Description: "Mercado", Total: 500, Tags: "Custos fixos", Date: date, Category: "Supermercado"},
		},
	}
}

// --- tests ---

func TestPlanningRunnerRunBudgetReturnsNoErrorWithValidPorts(t *testing.T) {
	date := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	r := newRunner(
		&fakeBillingPort{model: defaultBills(date)},
		&fakeInvoicingPort{model: defaultInvoices(date)},
		&fakeSyncPort{},
	)

	err := r.RunBudget(date)
	require.NoError(t, err)
}

func TestPlanningRunnerRunBudgetPropagatesBillingError(t *testing.T) {
	date := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	expected := errors.New("billing unavailable")
	r := newRunner(
		&fakeBillingPort{err: expected},
		&fakeInvoicingPort{},
		&fakeSyncPort{},
	)

	err := r.RunBudget(date)
	require.Error(t, err)
	assert.ErrorIs(t, err, expected)
}

func TestPlanningRunnerRunBudgetPropagatesInvoicingError(t *testing.T) {
	date := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	expected := errors.New("invoicing unavailable")
	r := newRunner(
		&fakeBillingPort{model: defaultBills(date)},
		&fakeInvoicingPort{err: expected},
		&fakeSyncPort{},
	)

	err := r.RunBudget(date)
	require.Error(t, err)
	assert.ErrorIs(t, err, expected)
}

func TestPlanningRunnerRunBudgetCardsAndOthersNoError(t *testing.T) {
	date := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	r := newRunner(
		&fakeBillingPort{model: defaultBills(date)},
		&fakeInvoicingPort{model: defaultInvoices(date)},
		&fakeSyncPort{},
	)

	err := r.RunBudgetCardsAndOthers(date)
	require.NoError(t, err)
}

func TestPlanningRunnerRunBudgetUnifiedNoError(t *testing.T) {
	date := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	r := newRunner(
		&fakeBillingPort{model: defaultBills(date)},
		&fakeInvoicingPort{model: defaultInvoices(date)},
		&fakeSyncPort{},
	)

	err := r.RunBudgetUnified(date)
	require.NoError(t, err)
}

func TestPlanningRunnerRunBudgetFullNoError(t *testing.T) {
	date := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	r := newRunner(
		&fakeBillingPort{model: defaultBills(date)},
		&fakeInvoicingPort{model: defaultInvoices(date)},
		&fakeSyncPort{},
	)

	err := r.RunBudgetFull(date)
	require.NoError(t, err)
}

func TestPlanningRunnerRunBalanceNoError(t *testing.T) {
	date := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	r := newRunner(
		&fakeBillingPort{model: defaultBills(date)},
		&fakeInvoicingPort{model: defaultInvoices(date)},
		&fakeSyncPort{},
	)

	err := r.RunBalance(date)
	require.NoError(t, err)
}

func TestPlanningRunnerRunBudgetByCategoryNoError(t *testing.T) {
	date := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	r := newRunner(
		&fakeBillingPort{model: defaultBills(date)},
		&fakeInvoicingPort{model: defaultInvoices(date)},
		&fakeSyncPort{},
	)

	err := r.RunBudgetByCategory(date, "Prazeres")
	require.NoError(t, err)
}

func TestPlanningRunnerRunBudgetByCategoryConfortoIncludesBills(t *testing.T) {
	date := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	r := newRunner(
		&fakeBillingPort{model: defaultBills(date)},
		&fakeInvoicingPort{model: defaultInvoices(date)},
		&fakeSyncPort{},
	)

	// Should not error even with Conforto category (includes bill filter logic)
	err := r.RunBudgetByCategory(date, "Conforto")
	require.NoError(t, err)
}

func TestPlanningRunnerRunSyncDelegatesToSyncPort(t *testing.T) {
	date := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	syncPort := &fakeSyncPort{}
	r := newRunner(
		&fakeBillingPort{model: defaultBills(date)},
		&fakeInvoicingPort{model: defaultInvoices(date)},
		syncPort,
	)

	err := r.RunSync()
	require.NoError(t, err)
	assert.True(t, syncPort.called)
}

func TestPlanningRunnerRunSyncPropagatesSyncError(t *testing.T) {
	date := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	expected := errors.New("sync failed")
	r := newRunner(
		&fakeBillingPort{model: defaultBills(date)},
		&fakeInvoicingPort{model: defaultInvoices(date)},
		&fakeSyncPort{err: expected},
	)

	err := r.RunSync()
	require.Error(t, err)
	assert.ErrorIs(t, err, expected)
}

func TestPlanningRunnerRunBudgetWithEmptyPortsNoError(t *testing.T) {
	// Both ports return empty models — verifies the runner handles nil/empty data gracefully.
	date := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	r := newRunner(&fakeBillingPort{}, &fakeInvoicingPort{}, &fakeSyncPort{})

	err := r.RunBudget(date)
	require.NoError(t, err)
}
