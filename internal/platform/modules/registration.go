package modules

import (
	"time"

	bootstrapcontainer "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/container"

	"github.com/gofiber/fiber/v2"
	"github.com/spf13/cobra"
)

// CLIRunner defines the root command capabilities that modules can attach to.
type CLIRunner interface {
	RunServer() error
	RunBudget(date time.Time) error
	RunBudgetCardsAndOthers(date time.Time) error
	RunBudgetUnified(date time.Time) error
	RunBudgetFull(date time.Time) error
	RunBalance(date time.Time) error
	RunBudgetByCategory(date time.Time, category string) error
	RunSync() error
}

// ModuleRegistration declares how a module plugs into HTTP and CLI bootstrap.
type ModuleRegistration struct {
	Name         string
	RegisterHTTP func(router fiber.Router, container *bootstrapcontainer.Container)
	RegisterCLI  func(root *cobra.Command, runners CLIRunner)
}
