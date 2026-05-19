package runtime_test

import (
	"context"
	"errors"
	"testing"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/observability/fake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ort "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/observability/runtime"
)

// gaugeCapture is a minimal observability.Metrics that captures gauge callbacks for inspection.
type gaugeCapture struct {
	callbacks   map[string]observability.GaugeCallback
	failOnGauge string // if non-empty, Gauge returns an error for this name
}

func newGaugeCapture() *gaugeCapture {
	return &gaugeCapture{callbacks: make(map[string]observability.GaugeCallback)}
}

func (g *gaugeCapture) Counter(_, _, _ string) observability.Counter {
	return noopCounter{}
}
func (g *gaugeCapture) Histogram(_, _, _ string) observability.Histogram {
	return noopHistogram{}
}
func (g *gaugeCapture) HistogramWithBuckets(_, _, _ string, _ []float64) observability.Histogram {
	return noopHistogram{}
}
func (g *gaugeCapture) UpDownCounter(_, _, _ string) observability.UpDownCounter {
	return noopUpDownCounter{}
}
func (g *gaugeCapture) Gauge(name, _, _ string, cb observability.GaugeCallback) error {
	if g.failOnGauge != "" && name == g.failOnGauge {
		return errors.New("gauge registration failed")
	}
	g.callbacks[name] = cb
	return nil
}

type noopCounter struct{}

func (noopCounter) Add(context.Context, int64, ...observability.Field) {}
func (noopCounter) Increment(context.Context, ...observability.Field)  {}

type noopHistogram struct{}

func (noopHistogram) Record(context.Context, float64, ...observability.Field) {}

type noopUpDownCounter struct{}

func (noopUpDownCounter) Add(context.Context, int64, ...observability.Field) {}

// capturingProvider wraps a gaugeCapture as observability.Observability.
type capturingProvider struct {
	*fake.Provider
	m *gaugeCapture
}

func newCapturingProvider() *capturingProvider {
	return &capturingProvider{
		Provider: fake.NewProvider(),
		m:        newGaugeCapture(),
	}
}

func (p *capturingProvider) Metrics() observability.Metrics { return p.m }

func TestRegisterSucceeds(t *testing.T) {
	cp := newCapturingProvider()
	err := ort.Register(cp)
	require.NoError(t, err)
}

func TestAllGaugesRegistered(t *testing.T) {
	cp := newCapturingProvider()
	require.NoError(t, ort.Register(cp))

	expected := []string{
		"go_goroutines",
		"go_gc_pause_total_ns",
		"go_gc_cycles_total",
		"process_resident_memory_bytes",
		"process_cpu_seconds_total",
	}
	for _, name := range expected {
		t.Run(name, func(t *testing.T) {
			_, ok := cp.m.callbacks[name]
			assert.True(t, ok, "gauge %q not registered", name)
		})
	}
}

func TestGoGoroutinesCallbackReturnsAtLeastOne(t *testing.T) {
	cp := newCapturingProvider()
	require.NoError(t, ort.Register(cp))

	cb, ok := cp.m.callbacks["go_goroutines"]
	require.True(t, ok)

	val := cb(context.Background())
	assert.GreaterOrEqual(t, val, float64(1), "at least one goroutine must be running")
}

func TestGCPauseTotalNsIsNonNegative(t *testing.T) {
	cp := newCapturingProvider()
	require.NoError(t, ort.Register(cp))

	cb := cp.m.callbacks["go_gc_pause_total_ns"]
	val := cb(context.Background())
	assert.GreaterOrEqual(t, val, float64(0))
}

func TestGCCyclesTotalIsNonNegative(t *testing.T) {
	cp := newCapturingProvider()
	require.NoError(t, ort.Register(cp))

	cb := cp.m.callbacks["go_gc_cycles_total"]
	val := cb(context.Background())
	assert.GreaterOrEqual(t, val, float64(0))
}

func TestProcessResidentMemoryIsPositive(t *testing.T) {
	cp := newCapturingProvider()
	require.NoError(t, ort.Register(cp))

	cb := cp.m.callbacks["process_resident_memory_bytes"]
	val := cb(context.Background())
	assert.Greater(t, val, float64(0), "process must have memory allocated")
}

func TestProcessCPUIsNonNegative(t *testing.T) {
	cp := newCapturingProvider()
	require.NoError(t, ort.Register(cp))

	cb := cp.m.callbacks["process_cpu_seconds_total"]
	val := cb(context.Background())
	assert.GreaterOrEqual(t, val, float64(0))
}

func TestRegisterPropagatesGaugeError(t *testing.T) {
	// Verify that if any Gauge registration fails, Register returns the error.
	gaugeNames := []string{
		"go_goroutines",
		"go_gc_pause_total_ns",
		"go_gc_cycles_total",
		"process_resident_memory_bytes",
		"process_cpu_seconds_total",
	}
	for _, name := range gaugeNames {
		t.Run(name, func(t *testing.T) {
			m := newGaugeCapture()
			m.failOnGauge = name
			cp := &capturingProvider{
				Provider: fake.NewProvider(),
				m:        m,
			}
			err := ort.Register(cp)
			assert.Error(t, err)
		})
	}
}
