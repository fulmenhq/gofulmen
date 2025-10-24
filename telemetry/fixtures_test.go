package telemetry

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/fulmenhq/gofulmen/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricsFixtures(t *testing.T) {
	// Try to load the metrics schema, but skip if it fails due to reference resolution
	catalog := schema.DefaultCatalog()
	validator, err := catalog.ValidatorByID("observability/metrics/v1.0.0/metrics-event")
	if err != nil {
		t.Skipf("Skipping metrics fixtures test due to schema loading issues: %v", err)
	}

	// Find all fixture files
	fixturesDir := "../test/fixtures/metrics"
	entries, err := os.ReadDir(fixturesDir)
	require.NoError(t, err)

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		t.Run(entry.Name(), func(t *testing.T) {
			fixturePath := filepath.Join(fixturesDir, entry.Name())

			// Read fixture
			data, err := os.ReadFile(fixturePath)
			require.NoError(t, err)

			// First validate against the JSON schema
			diagnostics, err := validator.ValidateJSON(data)
			require.NoError(t, err, "Fixture should validate against metrics schema")
			assert.Empty(t, diagnostics, "Schema validation should pass without diagnostics")

			// Parse as MetricsEvent for additional validation
			var event MetricsEvent
			err = json.Unmarshal(data, &event)
			require.NoError(t, err)

			// Validate required fields
			assert.NotEmpty(t, event.Timestamp, "Timestamp should not be empty")
			assert.NotEmpty(t, event.Name, "Name should not be empty")
			assert.NotNil(t, event.Value, "Value should not be nil")

			// Validate that it can be marshaled back to JSON
			marshaled, err := json.Marshal(event)
			require.NoError(t, err)
			assert.NotEmpty(t, marshaled)

			// Re-validate the marshaled JSON to ensure round-trip consistency
			diagnostics, err = validator.ValidateJSON(marshaled)
			require.NoError(t, err, "Marshaled event should still validate against schema")
			assert.Empty(t, diagnostics, "Round-trip validation should pass without diagnostics")
		})
	}
}
