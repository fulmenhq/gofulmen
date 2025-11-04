package appidentity

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// TestGetCaching verifies that Get caches identity after first load.
func TestGetCaching(t *testing.T) {
	// Create temporary identity file
	tmpDir := t.TempDir()
	identityDir := filepath.Join(tmpDir, ".fulmen")
	if err := os.MkdirAll(identityDir, 0755); err != nil {
		t.Fatalf("failed to create .fulmen dir: %v", err)
	}

	identityPath := filepath.Join(identityDir, "app.yaml")
	content := []byte("app:\n  binary_name: cachetest\n  vendor: cachetest\n  env_prefix: CACHETEST_\n  config_name: cachetest\n  description: Cache test application\n")
	if err := os.WriteFile(identityPath, content, 0644); err != nil {
		t.Fatalf("failed to write identity file: %v", err)
	}

	// Save current directory and change to tmpDir
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	defer func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Errorf("failed to restore directory: %v", err)
		}
	}()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	// Reset cache before test
	Reset()
	defer Reset()

	ctx := context.Background()

	// First call should load from disk
	identity1, err := Get(ctx)
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	if identity1.BinaryName != "cachetest" {
		t.Errorf("BinaryName = %q, want %q", identity1.BinaryName, "cachetest")
	}

	// Second call should return cached instance (same pointer)
	identity2, err := Get(ctx)
	if err != nil {
		t.Fatalf("Get() failed on second call: %v", err)
	}

	if identity1 != identity2 {
		t.Error("Get() should return same cached instance")
	}
}

// TestGetConcurrent verifies thread-safe concurrent access.
func TestGetConcurrent(t *testing.T) {
	// Create temporary identity file
	tmpDir := t.TempDir()
	identityDir := filepath.Join(tmpDir, ".fulmen")
	if err := os.MkdirAll(identityDir, 0755); err != nil {
		t.Fatalf("failed to create .fulmen dir: %v", err)
	}

	identityPath := filepath.Join(identityDir, "app.yaml")
	content := []byte("app:\n  binary_name: concurrent\n  vendor: concurrent\n  env_prefix: CONCURRENT_\n  config_name: concurrent\n  description: Concurrent test application\n")
	if err := os.WriteFile(identityPath, content, 0644); err != nil {
		t.Fatalf("failed to write identity file: %v", err)
	}

	// Save current directory and change to tmpDir
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	defer func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Errorf("failed to restore directory: %v", err)
		}
	}()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	// Reset cache before test
	Reset()
	defer Reset()

	ctx := context.Background()

	// Spawn multiple goroutines trying to Get simultaneously
	const goroutines = 50
	var wg sync.WaitGroup
	identities := make([]*Identity, goroutines)
	errors := make([]error, goroutines)

	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			identities[idx], errors[idx] = Get(ctx)
		}(i)
	}

	wg.Wait()

	// All should succeed
	for i, err := range errors {
		if err != nil {
			t.Errorf("goroutine %d failed: %v", i, err)
		}
	}

	// All should return the same instance
	first := identities[0]
	for i, identity := range identities[1:] {
		if identity != first {
			t.Errorf("goroutine %d got different instance", i+1)
		}
	}
}

// TestGetWithOptionsNoCache verifies NoCache option bypasses cache.
func TestGetWithOptionsNoCache(t *testing.T) {
	// Create temporary identity file
	tmpDir := t.TempDir()
	identityDir := filepath.Join(tmpDir, ".fulmen")
	if err := os.MkdirAll(identityDir, 0755); err != nil {
		t.Fatalf("failed to create .fulmen dir: %v", err)
	}

	identityPath := filepath.Join(identityDir, "app.yaml")
	content := []byte("app:\n  binary_name: nocache\n  vendor: nocache\n  env_prefix: NOCACHE_\n  config_name: nocache\n  description: NoCache test application\n")
	if err := os.WriteFile(identityPath, content, 0644); err != nil {
		t.Fatalf("failed to write identity file: %v", err)
	}

	// Save current directory and change to tmpDir
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	defer func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Errorf("failed to restore directory: %v", err)
		}
	}()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	// Reset cache before test
	Reset()
	defer Reset()

	ctx := context.Background()

	// First call with NoCache
	identity1, err := GetWithOptions(ctx, Options{NoCache: true})
	if err != nil {
		t.Fatalf("GetWithOptions() failed: %v", err)
	}

	// Second call with NoCache should return different instance
	identity2, err := GetWithOptions(ctx, Options{NoCache: true})
	if err != nil {
		t.Fatalf("GetWithOptions() failed on second call: %v", err)
	}

	if identity1 == identity2 {
		t.Error("GetWithOptions(NoCache: true) should not cache")
	}

	// Both should have same content
	if identity1.BinaryName != identity2.BinaryName {
		t.Error("Identities should have same content even if not cached")
	}
}

