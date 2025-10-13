package bootstrap

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestGetPlatform(t *testing.T) {
	platform := GetPlatform()

	if platform.OS == "" {
		t.Error("Platform OS should not be empty")
	}

	if platform.Arch == "" {
		t.Error("Platform Arch should not be empty")
	}

	expectedOS := normalizeOS(runtime.GOOS)
	if platform.OS != expectedOS {
		t.Errorf("Expected OS %s, got %s", expectedOS, platform.OS)
	}

	expectedArch := normalizeArch(runtime.GOARCH)
	if platform.Arch != expectedArch {
		t.Errorf("Expected Arch %s, got %s", expectedArch, platform.Arch)
	}
}

func TestPlatformString(t *testing.T) {
	platform := Platform{OS: "darwin", Arch: "arm64"}
	expected := "darwin-arm64"

	if platform.String() != expected {
		t.Errorf("Expected %s, got %s", expected, platform.String())
	}
}

func TestInterpolateURL(t *testing.T) {
	tests := []struct {
		name     string
		template string
		platform Platform
		expected string
	}{
		{
			name:     "Basic interpolation",
			template: "https://example.com/tool_{{os}}_{{arch}}.tar.gz",
			platform: Platform{OS: "darwin", Arch: "arm64"},
			expected: "https://example.com/tool_darwin_arm64.tar.gz",
		},
		{
			name:     "No placeholders",
			template: "https://example.com/tool.tar.gz",
			platform: Platform{OS: "linux", Arch: "amd64"},
			expected: "https://example.com/tool.tar.gz",
		},
		{
			name:     "Multiple occurrences",
			template: "https://{{os}}.example.com/{{arch}}/tool_{{os}}_{{arch}}.tar.gz",
			platform: Platform{OS: "linux", Arch: "amd64"},
			expected: "https://linux.example.com/amd64/tool_linux_amd64.tar.gz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := InterpolateURL(tt.template, tt.platform)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestIsPlatformSupported(t *testing.T) {
	tests := []struct {
		name            string
		platform        Platform
		expectSupported bool
		expectMessage   bool
	}{
		{
			name:            "macOS supported",
			platform:        Platform{OS: "darwin", Arch: "arm64"},
			expectSupported: true,
			expectMessage:   false,
		},
		{
			name:            "Linux supported",
			platform:        Platform{OS: "linux", Arch: "amd64"},
			expectSupported: true,
			expectMessage:   false,
		},
		{
			name:            "Windows supported with warning",
			platform:        Platform{OS: "windows", Arch: "amd64"},
			expectSupported: true,
			expectMessage:   true,
		},
		{
			name:            "Unsupported OS",
			platform:        Platform{OS: "freebsd", Arch: "amd64"},
			expectSupported: false,
			expectMessage:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			supported, msg := IsPlatformSupported(tt.platform)

			if supported != tt.expectSupported {
				t.Errorf("Expected supported=%v, got %v", tt.expectSupported, supported)
			}

			if tt.expectMessage && msg == "" {
				t.Error("Expected a message but got empty string")
			}

			if !tt.expectMessage && msg != "" {
				t.Errorf("Expected no message but got: %s", msg)
			}
		})
	}
}

func TestVerifySHA256(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")

	content := []byte("hello world")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	validChecksum := "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"

	t.Run("Valid checksum", func(t *testing.T) {
		err := VerifySHA256(testFile, validChecksum)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})

	t.Run("Invalid checksum", func(t *testing.T) {
		invalidChecksum := "0000000000000000000000000000000000000000000000000000000000000000"
		err := VerifySHA256(testFile, invalidChecksum)

		if err == nil {
			t.Error("Expected error for invalid checksum")
		}

		if _, ok := err.(*ChecksumMismatchError); !ok {
			t.Errorf("Expected ChecksumMismatchError, got %T", err)
		}
	})

	t.Run("File not found", func(t *testing.T) {
		err := VerifySHA256("/nonexistent/file", validChecksum)
		if err == nil {
			t.Error("Expected error for nonexistent file")
		}
	})
}

func TestComputeSHA256(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")

	content := []byte("hello world")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	expected := "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"

	hash, err := ComputeSHA256(testFile)
	if err != nil {
		t.Fatalf("ComputeSHA256 failed: %v", err)
	}

	if hash != expected {
		t.Errorf("Expected hash %s, got %s", expected, hash)
	}
}

func TestValidatePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "Valid relative path",
			path:    "bin/goneat",
			wantErr: false,
		},
		{
			name:    "Valid nested path",
			path:    "usr/local/bin/tool",
			wantErr: false,
		},
		{
			name:    "Path traversal with ..",
			path:    "../../../etc/passwd",
			wantErr: true,
		},
		{
			name:    "Absolute path",
			path:    "/usr/bin/tool",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadManifest(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("Valid manifest", func(t *testing.T) {
		manifestPath := filepath.Join(tempDir, "valid.yaml")
		content := `version: v1.0.0
binDir: ./bin
tools:
  - id: goneat
    description: Fulmen DX CLI
    required: true
    install:
      type: download
      url: https://example.com/goneat_{{os}}_{{arch}}.tar.gz
      binName: goneat
      checksum:
        darwin-arm64: abc123
`

		if err := os.WriteFile(manifestPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write manifest: %v", err)
		}

		manifest, err := LoadManifest(manifestPath)
		if err != nil {
			t.Fatalf("LoadManifest failed: %v", err)
		}

		if manifest.Version != "v1.0.0" {
			t.Errorf("Expected version v1.0.0, got %s", manifest.Version)
		}

		if len(manifest.Tools) != 1 {
			t.Fatalf("Expected 1 tool, got %d", len(manifest.Tools))
		}

		tool := manifest.Tools[0]
		if tool.ID != "goneat" {
			t.Errorf("Expected tool ID goneat, got %s", tool.ID)
		}

		if tool.Install.Type != "download" {
			t.Errorf("Expected type download, got %s", tool.Install.Type)
		}
	})

	t.Run("Missing version", func(t *testing.T) {
		manifestPath := filepath.Join(tempDir, "no-version.yaml")
		content := `binDir: ./bin
tools:
  - id: test
    install:
      type: verify
      command: git
`

		if err := os.WriteFile(manifestPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write manifest: %v", err)
		}

		_, err := LoadManifest(manifestPath)
		if err == nil {
			t.Error("Expected error for missing version")
		}
	})

	t.Run("Invalid YAML", func(t *testing.T) {
		manifestPath := filepath.Join(tempDir, "invalid.yaml")
		content := `this is not valid: [yaml syntax`

		if err := os.WriteFile(manifestPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write manifest: %v", err)
		}

		_, err := LoadManifest(manifestPath)
		if err == nil {
			t.Error("Expected error for invalid YAML")
		}
	})
}

func TestValidateTool(t *testing.T) {
	tests := []struct {
		name    string
		tool    Tool
		wantErr bool
	}{
		{
			name: "Valid go type",
			tool: Tool{
				ID: "test",
				Install: Install{
					Type:    "go",
					Module:  "github.com/example/tool",
					Version: "v1.0.0",
				},
			},
			wantErr: false,
		},
		{
			name: "Valid verify type",
			tool: Tool{
				ID: "test",
				Install: Install{
					Type:    "verify",
					Command: "git",
				},
			},
			wantErr: false,
		},
		{
			name: "Valid download type",
			tool: Tool{
				ID: "test",
				Install: Install{
					Type:    "download",
					URL:     "https://example.com/tool.tar.gz",
					BinName: "tool",
				},
			},
			wantErr: false,
		},
		{
			name: "Missing ID",
			tool: Tool{
				Install: Install{
					Type:    "verify",
					Command: "git",
				},
			},
			wantErr: true,
		},
		{
			name: "Go type missing module",
			tool: Tool{
				ID: "test",
				Install: Install{
					Type:    "go",
					Version: "v1.0.0",
				},
			},
			wantErr: true,
		},
		{
			name: "Verify type missing command",
			tool: Tool{
				ID: "test",
				Install: Install{
					Type: "verify",
				},
			},
			wantErr: true,
		},
		{
			name: "Download type missing URL",
			tool: Tool{
				ID: "test",
				Install: Install{
					Type:    "download",
					BinName: "tool",
				},
			},
			wantErr: true,
		},
		{
			name: "Unsupported type",
			tool: Tool{
				ID: "test",
				Install: Install{
					Type: "unknown",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTool(&tt.tool)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateTool() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
