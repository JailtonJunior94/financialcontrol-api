package observability_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bootstrapconfig "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/config"
	bootstrapobs "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/observability"
)

func makeIdentity(t *testing.T, name, version, env string) bootstrapconfig.ServiceIdentity {
	t.Helper()
	t.Setenv("SERVICE_NAME", name)
	t.Setenv("SERVICE_VERSION", version)
	t.Setenv("ENVIRONMENT", env)
	id, err := bootstrapconfig.NewServiceIdentity()
	require.NoError(t, err)
	return id
}

func TestLoadSettings_Protocol(t *testing.T) {
	tests := []struct {
		name         string
		protocolEnv  string
		endpointEnv  string
		wantProtocol string
		wantEndpoint string
		wantErr      error
	}{
		{
			name:         "default to grpc when env not set",
			protocolEnv:  "",
			endpointEnv:  "",
			wantProtocol: "grpc",
			wantEndpoint: "localhost:4317",
		},
		{
			name:         "grpc explicit uses grpc default endpoint",
			protocolEnv:  "grpc",
			endpointEnv:  "",
			wantProtocol: "grpc",
			wantEndpoint: "localhost:4317",
		},
		{
			name:         "http/protobuf explicit uses http default endpoint",
			protocolEnv:  "http/protobuf",
			endpointEnv:  "",
			wantProtocol: "http/protobuf",
			wantEndpoint: "localhost:4318",
		},
		{
			name:         "custom endpoint overrides grpc default",
			protocolEnv:  "grpc",
			endpointEnv:  "otel-collector:4317",
			wantProtocol: "grpc",
			wantEndpoint: "otel-collector:4317",
		},
		{
			name:         "custom endpoint overrides http default",
			protocolEnv:  "http/protobuf",
			endpointEnv:  "otel-collector:4318",
			wantProtocol: "http/protobuf",
			wantEndpoint: "otel-collector:4318",
		},
		{
			name:        "invalid protocol returns ErrInvalidOTLPProtocol",
			protocolEnv: "otlp-grpc",
			endpointEnv: "",
			wantErr:     bootstrapobs.ErrInvalidOTLPProtocol,
		},
		{
			name:        "stdout is no longer a valid protocol",
			protocolEnv: "stdout",
			endpointEnv: "",
			wantErr:     bootstrapobs.ErrInvalidOTLPProtocol,
		},
		{
			name:        "jaeger is invalid protocol",
			protocolEnv: "jaeger",
			endpointEnv: "",
			wantErr:     bootstrapobs.ErrInvalidOTLPProtocol,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			id := makeIdentity(t, "test-svc", "1.0.0", "development")
			t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", tc.protocolEnv)
			t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", tc.endpointEnv)

			s, err := bootstrapobs.LoadSettings(id)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tc.wantErr))
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.wantProtocol, s.OTLPProtocol)
			assert.Equal(t, tc.wantEndpoint, s.Endpoint)
		})
	}
}

func TestLoadSettings_LogLevel(t *testing.T) {
	tests := []struct {
		name        string
		logLevelEnv string
		wantLevel   string
		wantErr     error
	}{
		{
			name:        "default to info when LOG_LEVEL not set",
			logLevelEnv: "",
			wantLevel:   "info",
		},
		{
			name:        "debug level accepted",
			logLevelEnv: "debug",
			wantLevel:   "debug",
		},
		{
			name:        "warn level accepted",
			logLevelEnv: "warn",
			wantLevel:   "warn",
		},
		{
			name:        "error level accepted",
			logLevelEnv: "error",
			wantLevel:   "error",
		},
		{
			name:        "invalid level returns ErrInvalidLogLevel",
			logLevelEnv: "trace",
			wantErr:     bootstrapobs.ErrInvalidLogLevel,
		},
		{
			name:        "verbose is invalid log level",
			logLevelEnv: "verbose",
			wantErr:     bootstrapobs.ErrInvalidLogLevel,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			id := makeIdentity(t, "test-svc", "1.0.0", "development")
			t.Setenv("LOG_LEVEL", tc.logLevelEnv)
			t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "grpc")

			s, err := bootstrapobs.LoadSettings(id)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tc.wantErr))
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.wantLevel, string(s.LogLevel))
		})
	}
}

