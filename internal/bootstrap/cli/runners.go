package cli

import (
	"errors"
	"time"

	bootstrapcontainer "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/container"
	bootstraphttp "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/http"
)

// AppRunners implements platform/modules.CLIRunner.
// Planning CLI commands were removed together with the billing, transactions,
// invoicing, and planning/sync modules in task 9.0. All planning methods return
// errPlanningNotAvailable to satisfy the interface; they are never invoked because
// RegisterCLI no longer adds the corresponding cobra commands.
type AppRunners struct{}

var errPlanningNotAvailable = errors.New("planning CLI commands removed in task 9.0: use finance endpoints instead")

func NewAppRunners() (*AppRunners, error) {
	_, err := bootstrapcontainer.BuildRuntime()
	if err != nil {
		return nil, err
	}
	return &AppRunners{}, nil
}

func (r *AppRunners) RunServer() error {
	return bootstraphttp.RunServer()
}

func (r *AppRunners) RunBudget(_ time.Time) error {
	return errPlanningNotAvailable
}

func (r *AppRunners) RunBudgetCardsAndOthers(_ time.Time) error {
	return errPlanningNotAvailable
}

func (r *AppRunners) RunBudgetUnified(_ time.Time) error {
	return errPlanningNotAvailable
}

func (r *AppRunners) RunBudgetFull(_ time.Time) error {
	return errPlanningNotAvailable
}

func (r *AppRunners) RunBalance(_ time.Time) error {
	return errPlanningNotAvailable
}

func (r *AppRunners) RunBudgetByCategory(_ time.Time, _ string) error {
	return errPlanningNotAvailable
}

func (r *AppRunners) RunSync() error {
	return errPlanningNotAvailable
}
