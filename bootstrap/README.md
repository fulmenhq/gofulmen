# Bootstrap Package

Simple, dependency-free tool installation for Go repositories.

## Philosophy

**Bootstrap goneat, then let goneat manage everything else.**

The bootstrap package provides minimal tooling installation capabilities using only the Go standard library. Its primary purpose is to install goneat (the Fulmen DX CLI), which then handles sophisticated tool management.

## Features

- ✅ **Stdlib only** - No external dependencies
- ✅ **Schema-compliant** - Uses `crucible/schemas/tooling/external-tools/v1.0.0`
- ✅ **Secure** - Checksum verification, path validation, HTTPS-only
- ✅ **Cross-platform** - macOS, Linux, Windows (with prerequisites)
- ✅ **Simple** - Minimal API, clear error messages

## Platform Support

### Tier 1: Full Support

- ✅ **macOS** (arm64, amd64) - tar/gzip built-in
- ✅ **Linux** (arm64, amd64) - tar/gzip built-in

### Tier 2: Partial Support

- ⚠️ **Windows** (amd64, arm64) - Requires manual prerequisites

**Windows users must install tar and gzip first:**

```powershell
# Install scoop
iex "& {$(irm get.scoop.sh)} -RunAsAdmin"

# Install prerequisites
scoop install tar gzip

# Then use bootstrap normally
go run github.com/fulmenhq/gofulmen/cmd/bootstrap --install
```

## Installation Types

### Type 1: `verify` - PATH Check

Verifies that a command exists in the system PATH.

**Schema:**

```yaml
- id: git
  description: Version control system
  required: true
  install:
    type: verify
    command: git
```

**Behavior:**

- Uses `exec.LookPath()` to check if command exists
- Fails with helpful suggestion if not found
- No installation performed (user must install manually)

### Type 2: `go` - Go Module Installation

Installs a Go module using `go install`.

**Schema:**

```yaml
- id: golangci-lint
  description: Go linter aggregator
  required: false
  install:
    type: go
    module: github.com/golangci/golangci-lint/cmd/golangci-lint
    version: v1.55.2
```

**Behavior:**

- Runs: `go install <module>@<version>`
- Verifies Go is installed
- Checks that binary is accessible after install
- Warns if GOPATH/bin not in PATH

### Type 3: `download` - Binary Download

Downloads, verifies, extracts, and installs a pre-built binary.

**Schema:**

```yaml
- id: goneat
  description: Fulmen DX CLI
  required: true
  install:
    type: download
    url: https://github.com/fulmenhq/goneat/releases/download/v0.2.11/goneat_v0.2.11_{{os}}_{{arch}}.tar.gz
    binName: goneat
    destination: ./bin
    checksum:
      darwin-arm64: c01581b7835362a2f12b41225cada8ca3af1dc65fc454a0c8e68374831869283
      darwin-amd64: f8bba61d9b729e156ea02d429a53798ef33ee9a57cf38c1a08d3d0428fc57033
      linux-arm64: e9fbe921a869625a7aef99b111cfbcfc0ce1da527f5ad1664307021472692e00
      linux-amd64: 35066fb7b6ba615e780aa29bb52ae88771340cccfb54b4e984d79f0e80474ad2
```

**Behavior:**

1. **Interpolate URL** - Replace `{{os}}` and `{{arch}}` with platform values
2. **Download** - HTTP GET to temp location (HTTPS only)
3. **Verify Checksum** - SHA-256 must match for current platform
4. **Extract Archive** - Supports `.tar.gz` and `.zip`
5. **Install Binary** - Move to destination directory
6. **Make Executable** - `chmod +x` on Unix platforms

**Security:**

- ✅ HTTPS URLs required (rejects HTTP)
- ✅ Checksum verification mandatory
- ✅ Path validation (prevents `../` traversal)
- ✅ Extraction size limits (prevents zip bombs)
- ✅ Symlinks skipped (prevents symlink attacks)

## Usage

### CLI

