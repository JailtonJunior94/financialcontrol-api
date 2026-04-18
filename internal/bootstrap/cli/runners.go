package cli

import (
	"time"

	bootstraphttp "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/http"
)

type AppRunners struct{}

func NewAppRunners() *AppRunners {
	return &AppRunners{}
}

func (r *AppRunners) RunServer() error {
	return bootstraphttp.RunServer()
}

func (r *AppRunners) RunBudget(date time.Time) error {
	RunBudget(date)
	return nil
}

func (r *AppRunners) RunBudgetCardsAndOthers(date time.Time) error {
	RunBudgetCardAndOthers(date)
	return nil
}

func (r *AppRunners) RunBudgetUnified(date time.Time) error {
	RunBudgetUnified(date)
	return nil
}

func (r *AppRunners) RunBudgetFull(date time.Time) error {
	RunBudgetFullView(date)
	return nil
}

func (r *AppRunners) RunBalance(date time.Time) error {
	RunBalance(date)
	return nil
}

func (r *AppRunners) RunBudgetByCategory(date time.Time, category string) error {
	RunBudgetByCategory(date, category)
	return nil
}

func (r *AppRunners) RunSync() error {
	RunSync()
	return nil
}
