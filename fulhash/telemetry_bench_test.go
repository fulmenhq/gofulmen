package fulhash

import (
	"testing"
	"time"

	"github.com/fulmenhq/gofulmen/telemetry"
)

func BenchmarkHash_WithTelemetry(b *testing.B) {
	collector := &nopEmitter{}
	telSys, _ := telemetry.NewSystem(&telemetry.Config{
		Enabled: true,
		Emitter: collector,
	})
	SetTelemetrySystem(telSys)

	data := []byte("test data for benchmarking")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Hash(data)
	}
}

func BenchmarkHash_WithoutTelemetry(b *testing.B) {
	SetTelemetrySystem(nil)

	data := []byte("test data for benchmarking")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Hash(data)
	}
}

type nopEmitter struct{}

func (n *nopEmitter) Counter(name string, value float64, tags map[string]string) error {
	return nil
}

func (n *nopEmitter) Histogram(name string, duration time.Duration, tags map[string]string) error {
	return nil
}

func (n *nopEmitter) HistogramSummary(name string, summary telemetry.HistogramSummary, tags map[string]string) error {
	return nil
}

func (n *nopEmitter) Gauge(name string, value float64, tags map[string]string) error {
	return nil
}