```bash
# Install all tools from manifest
go run github.com/fulmenhq/gofulmen/cmd/bootstrap --install

# Verify all tools are available
go run github.com/fulmenhq/gofulmen/cmd/bootstrap --verify

# Custom manifest path
go run github.com/fulmenhq/gofulmen/cmd/bootstrap --manifest /path/to/tools.yaml --install

# Verbose output
go run github.com/fulmenhq/gofulmen/cmd/bootstrap --install --verbose

# Force reinstall
go run github.com/fulmenhq/gofulmen/cmd/bootstrap --install --force
```

### Makefile Integration

```makefile
.PHONY: bootstrap tools

bootstrap: ## Install external tools
	@go run github.com/fulmenhq/gofulmen/cmd/bootstrap@latest --install --verbose

tools: ## Verify tools available
	@go run github.com/fulmenhq/gofulmen/cmd/bootstrap@latest --verify
```

### Programmatic Usage

```go
package main

import (
    "github.com/fulmenhq/gofulmen/bootstrap"
)

func main() {
    opts := bootstrap.Options{
        ManifestPath: ".crucible/tools.yaml",
        Verbose:      true,
    }

    if err := bootstrap.InstallTools(opts); err != nil {
        panic(err)
    }
}
```

## Manifest Format

Create `.crucible/tools.yaml`:

```yaml
version: v1.0.0
binDir: ./bin

tools:
  # Download pre-built binary
  - id: goneat
    description: Fulmen DX CLI for formatting, linting, security
    required: true
    install:
      type: download
      url: https://github.com/fulmenhq/goneat/releases/download/v0.2.11/goneat_v0.2.11_{{os}}_{{arch}}.tar.gz
      binName: goneat
      destination: ./bin
      checksum:
        darwin-arm64: c01581b7835362a2f12b41225cada8ca3af1dc65fc454a0c8e68374831869283
        darwin-amd64: f8bba61d9b729e156ea02d429a53798ef33ee9a57cf38c1a08d3d0428fc57033
        linux-arm64: e9fbe921a869625a7aef99b111cfbcfc0ce1da527f5ad1664307021472692e00
        linux-amd64: 35066fb7b6ba615e780aa29bb52ae88771340cccfb54b4e984d79f0e80474ad2

  # Install Go module
  - id: golangci-lint
    description: Go linter aggregator
    required: false
    install:
      type: go
      module: github.com/golangci/golangci-lint/cmd/golangci-lint
      version: v1.55.2

  # Verify system command
  - id: git
    description: Version control system
    required: true
    install:
      type: verify
      command: git
```

### URL Interpolation

Use `{{os}}` and `{{arch}}` placeholders in download URLs:

```yaml
url: https://example.com/tool_{{os}}_{{arch}}.tar.gz
```

**Platform mapping:**

| Runtime     | `{{os}}`  | `{{arch}}` | Result                                  |
| ----------- | --------- | ---------- | --------------------------------------- |
| macOS ARM   | `darwin`  | `arm64`    | `tool_darwin_arm64.tar.gz`              |
| macOS Intel | `darwin`  | `amd64`    | `tool_darwin_amd64.tar.gz`              |
| Linux ARM   | `linux`   | `arm64`    | `tool_linux_arm64.tar.gz`               |
| Linux x64   | `linux`   | `amd64`    | `tool_linux_amd64.tar.gz`               |
| Windows x64 | `windows` | `amd64`    | `tool_windows_amd64.tar.gz` (or `.zip`) |

### Checksums

Checksums use the `${os}-${arch}` key format:

```yaml
checksum:
  darwin-arm64: c01581b7835362a2f12b41225cada8ca...
  darwin-amd64: f8bba61d9b729e156ea02d429a53798e...
  linux-arm64: e9fbe921a869625a7aef99b111cfbcfc...
  linux-amd64: 35066fb7b6ba615e780aa29bb52ae887...
```

**Generate checksums:**

```bash
# macOS / Linux
shasum -a 256 goneat_darwin_arm64.tar.gz

# Or using Go
go run github.com/fulmenhq/gofulmen/cmd/bootstrap --compute-checksum file.tar.gz
```

