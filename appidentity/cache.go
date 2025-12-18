package appidentity

import (
	"context"
	"sync"
)

// Package-level cache for process-wide identity singleton.
//
// Cache behavior:
//   - The first successful discovery is cached for the process lifetime.
//   - Errors are not cached, so callers can retry after fixing environment
//     conditions (changing CWD, setting env vars, creating files, etc.).
var (
	cacheMu      sync.Mutex
	cacheCond    = sync.NewCond(&cacheMu)
	cacheLoading bool

	cachedIdentity *Identity
)

// Get loads the application identity using automatic discovery and caching.
//
// Identity is loaded once per process and cached on the first successful
// discovery. Subsequent calls return the cached instance. Discovery follows
// this precedence:
//
//  1. Context injection (via WithIdentity) - highest priority
//  2. ExplicitPath in options (via GetWithOptions)
//  3. Environment variable (FULMEN_APP_IDENTITY_PATH)
//  4. Nearest ancestor search from current directory
//  5. Fallback: Nearest ancestor search from executable directory
//  6. Embedded identity (registered via RegisterEmbeddedIdentityYAML)
//
// This function is thread-safe and ensures only one discovery attempt runs at
// a time under concurrent access.
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
//  5. Fallback: Nearest ancestor search from executable directory
//  6. Embedded identity (registered via RegisterEmbeddedIdentityYAML)
func GetWithOptions(ctx context.Context, opts Options) (*Identity, error) {
	// Priority 1: Check for context injection (override).
	if identity := fromContext(ctx); identity != nil {
		return identity, nil
	}

	// If NoCache is set, bypass the cache (useful for testing).
	if opts.NoCache {
		return discoverIdentity(ctx, opts)
	}

	cacheMu.Lock()
	for {
		if cachedIdentity != nil {
			identity := cachedIdentity
			cacheMu.Unlock()
			return identity, nil
		}

		if !cacheLoading {
			cacheLoading = true
			break
		}

		cacheCond.Wait()
	}
	cacheMu.Unlock()

	identity, err := discoverIdentity(ctx, opts)

	cacheMu.Lock()
	cacheLoading = false
	if err == nil && identity != nil {
		cachedIdentity = identity
	}
	cacheCond.Broadcast()
	cacheMu.Unlock()

	return identity, err
}

// Must loads the application identity and panics on error.
//
// This is a convenience wrapper around Get for use in main() or init()
// functions where identity is required for the application to function.
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
// identity configuration between test cases.
//
// IMPORTANT: Reset is NOT safe to call concurrently with Get/GetWithOptions.
func Reset() {
	cacheMu.Lock()
	cachedIdentity = nil
	cacheLoading = false
	cacheCond.Broadcast()
	cacheMu.Unlock()

	resetEmbeddedIdentity()
}
