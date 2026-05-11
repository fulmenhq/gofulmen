package fulpack

import (
	"os"
	"testing"
)

func TestCheckedUint64ToInt64RejectsOverflow(t *testing.T) {
	_, err := checkedUint64ToInt64(maxInt64Uint+1, "zip uncompressed size")
	if err == nil {
		t.Fatal("expected overflow error")
	}
}

func TestCheckedAddUint64RejectsOverflow(t *testing.T) {
	_, err := checkedAddUint64(int64(maxInt64Uint), 1, "zip uncompressed size")
	if err == nil {
		t.Fatal("expected overflow error")
	}
}

func TestTarPermissionModeMasksToPermissionBits(t *testing.T) {
	mode := tarPermissionMode(0o100777)
	if mode != 0o777 {
		t.Fatalf("mode = %v, want %v", mode, os.FileMode(0o777))
	}
}

func TestExtractionModeFallsBackForZeroPermissions(t *testing.T) {
	mode := extractionMode(0, 0o644, true)
	if mode != 0o644 {
		t.Fatalf("mode = %v, want %v", mode, os.FileMode(0o644))
	}
}
