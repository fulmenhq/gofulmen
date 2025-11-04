package appidentity

import "context"

// contextKey is a private type for context keys to avoid collisions.
type contextKey int

const (
	// identityContextKey is the context key for identity injection.
	identityContextKey contextKey = iota
)

// WithIdentity returns a new context with the given identity attached.
//
// This function enables context-based identity injection, primarily for testing.
// When an identity is attached to a context, Get() and GetWithOptions() will
// return it instead of loading from disk or cache.
//
// This is the highest priority in the discovery precedence chain, overriding
// all other sources (cache, env var, file discovery).
//
// Example (testing):
//
//	func TestMyFunction(t *testing.T) {
//	    testIdentity := &appidentity.Identity{
//	        BinaryName: "testapp",
//	        Vendor:     "testvendor",
//	        EnvPrefix:  "TESTAPP_",
//	        ConfigName: "testapp",
//	        Description: "Test application for unit tests",
//	    }
//	    ctx := appidentity.WithIdentity(context.Background(), testIdentity)
//
//	    // Code under test will use testIdentity
//	    identity, _ := appidentity.Get(ctx)
//	    // identity == testIdentity
//	}
func WithIdentity(ctx context.Context, identity *Identity) context.Context {
	return context.WithValue(ctx, identityContextKey, identity)
}

// fromContext retrieves an identity from the context if one was injected.
//
// Returns nil if no identity is attached to the context.
func fromContext(ctx context.Context) *Identity {
	if ctx == nil {
		return nil
	}

	identity, ok := ctx.Value(identityContextKey).(*Identity)
	if !ok {
		return nil
	}

	return identity
}