func TestLoadSettings_SlowQueryThreshold(t *testing.T) {
	tests := []struct {
		name      string
		envValue  string
		wantMS    int64
	}{
		{
			name:     "default 1000ms when not set",
			envValue: "",
			wantMS:   1000,
		},
		{
			name:     "custom value 500ms",
			envValue: "500",
			wantMS:   500,
		},
		{
			name:     "custom value 5000ms",
			envValue: "5000",
			wantMS:   5000,
		},
		{
			name:     "invalid non-numeric falls back to default",
			envValue: "notanumber",
			wantMS:   1000,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			id := makeIdentity(t, "test-svc", "1.0.0", "development")
			t.Setenv("SLOW_QUERY_THRESHOLD_MS", tc.envValue)
			t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "grpc")
			t.Setenv("LOG_LEVEL", "info")

			s, err := bootstrapobs.LoadSettings(id)

			require.NoError(t, err)
			assert.Equal(t, tc.wantMS, s.SlowQueryThresholdMS)
		})
	}
}

func TestLoadSettings_TracesSampler(t *testing.T) {
	tests := []struct {
		name        string
		samplerEnv  string
		argEnv      string
		wantSampler string
		wantArg     string
	}{
		{
			name:        "default sampler when not set",
			samplerEnv:  "",
			argEnv:      "",
			wantSampler: "parentbased_always_on",
			wantArg:     "",
		},
		{
			name:        "custom sampler with arg",
			samplerEnv:  "traceidratio",
			argEnv:      "0.5",
			wantSampler: "traceidratio",
			wantArg:     "0.5",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			id := makeIdentity(t, "test-svc", "1.0.0", "development")
			t.Setenv("OTEL_TRACES_SAMPLER", tc.samplerEnv)
			t.Setenv("OTEL_TRACES_SAMPLER_ARG", tc.argEnv)
			t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "grpc")
			t.Setenv("LOG_LEVEL", "info")

			s, err := bootstrapobs.LoadSettings(id)

			require.NoError(t, err)
			assert.Equal(t, tc.wantSampler, s.TracesSampler)
			assert.Equal(t, tc.wantArg, s.TracesSamplerArg)
		})
	}
}

func TestLoadSettings_PropagatesServiceIdentity(t *testing.T) {
	id := makeIdentity(t, "my-service", "2.5.0", "staging")
	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "grpc")
	t.Setenv("LOG_LEVEL", "info")

	s, err := bootstrapobs.LoadSettings(id)

	require.NoError(t, err)
	assert.Equal(t, "my-service", s.Identity.Name())
	assert.Equal(t, "2.5.0", s.Identity.Version())
	assert.Equal(t, "staging", s.Identity.Environment())
}

// TestLoadSettings_K8sAttrs_Absent verifies that missing K8s downward-API env vars
// cause all K8sAttrs fields to fall back to "unknown" without returning an error.
func TestLoadSettings_K8sAttrs_Absent(t *testing.T) {
	id := makeIdentity(t, "test-svc", "1.0.0", "development")
	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "grpc")
	t.Setenv("LOG_LEVEL", "info")
	t.Setenv("POD_UID", "")
	t.Setenv("POD_NAME", "")
	t.Setenv("POD_NAMESPACE", "")
	t.Setenv("CONTAINER_NAME", "")

	s, err := bootstrapobs.LoadSettings(id)

	require.NoError(t, err)
	assert.Equal(t, "unknown", s.K8s.PodUID)
	assert.Equal(t, "unknown", s.K8s.PodName)
	assert.Equal(t, "unknown", s.K8s.PodNamespace)
	assert.Equal(t, "unknown", s.K8s.ContainerName)
}

// TestLoadSettings_K8sAttrs_Present verifies that set K8s downward-API env vars
// are propagated verbatim to K8sAttrs.
func TestLoadSettings_K8sAttrs_Present(t *testing.T) {
	id := makeIdentity(t, "test-svc", "1.0.0", "development")
	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "grpc")
	t.Setenv("LOG_LEVEL", "info")
	t.Setenv("POD_UID", "uid-abc1234")
	t.Setenv("POD_NAME", "financialapi-pod-xyz")
	t.Setenv("POD_NAMESPACE", "financialcontrol")
	t.Setenv("CONTAINER_NAME", "financialapi")

	s, err := bootstrapobs.LoadSettings(id)

	require.NoError(t, err)
	assert.Equal(t, "uid-abc1234", s.K8s.PodUID)
	assert.Equal(t, "financialapi-pod-xyz", s.K8s.PodName)
	assert.Equal(t, "financialcontrol", s.K8s.PodNamespace)
	assert.Equal(t, "financialapi", s.K8s.ContainerName)
}

