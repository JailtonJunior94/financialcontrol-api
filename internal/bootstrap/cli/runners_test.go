package cli

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/observability/noop"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	bootstrapcontainer "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/container"
	migrationmocks "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/migration/mocks"
)

// TestRunServer_GracefulShutdown verifies that runServer shuts down observability and db
// after the start function returns (ctx cancelled programmatically, no time.Sleep).
func TestRunServer_GracefulShutdown(t *testing.T) {
	mockMgr := migrationmocks.NewMockManager(t)
	mockMgr.On("Shutdown", mock.Anything).Return(nil)

	fakeContainer := &bootstrapcontainer.Container{
		DBManager:     mockMgr,
		Observability: noop.NewProvider(),
	}

	buildContainer := func(_ context.Context) (*bootstrapcontainer.Container, error) {
		return fakeContainer, nil
	}

	// start blocks until ctx is cancelled — simulates a running HTTP server.
	start := func(ctx context.Context, _ *bootstrapcontainer.Container) error {
		<-ctx.Done()
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := runServer(ctx, buildContainer, start)
	require.NoError(t, err)
}

// TestRunServer_FailsWhenContainerErrors verifies fail-fast: when the container build
// fails, the start function is never called and the error is propagated.
func TestRunServer_FailsWhenContainerErrors(t *testing.T) {
	buildErr := errors.New("db unreachable")

	buildContainer := func(_ context.Context) (*bootstrapcontainer.Container, error) {
		return nil, buildErr
	}

	start := func(_ context.Context, _ *bootstrapcontainer.Container) error {
		t.Fatal("start must not be called when container build fails")
		return nil
	}

	err := runServer(context.Background(), buildContainer, start)
	require.ErrorIs(t, err, buildErr)
}

// TestRunServer_AggregatesShutdownErrors verifies that errors from observability and db
// shutdown are joined and returned alongside the start error.
func TestRunServer_AggregatesShutdownErrors(t *testing.T) {
	dbErr := errors.New("db shutdown failed")

	mockMgr := migrationmocks.NewMockManager(t)
	mockMgr.On("Shutdown", mock.Anything).Return(dbErr)

	fakeContainer := &bootstrapcontainer.Container{
		DBManager:     mockMgr,
		Observability: noop.NewProvider(),
	}

	buildContainer := func(_ context.Context) (*bootstrapcontainer.Container, error) {
		return fakeContainer, nil
	}

	startErr := errors.New("server start error")
	start := func(_ context.Context, _ *bootstrapcontainer.Container) error {
		return startErr
	}

	err := runServer(context.Background(), buildContainer, start)
	require.ErrorIs(t, err, startErr)
	require.ErrorIs(t, err, dbErr)
}
