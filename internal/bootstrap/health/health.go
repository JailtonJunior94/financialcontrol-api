package health

import (
	"context"

	"github.com/JailtonJunior94/devkit-go/pkg/database/manager"
	serverfiber "github.com/JailtonJunior94/devkit-go/pkg/http_server/server_fiber"
)

// Checks is a first-class collection of named health check functions.
// The "mssql" check is always present after construction (invariant).
type Checks struct {
	items map[string]serverfiber.HealthCheckFunc
}

// New builds a Checks collection with the mandatory "mssql" probe.
// The probe delegates to mgr.Ping which respects the ctx timeout set by the lib.
func New(mgr manager.Manager) Checks {
	c := Checks{items: make(map[string]serverfiber.HealthCheckFunc)}
	c.items["mssql"] = func(ctx context.Context) error {
		return mgr.Ping(ctx)
	}
	return c
}

// Map returns the named health check functions ready for WithHealthChecks.
func (c Checks) Map() map[string]serverfiber.HealthCheckFunc {
	return c.items
}
