package container_test

import (
	"errors"
	"testing"

	bootstrapcontainer "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/container"
	pkgjwt "github.com/jailtonjunior94/financialcontrol-api/pkg/jwt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBuildRuntimeErrorOnMissingEnv validates that BuildRuntime propagates the
// config.Load error when ENVIRONMENT is absent.
func TestBuildRuntimeErrorOnMissingEnv(t *testing.T) {
	t.Setenv("ENVIRONMENT", "")

	c, err := bootstrapcontainer.BuildRuntime()

	require.Error(t, err)
	assert.Nil(t, c)
}

// TestBuildReturnsErrSecretTooShortForEmptyJWTSecret validates that Build
// propagates pkgjwt.ErrSecretTooShort instead of calling log.Fatalf.
func TestBuildReturnsErrSecretTooShortForEmptyJWTSecret(t *testing.T) {
	// config.JwtSecret is empty ("") at test time → NewIssuer returns ErrSecretTooShort.
	c, err := bootstrapcontainer.Build(nil)

	require.Error(t, err)
	assert.Nil(t, c)
	assert.True(t, errors.Is(err, pkgjwt.ErrSecretTooShort), "expected ErrSecretTooShort, got: %v", err)
}
