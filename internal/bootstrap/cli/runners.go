package cli

import (
	"context"
	"errors"
	"fmt"
	"time"

	bootstrapcontainer "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/container"
	bootstraphttp "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/http"
)

// AppRunners implements platform/modules.CLIRunner. The only CLI capability is
// running the HTTP server; legacy budget/sync commands were removed together
// with the planning module.
type AppRunners struct{}

func NewAppRunners() (*AppRunners, error) {
	return &AppRunners{}, nil
}

func (r *AppRunners) RunServer() error {
	return RunServer(context.Background())
}

// RunServer builds the runtime container, mounts the HTTP server and starts it.
// Graceful shutdown executes in order: server (via Start) → observability → db.
// The caller must pass a context tied to OS signals so cancellation propagates correctly.
func RunServer(ctx context.Context) error {
	return runServer(ctx, bootstrapcontainer.BuildRuntime, func(ctx context.Context, c *bootstrapcontainer.Container) error {
		srv, err := bootstraphttp.NewServer(c)
		if err != nil {
			return fmt.Errorf("http server: %w", err)
		}
		return srv.Start(ctx)
	})
}

// runServer is the testable core. It accepts factory functions so unit tests can
// inject mock containers and controlled start functions without real infrastructure.
func runServer(
	ctx context.Context,
	buildContainer func(context.Context) (*bootstrapcontainer.Container, error),
	start func(context.Context, *bootstrapcontainer.Container) error,
) error {
	c, err := buildContainer(ctx)
	if err != nil {
		return fmt.Errorf("run server: %w", err)
	}

	startErr := start(ctx, c)

	timeout := c.ShutdownTimeout.Duration()
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	obsErr := c.Observability.Shutdown(shutdownCtx)
	dbErr := c.DBManager.Shutdown(shutdownCtx)

	return errors.Join(startErr, obsErr, dbErr)
}
