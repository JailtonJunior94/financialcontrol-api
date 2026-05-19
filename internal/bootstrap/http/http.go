package http

import (
	"fmt"
	"os"

	bootstrapcontainer "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/container"
	bootstraphealth "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/health"

	serverfiber "github.com/JailtonJunior94/devkit-go/pkg/http_server/server_fiber"
)

// NewServer mounts *serverfiber.Server with middlewares, observability, health checks
// and the /api/v1 router. It does not call Start; lifecycle is owned by the caller
// (see internal/bootstrap/cli/runners.go).
func NewServer(c *bootstrapcontainer.Container) (*serverfiber.Server, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	srv, err := serverfiber.New(
		c.Observability,
		serverfiber.WithPort(port),
		serverfiber.WithCORS("*"),
		serverfiber.WithTracing(),
		serverfiber.WithOTelMetrics(),
		serverfiber.WithShutdownTimeout(c.ShutdownTimeout.Duration()),
		serverfiber.WithHealthChecks(bootstraphealth.New(c.DBManager).Map()),
		serverfiber.WithServiceName(c.Identity.Name()),
		serverfiber.WithServiceVersion(c.Identity.Version()),
		serverfiber.WithEnvironment(c.Identity.Environment()),
	)
	if err != nil {
		return nil, fmt.Errorf("bootstrap http: %w", err)
	}

	srv.RegisterRouters(newAPIV1Router(c))
	return srv, nil
}
