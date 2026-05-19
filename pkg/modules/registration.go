package modules

import (
	bootstrapcontainer "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/container"

	"github.com/gofiber/fiber/v2"
)

// CLIRunner defines the root command capabilities that modules can attach to.
type CLIRunner interface {
	RunServer() error
}

// ModuleRegistration declares how a module plugs into HTTP bootstrap.
type ModuleRegistration struct {
	Name         string
	RegisterHTTP func(router fiber.Router, container *bootstrapcontainer.Container)
}
