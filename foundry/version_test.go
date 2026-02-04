package foundry

import (
	"os"
	"strings"
	"testing"
)

func TestVersionConstantMatchesVERSIONFile(t *testing.T) {
	b, err := os.ReadFile("../VERSION")
	if err != nil {
		t.Fatalf("read VERSION: %v", err)
	}
	fileVersion := strings.TrimSpace(string(b))
	if fileVersion == "" {
		t.Fatalf("VERSION file is empty")
	}
	if version != fileVersion {
		t.Fatalf("foundry version constant=%q, VERSION=%q", version, fileVersion)
	}
}
