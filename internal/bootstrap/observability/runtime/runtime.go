package runtime

import (
	"context"
	gort "runtime"
	"runtime/metrics"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
)

// Register registers Go runtime metrics as OTel observable gauges.
// Instruments: go_goroutines, go_gc_pause_total_ns, go_gc_cycles_total,
// process_resident_memory_bytes, process_cpu_seconds_total.
func Register(obs observability.Observability) error {
	return register(obs.Metrics())
}

// register is the testable core: accepts observability.Metrics so tests can inject a fake.
func register(m observability.Metrics) error {
	if err := m.Gauge("go_goroutines",
		"Number of goroutines currently running",
		"1",
		goroutinesValue,
	); err != nil {
		return err
	}

	if err := m.Gauge("go_gc_pause_total_ns",
		"Cumulative GC stop-the-world pause duration in nanoseconds",
		"ns",
		gcPauseTotalNs,
	); err != nil {
		return err
	}

	if err := m.Gauge("go_gc_cycles_total",
		"Total number of completed GC cycles",
		"1",
		gcCyclesTotal,
	); err != nil {
		return err
	}

	if err := m.Gauge("process_resident_memory_bytes",
		"Total memory obtained from the OS in bytes",
		"bytes",
		processResidentMemoryBytes,
	); err != nil {
		return err
	}

	if err := m.Gauge("process_cpu_seconds_total",
		"Total CPU time consumed by the process in seconds",
		"s",
		processCPUSecondsTotal,
	); err != nil {
		return err
	}

	return nil
}

func goroutinesValue(_ context.Context) float64 {
	return float64(gort.NumGoroutine())
}

func gcPauseTotalNs(_ context.Context) float64 {
	var ms gort.MemStats
	gort.ReadMemStats(&ms)
	return float64(ms.PauseTotalNs)
}

func gcCyclesTotal(_ context.Context) float64 {
	samples := []metrics.Sample{
		{Name: "/gc/cycles/total:gc-cycles"},
	}
	metrics.Read(samples)
	if samples[0].Value.Kind() == metrics.KindUint64 {
		return float64(samples[0].Value.Uint64())
	}
	return 0
}

func processResidentMemoryBytes(_ context.Context) float64 {
	var ms gort.MemStats
	gort.ReadMemStats(&ms)
	return float64(ms.Sys)
}

func processCPUSecondsTotal(_ context.Context) float64 {
	samples := []metrics.Sample{
		{Name: "/cpu/classes/total:cpu-seconds"},
	}
	metrics.Read(samples)
	if samples[0].Value.Kind() == metrics.KindFloat64 {
		return samples[0].Value.Float64()
	}
	return 0
}