## Error Messages

Bootstrap provides helpful, actionable error messages:

### Download Failure

```
❌ Failed to download goneat:
   Platform: darwin-arm64
   URL: https://github.com/fulmenhq/goneat/releases/download/v0.2.11/goneat_v0.2.11_darwin_arm64.tar.gz
   Error: 404 Not Found

   Possible solutions:
   - Check if the release exists for darwin-arm64
   - Verify the URL pattern in .crucible/tools.yaml
```

### Checksum Mismatch

```
❌ Checksum verification failed:
   File: /tmp/bootstrap-123/archive
   Expected: c01581b7835362a2f12b41225cada8ca...
   Actual:   XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX...

   This could indicate:
   - File was corrupted during download
   - File has been tampered with
   - Wrong checksum in manifest
```

### Windows Prerequisites

```
⚠️  Windows platform detected:
   Bootstrap requires tar and gzip to extract archives.

   Please install prerequisites:
   1. Install scoop: https://scoop.sh
   2. Run: scoop install tar gzip
   3. Then retry bootstrap
```

## Limitations

Bootstrap is intentionally minimal. For advanced features, use goneat:

| Feature                      | Bootstrap | Goneat |
| ---------------------------- | --------- | ------ |
| Install from GitHub releases | ✅        | ✅     |
| Checksum verification        | ✅        | ✅     |
| Version constraints          | ❌        | ✅     |
| Package managers (brew, apt) | ❌        | ✅     |
| Auto-update                  | ❌        | ✅     |
| Dependency resolution        | ❌        | ✅     |
| Platform-specific logic      | ❌        | ✅     |

**Recommended workflow:**

```bash
make bootstrap  # Installs goneat
goneat doctor   # Manages all other tools
```

## Testing

```bash
# Run all tests
go test ./bootstrap -v

# Run specific test
go test ./bootstrap -v -run TestInterpolateURL

# With coverage
go test ./bootstrap -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## API Reference

### Functions

```go
// InstallTools installs all tools from the manifest
func InstallTools(opts Options) error

// VerifyTools verifies all tools are available
func VerifyTools(opts Options) error

// GetPlatform returns the current OS and architecture
func GetPlatform() Platform

// InterpolateURL replaces {{os}} and {{arch}} in URLs
func InterpolateURL(urlTemplate string, platform Platform) string

// VerifySHA256 verifies a file's SHA-256 checksum
func VerifySHA256(filePath string, expectedHex string) error

// ComputeSHA256 computes a file's SHA-256 checksum
func ComputeSHA256(filePath string) (string, error)

// ExtractArchive extracts .tar.gz or .zip archives
func ExtractArchive(archivePath, destDir string) error

// LoadManifest loads and validates a tools manifest
func LoadManifest(path string) (*Manifest, error)
```

### Types

```go
type Options struct {
    ManifestPath string  // Path to tools.yaml (default: .crucible/tools.yaml)
    Force        bool    // Force reinstall
    Verbose      bool    // Verbose output
}

type Platform struct {
    OS   string  // darwin, linux, windows
    Arch string  // amd64, arm64
}

type Manifest struct {
    Version string
    BinDir  string
    Tools   []Tool
}

type Tool struct {
    ID          string
    Description string
    Required    bool
    Install     Install
}

type Install struct {
    Type        string            // verify, go, download
    Module      string            // For type: go
    Version     string            // For type: go
    Command     string            // For type: verify
    URL         string            // For type: download
    BinName     string            // For type: download
    Destination string            // For type: download
    Checksum    map[string]string // For type: download
}
```

## Examples

See the [implementation plan](../.plans/active/v0.1.0/bootstrap-implementation-plan.md) for detailed examples and design decisions.

## Contributing

Contributions are welcome! Please ensure:

- Tests pass: `go test ./bootstrap`
- Code follows Go standards
- Error messages are helpful and actionable
- Security best practices are maintained

## License

MIT License - see [LICENSE](../LICENSE) for details.
