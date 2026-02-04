package crucible

import (
	"os"
	"strings"
	"testing"
)

func TestGofulmenVersionConstantMatchesVERSIONFile(t *testing.T) {
	b, err := os.ReadFile("../VERSION")
	if err != nil {
		t.Fatalf("read VERSION: %v", err)
	}
	fileVersion := strings.TrimSpace(string(b))
	if fileVersion == "" {
		t.Fatalf("VERSION file is empty")
	}
	if GofulmenVersion != fileVersion {
		t.Fatalf("crucible GofulmenVersion=%q, VERSION=%q", GofulmenVersion, fileVersion)
	}
}
