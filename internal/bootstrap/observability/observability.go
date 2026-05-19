package observability

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/observability/noop"
	"github.com/JailtonJunior94/devkit-go/pkg/observability/otel"

	bootstrapconfig "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/config"
	"github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/observability/redactor"
)

var (
	// ErrInvalidOTLPProtocol is returned when OTEL_EXPORTER_OTLP_PROTOCOL has an unrecognised value.
	ErrInvalidOTLPProtocol = errors.New("OTEL_EXPORTER_OTLP_PROTOCOL must be one of: grpc, http/protobuf")
	// ErrInvalidLogLevel is returned when LOG_LEVEL has an unrecognised value.
	ErrInvalidLogLevel = errors.New("LOG_LEVEL must be one of: debug, info, warn, error")
	// ErrExporterUnreachable wraps errors from the OTLP provider initialisation.
	ErrExporterUnreachable = errors.New("OTLP exporter unreachable")
)

const (
	defaultEndpointGRPC = "localhost:4317"
	defaultEndpointHTTP = "localhost:4318"
	defaultSlowQueryMS  = int64(1000)
	defaultSampler      = "parentbased_always_on"
)

// K8sAttrs holds Kubernetes downward-API attributes for the OTel resource.
// All fields fall back to "unknown" when the corresponding env var is absent.
type K8sAttrs struct {
	PodUID        string // maps to service.instance.id (POD_UID)
	PodName       string // maps to k8s.pod.name       (POD_NAME)
	PodNamespace  string // maps to k8s.namespace.name  (POD_NAMESPACE)
	ContainerName string // maps to k8s.container.name  (CONTAINER_NAME)
}

// Settings holds validated OTel configuration derived from environment variables.
type Settings struct {
	Identity             bootstrapconfig.ServiceIdentity
	OTLPProtocol         string // "grpc" | "http/protobuf"
	Endpoint             string
	TracesSampler        string
	TracesSamplerArg     string
	LogLevel             observability.LogLevel // "debug" | "info" | "warn" | "error"
	SlowQueryThresholdMS int64
	K8s                  K8sAttrs
}

// loadK8sAttrs reads POD_UID, POD_NAME, POD_NAMESPACE and CONTAINER_NAME from env.
// Each missing value falls back to "unknown" and emits a WARN via the default slog handler.
func loadK8sAttrs() K8sAttrs {
	read := func(key string) string {
		v := os.Getenv(key)
		if v == "" {
			slog.Warn("K8s env var absent; resource attribute falls back to unknown", "env", key)
			return "unknown"
		}
		return v
	}
	return K8sAttrs{
		PodUID:        read("POD_UID"),
		PodName:       read("POD_NAME"),
		PodNamespace:  read("POD_NAMESPACE"),
		ContainerName: read("CONTAINER_NAME"),
	}
}

// LoadSettings reads OTel environment variables, applies defaults and validates values.
// See ADR-005 for the breaking removal of the legacy exporter-type variable.
// Reads: OTEL_EXPORTER_OTLP_PROTOCOL, OTEL_TRACES_SAMPLER, OTEL_TRACES_SAMPLER_ARG,
// LOG_LEVEL, SLOW_QUERY_THRESHOLD_MS, POD_UID, POD_NAME, POD_NAMESPACE, CONTAINER_NAME.
func LoadSettings(id bootstrapconfig.ServiceIdentity) (Settings, error) {
	protocol := os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL")
	if protocol == "" {
		protocol = "grpc"
	}
	switch protocol {
	case "grpc", "http/protobuf":
	default:
		return Settings{}, fmt.Errorf("bootstrap observability: %w", ErrInvalidOTLPProtocol)
	}

	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		switch protocol {
		case "grpc":
			endpoint = defaultEndpointGRPC
		case "http/protobuf":
			endpoint = defaultEndpointHTTP
		}
	}

	sampler := os.Getenv("OTEL_TRACES_SAMPLER")
	if sampler == "" {
		sampler = defaultSampler
	}
	samplerArg := os.Getenv("OTEL_TRACES_SAMPLER_ARG")

	rawLevel := os.Getenv("LOG_LEVEL")
	if rawLevel == "" {
		rawLevel = string(observability.LogLevelInfo)
	}
	logLevel := observability.LogLevel(rawLevel)
	switch logLevel {
	case observability.LogLevelDebug, observability.LogLevelInfo, observability.LogLevelWarn, observability.LogLevelError:
	default:
		return Settings{}, fmt.Errorf("bootstrap observability: %w", ErrInvalidLogLevel)
	}

	slowQueryMS := defaultSlowQueryMS
	if raw := os.Getenv("SLOW_QUERY_THRESHOLD_MS"); raw != "" {
		if parsed, err := strconv.ParseInt(raw, 10, 64); err == nil && parsed > 0 {
			slowQueryMS = parsed
		}
	}

	return Settings{
		Identity:             id,
		OTLPProtocol:         protocol,
		Endpoint:             endpoint,
		TracesSampler:        sampler,
		TracesSamplerArg:     samplerArg,
		LogLevel:             logLevel,
		SlowQueryThresholdMS: slowQueryMS,
		K8s:                  loadK8sAttrs(),
	}, nil
}

// New builds an observability.Observability provider from the given settings.
//
// Protocol mapping: "grpc" → ProtocolGRPC (port 4317); "http/protobuf" → ProtocolHTTP (port 4318).
// The returned provider is wrapped so logs and spans are redacted before they reach
// the underlying devkit implementation.
func New(ctx context.Context, s Settings) (observability.Observability, error) {
	if s.Endpoint == "" {
		return noop.NewProvider(), nil
	}

	protocol := otel.ProtocolGRPC
	if s.OTLPProtocol == "http/protobuf" {
		protocol = otel.ProtocolHTTP
	}

	cfg := &otel.Config{
		ServiceName:          s.Identity.Name(),
		ServiceVersion:       s.Identity.Version(),
		Environment:          s.Identity.Environment(),
		OTLPEndpoint:         s.Endpoint,
		OTLPProtocol:         protocol,
		Insecure:             true,
		TraceSampleRate:      1.0,
		LogLevel:             s.LogLevel,
		LogFormat:            observability.LogFormatJSON,
		MetricExportInterval: 60,
		ResourceAttributes: map[string]string{
			"service.instance.id": s.K8s.PodUID,
			"k8s.pod.name":        s.K8s.PodName,
			"k8s.namespace.name":  s.K8s.PodNamespace,
			"k8s.container.name":  s.K8s.ContainerName,
		},
	}

	provider, err := otel.NewProvider(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("bootstrap observability: %w: %w", ErrExporterUnreachable, err)
	}

	return wrapWithRedaction(provider, redactor.DefaultDenylist), nil
}
