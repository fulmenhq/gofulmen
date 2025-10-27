// Package main demonstrates Phase 5 advanced telemetry features including gauge metrics and custom exporters
//
// SECURITY NOTE: This demo uses math/rand for generating fake metric values (CPU percentages,
// memory usage, etc.). For production code that requires cryptographically secure randomness,
// replace "math/rand" with "crypto/rand". The use of math/rand here is intentional and safe
// for demonstration purposes since we're only generating synthetic monitoring data.
package main

import (
	"fmt"
	"log"
	"math/rand" // #nosec G404 - Intentional use for demo data generation, not security-sensitive
	"strings"
	"time"

	"github.com/fulmenhq/gofulmen/telemetry"
	"github.com/fulmenhq/gofulmen/telemetry/exporters"
)

// SystemMonitor demonstrates real-world gauge metric usage
func SystemMonitor() {
	// Create telemetry system with default config
	sys, err := telemetry.NewSystem(telemetry.DefaultConfig())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== System Monitor Demo ===")
	fmt.Println("Monitoring system metrics with gauge metrics...")

	// Simulate monitoring CPU usage every second
	for i := 0; i < 5; i++ {
		cpuUsage := 20.0 + rand.Float64()*60.0 // #nosec G404 - Demo data only, not security-sensitive

		err = sys.Gauge("system_cpu_usage_percent", cpuUsage, map[string]string{
			"host": "web-server-01",
			"cpu":  "cpu0",
		})
		if err != nil {
			log.Printf("Error emitting CPU gauge: %v", err)
		}

		// Memory usage in MB
		memoryUsage := 2048.0 + rand.Float64()*4096.0 // #nosec G404 - Demo data only, not security-sensitive
		err = sys.Gauge("system_memory_usage_mb", memoryUsage, map[string]string{
			"host": "web-server-01",
			"type": "used",
		})
		if err != nil {
			log.Printf("Error emitting memory gauge: %v", err)
		}

		// Temperature in Celsius
		temperature := 20.0 + rand.Float64()*15.0 // #nosec G404 - Demo data only, not security-sensitive
		err = sys.Gauge("environment_temperature_celsius", temperature, map[string]string{
			"sensor": "indoor",
			"room":   "server-room",
		})
		if err != nil {
			log.Printf("Error emitting temperature gauge: %v", err)
		}

		time.Sleep(1 * time.Second)
	}

	fmt.Println("System monitoring complete!")
}

// PrometheusExporterDemo demonstrates custom exporter usage
func PrometheusExporterDemo() {
	fmt.Println("\n=== Prometheus Exporter Demo ===")
	fmt.Println("Starting Prometheus metrics exporter...")

	// Create a Prometheus exporter
	exporter := exporters.NewPrometheusExporter("demo", ":9090")

	// Start the HTTP server
	err := exporter.Start()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if stopErr := exporter.Stop(); stopErr != nil {
			log.Printf("Error stopping exporter: %v", stopErr)
		}
	}()

	fmt.Println("Prometheus exporter started on http://localhost:9090/metrics")
	fmt.Println("You can view metrics at: http://localhost:9090/metrics")
	fmt.Println("Main page: http://localhost:9090/")
	fmt.Println("\nSimulating application metrics for 10 seconds...")

	// Simulate application metrics
	startTime := time.Now()
	for time.Since(startTime) < 10*time.Second {
		// HTTP requests counter
		requests := float64(100 + rand.Intn(50)) // #nosec G404 - Demo data only, not security-sensitive
		if err := exporter.Counter("http_requests_total", requests, map[string]string{
			"status": "200",
			"method": "GET",
		}); err != nil {
			log.Printf("Error emitting counter: %v", err)
		}

		// Request duration histogram
		duration := time.Duration(50+rand.Intn(150)) * time.Millisecond // #nosec G404 - Demo data only, not security-sensitive
		if err := exporter.Histogram("http_request_duration_ms", duration, map[string]string{
			"endpoint": "/api/users",
		}); err != nil {
			log.Printf("Error emitting histogram: %v", err)
		}

		// System gauges
		if err := exporter.Gauge("cpu_usage_percent", 30.0+rand.Float64()*40.0, map[string]string{ // #nosec G404 - Demo data only, not security-sensitive
			"host": "demo-server",
		}); err != nil {
			log.Printf("Error emitting CPU gauge: %v", err)
		}

		if err := exporter.Gauge("memory_usage_percent", 50.0+rand.Float64()*30.0, map[string]string{ // #nosec G404 - Demo data only, not security-sensitive
			"host": "demo-server",
		}); err != nil {
			log.Printf("Error emitting memory gauge: %v", err)
		}

		// Active connections gauge
		connections := float64(50 + rand.Intn(100)) // #nosec G404 - Demo data only, not security-sensitive
		if err := exporter.Gauge("active_connections", connections, map[string]string{
			"protocol": "http",
		}); err != nil {
			log.Printf("Error emitting connections gauge: %v", err)
		}

		time.Sleep(500 * time.Millisecond)
	}

	fmt.Println("\nPrometheus exporter demo complete!")
	fmt.Println("Metrics were available at http://localhost:9090/metrics")
}

