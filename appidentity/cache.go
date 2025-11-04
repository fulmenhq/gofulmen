package appidentity

import (
	"context"
	"sync"
)

// Package-level cache for process-wide identity singleton
var (
	cachedIdentity *Identity
	cacheErr       error
	cacheOnce      sync.Once
	cacheMu        sync.RWMutex
)

// Get loads the application identity using automatic discovery and caching.
//
// Identity is loaded once per process and cached. Subsequent calls return
// the cached instance. Discovery follows this precedence:
//
//  1. Context injection (via WithIdentity) - highest priority
//  2. ExplicitPath in options (via GetWithOptions)
//  3. Environment variable (FULMEN_APP_IDENTITY_PATH)
//  4. Nearest ancestor search from current directory
//
// This function is thread-safe and uses sync.Once to ensure the identity
// is loaded exactly once, even under concurrent access.
//
// Example:
//
//	identity, err := appidentity.Get(ctx)
//	if err != nil {
//	    return fmt.Errorf("failed to load identity: %w", err)
//	}
//	fmt.Println("Binary:", identity.Binary())
func Get(ctx context.Context) (*Identity, error) {
	return GetWithOptions(ctx, Options{})
}

// GetWithOptions loads the application identity with explicit options.
//
// This function provides fine-grained control over identity loading:
//   - ExplicitPath: Load from a specific file path
//   - RepoRoot: Start ancestor search from a specific directory
//   - NoCache: Bypass the process-level cache (useful for testing)
//
// Discovery precedence (highest to lowest):
//  1. Context injection (via WithIdentity)
//  2. opts.ExplicitPath
//  3. Environment variable (FULMEN_APP_IDENTITY_PATH)
//  4. Nearest ancestor search from opts.RepoRoot (default: cwd)
//
// Example:
//
//	identity, err := appidentity.GetWithOptions(ctx, appidentity.Options{
//	    ExplicitPath: "/custom/path/app.yaml",
//	    NoCache:      true,
//	})
func GetWithOptions(ctx context.Context, opts Options) (*Identity, error) {
	// Priority 1: Check for context injection (override)
	if identity := fromContext(ctx); identity != nil {
		return identity, nil
	}

	// If NoCache is set, bypass the cache (useful for testing)
	if opts.NoCache {
		return discoverIdentity(ctx, opts)
	}

	// Use process-level cache with sync.Once
	cacheOnce.Do(func() {
		cachedIdentity, cacheErr = discoverIdentity(ctx, opts)
	})

	return cachedIdentity, cacheErr
}

// Must loads the application identity and panics on error.
//
// This is a convenience wrapper around Get for use in main() or init()
// functions where identity is required for the application to function.
//
// Example:
//
//	func main() {
//	    identity := appidentity.Must(context.Background())
//	    fmt.Println("Starting", identity.Binary())
//	    // ... rest of application
//	}
func Must(ctx context.Context) *Identity {
	identity, err := Get(ctx)
	if err != nil {
		panic("failed to load application identity: " + err.Error())
	}
	return identity
}

// Reset clears the process-level cache.
//
// This function is intended for testing only. It allows tests to reload
// identity configuration between test cases. In production code, identity
// should be loaded once and remain constant for the process lifetime.
//
// Example:
//
//	func TestMultipleIdentities(t *testing.T) {
//	    defer appidentity.Reset() // Clean up after test
//	    // ... test code
//	}
//
// IMPORTANT: Reset is NOT safe to call concurrently with Get/GetWithOptions.
// It takes cacheMu lock to protect cache state, but Get/GetWithOptions use
// sync.Once without taking the lock (for performance). Only call Reset in
// single-threaded test contexts, or ensure all Get/GetWithOptions calls have
// completed before calling Reset.
func Reset() {
	cacheMu.Lock()
	defer cacheMu.Unlock()

	cachedIdentity = nil
	cacheErr = nil
	cacheOnce = sync.Once{}
}
