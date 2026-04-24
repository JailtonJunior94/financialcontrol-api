package cli

import (
	"time"

	bootstrapcontainer "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/container"
	bootstraphttp "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/http"
	planningapp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/planning/application"
	planninginfra "github.com/jailtonjunior94/financialcontrol-api/internal/modules/planning/infrastructure"
)

// AppRunners implements platform/modules.CLIRunner. Planning operations are
// delegated to PlanningRunner, which uses explicit ports and receives all
// dependencies through injection — no container is built inside each command.
// RunServer is handled separately because the HTTP bootstrap manages its own
// lifecycle.
type AppRunners struct {
	planning *planningapp.PlanningRunner
}

func NewAppRunners() *AppRunners {
	c := bootstrapcontainer.BuildRuntime()
	return &AppRunners{planning: buildPlanningRunner(c)}
}

// buildPlanningRunner wires the planning infrastructure adapters and the
// runner using repositories already created by the container. This keeps the
// container free of planning-specific imports (avoiding circular deps) while
// still reusing the single database connection and repository instances.
func buildPlanningRunner(c *bootstrapcontainer.Container) *planningapp.PlanningRunner {
	billing := planninginfra.NewBillingReadAdapter(c.BillRepository)
	invoicing := planninginfra.NewInvoicingReadAdapter(c.InvoiceRepository)
	sync := planninginfra.NewSyncAdapter(c.UpdateTransactionBill, c.UpdateUseCase)
	return planningapp.NewPlanningRunner(billing, invoicing, sync)
}

func (r *AppRunners) RunServer() error {
	return bootstraphttp.RunServer()
}

func (r *AppRunners) RunBudget(date time.Time) error {
	return r.planning.RunBudget(date)
}

func (r *AppRunners) RunBudgetCardsAndOthers(date time.Time) error {
	return r.planning.RunBudgetCardsAndOthers(date)
}

func (r *AppRunners) RunBudgetUnified(date time.Time) error {
	return r.planning.RunBudgetUnified(date)
}

func (r *AppRunners) RunBudgetFull(date time.Time) error {
	return r.planning.RunBudgetFull(date)
}

func (r *AppRunners) RunBalance(date time.Time) error {
	return r.planning.RunBalance(date)
}

func (r *AppRunners) RunBudgetByCategory(date time.Time, category string) error {
	return r.planning.RunBudgetByCategory(date, category)
}

func (r *AppRunners) RunSync() error {
	return r.planning.RunSync()
}