// TestMust verifies Must panics on error.
func TestMust(t *testing.T) {
	// Create a context with no identity file available
	tmpDir := t.TempDir()

	// Save current directory and change to tmpDir
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	defer func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Errorf("failed to restore directory: %v", err)
		}
	}()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	// Reset cache before test
	Reset()
	defer Reset()

	ctx := context.Background()

	// Should panic because no identity file exists
	defer func() {
		if r := recover(); r == nil {
			t.Error("Must() should panic when identity cannot be loaded")
		}
	}()

	Must(ctx)
}

// TestMustSuccess verifies Must returns identity when available.
func TestMustSuccess(t *testing.T) {
	// Create temporary identity file
	tmpDir := t.TempDir()
	identityDir := filepath.Join(tmpDir, ".fulmen")
	if err := os.MkdirAll(identityDir, 0755); err != nil {
		t.Fatalf("failed to create .fulmen dir: %v", err)
	}

	identityPath := filepath.Join(identityDir, "app.yaml")
	content := []byte("app:\n  binary_name: musttest\n  vendor: musttest\n  env_prefix: MUSTTEST_\n  config_name: musttest\n  description: Must test application\n")
	if err := os.WriteFile(identityPath, content, 0644); err != nil {
		t.Fatalf("failed to write identity file: %v", err)
	}

	// Save current directory and change to tmpDir
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	defer func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Errorf("failed to restore directory: %v", err)
		}
	}()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	// Reset cache before test
	Reset()
	defer Reset()

	ctx := context.Background()

	// Should not panic
	identity := Must(ctx)
	if identity == nil {
		t.Fatal("Must() returned nil")
	}

	if identity.BinaryName != "musttest" {
		t.Errorf("BinaryName = %q, want %q", identity.BinaryName, "musttest")
	}
}

// TestReset verifies Reset clears the cache.
func TestReset(t *testing.T) {
	// Create temporary identity file
	tmpDir := t.TempDir()
	identityDir := filepath.Join(tmpDir, ".fulmen")
	if err := os.MkdirAll(identityDir, 0755); err != nil {
		t.Fatalf("failed to create .fulmen dir: %v", err)
	}

	identityPath := filepath.Join(identityDir, "app.yaml")
	content := []byte("app:\n  binary_name: resettest\n  vendor: resettest\n  env_prefix: RESETTEST_\n  config_name: resettest\n  description: Reset test application\n")
	if err := os.WriteFile(identityPath, content, 0644); err != nil {
		t.Fatalf("failed to write identity file: %v", err)
	}

	// Save current directory and change to tmpDir
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	defer func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Errorf("failed to restore directory: %v", err)
		}
	}()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	// Reset cache before test
	Reset()
	defer Reset()

	ctx := context.Background()

	// Load identity
	identity1, err := Get(ctx)
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	// Reset cache
	Reset()

	// Load again should create new instance
	identity2, err := Get(ctx)
	if err != nil {
		t.Fatalf("Get() failed after Reset: %v", err)
	}

	if identity1 == identity2 {
		t.Error("Get() after Reset() should return new instance")
	}

	// But content should be same
	if identity1.BinaryName != identity2.BinaryName {
		t.Error("Identities should have same content")
	}
}
