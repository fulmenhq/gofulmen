package appidentity

import (
	"context"
	"testing"
)

// TestWithIdentity verifies context-based identity injection.
func TestWithIdentity(t *testing.T) {
	ctx := context.Background()

	testIdentity := &Identity{
		BinaryName:  "testapp",
		Vendor:      "testvendor",
		EnvPrefix:   "TESTAPP_",
		ConfigName:  "testapp",
		Description: "Test application for context injection",
	}

	// Inject identity into context
	ctx = WithIdentity(ctx, testIdentity)

	// Get should return injected identity
	identity, err := Get(ctx)
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	if identity != testIdentity {
		t.Error("Get() should return injected identity from context")
	}

	if identity.BinaryName != "testapp" {
		t.Errorf("BinaryName = %q, want %q", identity.BinaryName, "testapp")
	}
}

// TestWithIdentityOverridesCache verifies context injection overrides cache.
func TestWithIdentityOverridesCache(t *testing.T) {
	// Reset cache before test
	Reset()
	defer Reset()

	ctx := context.Background()

	cachedIdentity := &Identity{
		BinaryName:  "cached",
		Vendor:      "cached",
		EnvPrefix:   "CACHED_",
		ConfigName:  "cached",
		Description: "Cached identity for override test",
	}

	overrideIdentity := &Identity{
		BinaryName:  "override",
		Vendor:      "override",
		EnvPrefix:   "OVERRIDE_",
		ConfigName:  "override",
		Description: "Override identity for context test",
	}

	// Simulate cached identity by injecting and calling Get
	ctxCached := WithIdentity(ctx, cachedIdentity)
	_, _ = Get(ctxCached)

	// Now use different identity via context injection
	ctxOverride := WithIdentity(ctx, overrideIdentity)
	identity, err := Get(ctxOverride)
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	// Should return override identity, not cached
	if identity != overrideIdentity {
		t.Error("Context injection should override cache")
	}

	if identity.BinaryName != "override" {
		t.Errorf("BinaryName = %q, want %q (override)", identity.BinaryName, "override")
	}
}

// TestWithIdentityNoCache verifies context injection works with NoCache option.
func TestWithIdentityNoCache(t *testing.T) {
	ctx := context.Background()

	testIdentity := &Identity{
		BinaryName:  "nocachetest",
		Vendor:      "nocachetest",
		EnvPrefix:   "NOCACHETEST_",
		ConfigName:  "nocachetest",
		Description: "NoCache test with context injection",
	}

	ctx = WithIdentity(ctx, testIdentity)

	// GetWithOptions with NoCache should still respect context injection
	identity, err := GetWithOptions(ctx, Options{NoCache: true})
	if err != nil {
		t.Fatalf("GetWithOptions() failed: %v", err)
	}

	if identity != testIdentity {
		t.Error("Context injection should work with NoCache option")
	}
}

// TestFromContextNil verifies fromContext handles nil context.
func TestFromContextNil(t *testing.T) {
	identity := fromContext(context.TODO())
	if identity != nil {
		t.Error("fromContext(context.TODO()) should return nil")
	}
}

// TestFromContextEmpty verifies fromContext returns nil for empty context.
func TestFromContextEmpty(t *testing.T) {
	ctx := context.Background()
	identity := fromContext(ctx)
	if identity != nil {
		t.Error("fromContext() should return nil for context without identity")
	}
}

// TestWithIdentityNested verifies nested contexts work correctly.
func TestWithIdentityNested(t *testing.T) {
	ctx := context.Background()

	identity1 := &Identity{
		BinaryName:  "first",
		Vendor:      "first",
		EnvPrefix:   "FIRST_",
		ConfigName:  "first",
		Description: "First identity in nested test",
	}

	identity2 := &Identity{
		BinaryName:  "second",
		Vendor:      "second",
		EnvPrefix:   "SECOND_",
		ConfigName:  "second",
		Description: "Second identity in nested test",
	}

	// Create nested contexts
	ctx1 := WithIdentity(ctx, identity1)
	ctx2 := WithIdentity(ctx1, identity2)

	// ctx1 should have identity1
	got1, err := Get(ctx1)
	if err != nil {
		t.Fatalf("Get(ctx1) failed: %v", err)
	}
	if got1 != identity1 {
		t.Error("ctx1 should have identity1")
	}

	// ctx2 should have identity2 (overrides parent)
	got2, err := Get(ctx2)
	if err != nil {
		t.Fatalf("Get(ctx2) failed: %v", err)
	}
	if got2 != identity2 {
		t.Error("ctx2 should have identity2 (child overrides parent)")
	}
}

// TestWithIdentityIsolation verifies contexts are isolated.
func TestWithIdentityIsolation(t *testing.T) {
	ctx := context.Background()

	identity1 := &Identity{
		BinaryName:  "isolated1",
		Vendor:      "isolated1",
		EnvPrefix:   "ISOLATED1_",
		ConfigName:  "isolated1",
		Description: "First isolated identity",
	}

	identity2 := &Identity{
		BinaryName:  "isolated2",
		Vendor:      "isolated2",
		EnvPrefix:   "ISOLATED2_",
		ConfigName:  "isolated2",
		Description: "Second isolated identity",
	}

	// Create separate contexts
	ctx1 := WithIdentity(ctx, identity1)
	ctx2 := WithIdentity(ctx, identity2)

	// Each context should have its own identity
	got1, err := Get(ctx1)
	if err != nil {
		t.Fatalf("Get(ctx1) failed: %v", err)
	}
	if got1 != identity1 {
		t.Error("ctx1 should have identity1")
	}

	got2, err := Get(ctx2)
	if err != nil {
		t.Fatalf("Get(ctx2) failed: %v", err)
	}
	if got2 != identity2 {
		t.Error("ctx2 should have identity2")
	}

	// Original context should not have identity
	gotOrig := fromContext(ctx)
	if gotOrig != nil {
		t.Error("Original context should not have identity")
	}
}