// TestLoadSettings_OTELExporterTypeAbsent confirms that the old env variable no longer
// controls behaviour: with OTEL_EXPORTER_TYPE set and OTEL_EXPORTER_OTLP_PROTOCOL absent,
// the new default (grpc) is applied and no error is returned.
func TestLoadSettings_OTELExporterTypeAbsent(t *testing.T) {
	id := makeIdentity(t, "test-svc", "1.0.0", "development")
	t.Setenv("OTEL_EXPORTER_TYPE", "otlp-http")   // old var — must be ignored
	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "")   // new var absent → default grpc
	t.Setenv("LOG_LEVEL", "info")

	s, err := bootstrapobs.LoadSettings(id)

	require.NoError(t, err)
	assert.Equal(t, "grpc", s.OTLPProtocol, "old OTEL_EXPORTER_TYPE must be ignored")
}

// TestNew_EmptyEndpoint_ReturnsNoop verifies that when Endpoint is empty the function
// returns a noop provider without error. This exercises the no-op branch of New().
func TestNew_EmptyEndpoint_ReturnsNoop(t *testing.T) {
	id := makeIdentity(t, "smoke-svc", "1.0.0", "development")
	s := bootstrapobs.Settings{
		Identity:             id,
		OTLPProtocol:         "grpc",
		Endpoint:             "", // empty → noop
		TracesSampler:        "parentbased_always_on",
		LogLevel:             "info",
		SlowQueryThresholdMS: 1000,
	}

	obs, err := bootstrapobs.New(t.Context(), s)

	require.NoError(t, err)
	assert.NotNil(t, obs)
	assert.NotNil(t, obs.Logger())
	assert.NotNil(t, obs.Tracer())
	assert.NotNil(t, obs.Metrics())
}

// TestNew_HTTPProtocol_EmptyEndpoint verifies the http/protobuf branch is reached when
// the endpoint is empty (returns noop; exercises protocol field handling).
func TestNew_HTTPProtocol_EmptyEndpoint(t *testing.T) {
	id := makeIdentity(t, "smoke-svc", "1.0.0", "development")
	s := bootstrapobs.Settings{
		Identity:             id,
		OTLPProtocol:         "http/protobuf",
		Endpoint:             "",
		LogLevel:             "debug",
		SlowQueryThresholdMS: 500,
	}

	obs, err := bootstrapobs.New(t.Context(), s)

	require.NoError(t, err)
	assert.NotNil(t, obs)
}

// TestNew_GRPC_LazyDial verifies that New() successfully initialises a provider with a
// gRPC endpoint even when the collector is unreachable (gRPC dials lazily by design).
// The Shutdown call is intentionally fast-failed via a very short timeout.
func TestNew_GRPC_LazyDial(t *testing.T) {
	id := makeIdentity(t, "smoke-svc", "1.0.0", "development")
	s := bootstrapobs.Settings{
		Identity:             id,
		OTLPProtocol:         "grpc",
		Endpoint:             "localhost:4317",
		TracesSampler:        "parentbased_always_on",
		LogLevel:             "info",
		SlowQueryThresholdMS: 1000,
	}

	obs, err := bootstrapobs.New(t.Context(), s)
	if err != nil {
		// If the provider fails to start (e.g., in an environment that eagerly dials),
		// verify the error wraps ErrExporterUnreachable and skip further assertions.
		require.ErrorIs(t, err, bootstrapobs.ErrExporterUnreachable)
		return
	}

	assert.NotNil(t, obs)

	shutCtx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	_ = obs.Shutdown(shutCtx) // best-effort; may time out if collector is absent
}

// TestNew_HTTP_LazyDial verifies that New() handles the http/protobuf protocol path.
func TestNew_HTTP_LazyDial(t *testing.T) {
	id := makeIdentity(t, "smoke-svc", "1.0.0", "development")
	s := bootstrapobs.Settings{
		Identity:             id,
		OTLPProtocol:         "http/protobuf",
		Endpoint:             "localhost:4318",
		TracesSampler:        "parentbased_always_on",
		LogLevel:             "debug",
		SlowQueryThresholdMS: 500,
	}

	obs, err := bootstrapobs.New(t.Context(), s)
	if err != nil {
		require.ErrorIs(t, err, bootstrapobs.ErrExporterUnreachable)
		return
	}

	assert.NotNil(t, obs)

	shutCtx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	_ = obs.Shutdown(shutCtx)
}