// BatchMetricsDemo demonstrates batching with mixed metric types
func BatchMetricsDemo() {
	fmt.Println("\n=== Batch Metrics Demo ===")
	fmt.Println("Demonstrating batching with counters, gauges, and histograms...")

	// Create telemetry system with batching enabled
	config := &telemetry.Config{
		Enabled:       true,
		BatchSize:     5,               // Flush after 5 metrics
		BatchInterval: 1 * time.Second, // Or flush every second
	}

	sys, err := telemetry.NewSystem(config)
	if err != nil {
		log.Fatal(err)
	}

	// Emit mixed metric types
	for i := 0; i < 12; i++ {
		// Counter - total requests
		err = sys.Counter("requests_total", 1, map[string]string{
			"status": "200",
		})
		if err != nil {
			log.Printf("Error with counter: %v", err)
		}

		// Gauge - current memory usage
		memory := 1024.0 + float64(i)*100.0
		err = sys.Gauge("memory_usage_mb", memory, map[string]string{
			"type": "heap",
		})
		if err != nil {
			log.Printf("Error with gauge: %v", err)
		}

		// Histogram - request duration
		duration := time.Duration(50+i*10) * time.Millisecond
		err = sys.Histogram("request_duration_ms", duration, map[string]string{
			"endpoint": "/api/data",
		})
		if err != nil {
			log.Printf("Error with histogram: %v", err)
		}

		time.Sleep(200 * time.Millisecond)
	}

	// Manual flush to ensure all metrics are emitted
	err = sys.Flush()
	if err != nil {
		log.Printf("Error flushing metrics: %v", err)
	}

	fmt.Println("Batch metrics demo complete!")
}

// CrossLanguagePatternsDemo demonstrates patterns for TS/Py teams
func CrossLanguagePatternsDemo() {
	fmt.Println("\n=== Cross-Language Patterns Demo ===")
	fmt.Println("Demonstrating patterns for TypeScript and Python teams...")

	// Pattern 1: Consistent metric naming
	fmt.Println("\n1. Consistent Metric Naming:")
	fmt.Println("   - Use snake_case: system_cpu_usage_percent")
	fmt.Println("   - End with unit: _ms, _percent, _bytes")
	fmt.Println("   - Include context: host, service, region tags")

	// Pattern 2: Error handling with structured context
	fmt.Println("\n2. Structured Error Context:")
	fmt.Println("   - Always include component, operation, error_type")
	fmt.Println("   - Use correlation IDs for tracing")
	fmt.Println("   - Log errors but don't fail telemetry operations")

	// Pattern 3: Batch-friendly configuration
	fmt.Println("\n3. Batch-Friendly Configuration:")
	fmt.Println("   - Default to no batching (immediate emission)")
	fmt.Println("   - Allow opt-in batching for high-frequency scenarios")
	fmt.Println("   - Provide flush methods for manual control")

	// Pattern 4: Schema validation
	fmt.Println("\n4. Schema Validation:")
	fmt.Println("   - Validate against canonical taxonomy")
	fmt.Println("   - Graceful degradation when validation fails")
	fmt.Println("   - Use JSON schema for cross-language consistency")

	// Pattern 5: Performance considerations
	fmt.Println("\n5. Performance Patterns:")
	fmt.Println("   - <5% overhead target")
	fmt.Println("   - Lazy initialization")
	fmt.Println("   - Efficient data structures")
	fmt.Println("   - Proper error handling without blocking")

	fmt.Println("\nCross-language patterns demo complete!")
}

// RunPhase5Demo runs the Phase 5 telemetry demo
func RunPhase5Demo() {
	fmt.Println("🚀 Fulmen Telemetry - Phase 5 Advanced Features Demo")
	fmt.Println(strings.Repeat("=", 51))

	// Run all demos
	SystemMonitor()

	// Uncomment to run individual demos:
	// PrometheusExporterDemo()
	// BatchMetricsDemo()
	// CrossLanguagePatternsDemo()

	fmt.Println("\n✅ All Phase 5 demos completed successfully!")
	fmt.Println("\nNext steps for TS/Py teams:")
	fmt.Println("1. Implement gauge metrics with same interface")
	fmt.Println("2. Create language-specific exporters (Prometheus, Datadog)")
	fmt.Println("3. Add metric aggregation and sampling")
	fmt.Println("4. Implement schema validation against canonical taxonomy")
	fmt.Println("5. Ensure <5% performance overhead")
}
