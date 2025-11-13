---
title: "Fulpack Archive Module Standard"
description: "Canonical API specification for archive operations with Pathfinder integration"
author: "Schema Cartographer"
date: "2025-11-15"
version: "1.0.0"
status: "active"
tier: "common"
tags: ["archive", "compression", "pathfinder", "tar", "zip", "gzip"]
---

# Fulpack Archive Module Standard

## Overview

**Fulpack** is a Common-tier helper library module providing canonical archive operations (create, extract, scan, verify, info) with **mandatory Pathfinder integration** for glob-based file discovery within archives. This specification defines the cross-language API contract that all Fulmen helper libraries must implement.

**Module Tier**: Common
**Version**: 1.0.0
**Added**: Crucible v0.2.10
**Dependencies**: pathfinder (required)

## Architecture Principles

### Canonical Façade Principle

Per the [Fulmen Helper Library Standard](../../../architecture/fulmen-helper-library-standard.md#canonical-façade-principle), fulpack provides **canonical façades** wrapping standard library functionality to ensure:

1. **Cross-Language Interface Consistency** - Same operations, same error envelopes, same behavior
2. **Pathfinder Integration** - Unified glob search across filesystems and archives
3. **Security by Default** - Path traversal protection, decompression bomb detection
4. **Taxonomy-Driven Design** - Archive formats, operations, and entry types defined in SSOT

### Why Common Tier?

**Common tier assignment rationale**:

- **Universal need**: Most applications need basic archive handling (tar.gz, zip, gzip)
- **Zero external dependencies**: Uses language stdlib only (tarfile, zipfile, gzip in Python; archive/tar, archive/zip, compress/gzip in Go; tar-stream, archiver in TypeScript)
- **Pathfinder integration**: Enables unified file discovery API across filesystem and archives
- **Security requirement**: Consistent path traversal protection and bomb detection

**This is NOT a specialized module** despite wrapping stdlib—the Canonical Façade Principle mandates façades for universal capabilities regardless of implementation strategy.

## Taxonomy-Driven Design

### Archive Formats Taxonomy

**Location**: `schemas/taxonomy/library/fulpack/archive-formats/v1.0.0/formats.yaml`

**Supported formats** (v1.0.0):

- `tar.gz` - POSIX tar with gzip compression
- `zip` - ZIP archive with deflate compression
- `gzip` - GZIP compressed single file

**Format features**:

```yaml
tar.gz:
  preserves_permissions: true
  preserves_timestamps: true
  supports_symlinks: true
  supports_directories: true

zip:
  preserves_permissions: false # Limited on Windows
  preserves_timestamps: true
  supports_symlinks: false
  supports_directories: true

gzip:
  preserves_permissions: false
  preserves_timestamps: true
  supports_symlinks: false
  supports_directories: false # Single file only
```

Libraries generate enums from this taxonomy:

```python
# Generated in pyfulmen
class ArchiveFormat(Enum):
    TAR_GZ = "tar.gz"
    ZIP = "zip"
    GZIP = "gzip"
```

### Operations Taxonomy

**Location**: `schemas/taxonomy/library/fulpack/operations/v1.0.0/operations.yaml`

**Five canonical operations**:

1. `create` - Create archive from source files/directories
2. `extract` - Extract archive contents to destination
3. `scan` - List archive entries (for Pathfinder integration)
4. `verify` - Validate archive integrity and checksums
5. `info` - Get archive metadata without extraction

### Entry Types Taxonomy

**Location**: `schemas/taxonomy/library/fulpack/entry-types/v1.0.0/types.yaml`

**Canonical entry types**:

- `file` - Regular file with data
- `directory` - Directory/folder entry
- `symlink` - Symbolic link (requires security validation)

## Canonical API Specification

### 1. create() - Create Archive

**Signature** (TypeScript pseudocode):

```typescript
create(
  source: string | string[],
  output: string,
  format: ArchiveFormat,
  options?: CreateOptions
): ArchiveInfo
```

**Parameters**:

- `source`: Single path or array of paths to archive
- `output`: Output archive file path
- `format`: `ArchiveFormat.TAR_GZ`, `ArchiveFormat.ZIP`, or `ArchiveFormat.GZIP`
- `options`: Optional creation options

**CreateOptions** (from `schemas/library/fulpack/v1.0.0/create-options.schema.json`):

```typescript
{
  compression_level?: number,       // 1-9 (default: 6)
  include_patterns?: string[],      // e.g., ["**/*.py", "**/*.md"]
  exclude_patterns?: string[],      // e.g., ["**/__pycache__", "**/.git"]
  checksum_algorithm?: string,      // "sha256" | "sha512" | "sha1" | "md5" (default: "sha256")
  preserve_permissions?: boolean,   // default: true
  follow_symlinks?: boolean         // default: false
}
```

**Returns**: `ArchiveInfo` with metadata (entry_count, sizes, checksums)

**Security**:

- Validates paths before archiving
- Applies path traversal protection
- Symlinks: Only followed if `follow_symlinks: true`

**Example** (Python):

```python
from pyfulmen import fulpack
from pyfulmen.fulpack import ArchiveFormat

info = fulpack.create(
    source=["src/", "docs/", "README.md"],
    output="release.tar.gz",
    format=ArchiveFormat.TAR_GZ,
    options={
        "exclude_patterns": ["**/__pycache__", "**/.git"],
        "compression_level": 9
    }
)
print(f"Created archive with {info.entry_count} entries, {info.compression_ratio}x compression")
```

---

### 2. extract() - Extract Archive

**Signature**:

```typescript
extract(
  archive: string,
  destination: string,
  options?: ExtractOptions
): ExtractResult
```

**Parameters**:

- `archive`: Archive file path
- `destination`: Target directory (must be explicit, no CWD extraction)
- `options`: Optional extraction options

**ExtractOptions** (from `schemas/library/fulpack/v1.0.0/extract-options.schema.json`):

```typescript
{
  overwrite?: "error" | "skip" | "overwrite",  // default: "error"
  verify_checksums?: boolean,                   // default: true
  preserve_permissions?: boolean,               // default: true
  include_patterns?: string[],                  // e.g., ["**/*.csv"]
  max_size?: number,                            // default: 1GB (bomb protection)
  max_entries?: number                          // default: 10000 (bomb protection)
}
```

**Returns**: `ExtractResult` with extracted_count, skipped, errors

**Security** (MANDATORY for all implementations):

- **Path traversal protection**: Reject `../`, absolute paths, escapes outside destination
- **Symlink validation**: Reject symlinks targeting outside destination
- **Decompression bomb protection**: Enforce `max_size` and `max_entries` limits
- **Checksum verification**: Verify checksums if present (unless disabled)

**Example** (Go):

```go
import "github.com/fulmenhq/gofulmen/fulpack"

result, err := fulpack.Extract(
    "data.tar.gz",
    "/tmp/extracted",
    &fulpack.ExtractOptions{
        IncludePatterns: []string{"**/*.csv"},
        VerifyChecksums: true,
    },
)
if err != nil {
    return err
}
fmt.Printf("Extracted %d files, skipped %d\n", result.ExtractedCount, result.SkippedCount)
```

---

### 3. scan() - Scan Archive (Pathfinder Integration)

**Signature**:

```typescript
scan(
  archive: string,
  options?: ScanOptions
): ArchiveEntry[]
```

**Parameters**:

- `archive`: Archive file path
- `options`: Optional scan options

**ScanOptions** (from `schemas/library/fulpack/v1.0.0/scan-options.schema.json`):

```typescript
{
  include_metadata?: boolean,       // default: true
  entry_types?: string[],           // ["file", "directory", "symlink"]
  max_depth?: number | null,        // default: null (unlimited)
  max_entries?: number              // default: 100000 (safety limit)
}
```

**Returns**: Array of `ArchiveEntry` objects with:

- `path` - Normalized entry path
- `type` - "file" | "directory" | "symlink"
- `size` - Uncompressed size
- `compressed_size` - Compressed size (if available)
- `modified` - Modification timestamp
- `checksum` - SHA-256 checksum (if available)
- `mode` - Unix permissions (if available)

**Purpose**: Enables Pathfinder glob searches within archives without extraction

**Performance**: Lazy evaluation—reads archive TOC/directory, does NOT extract files

**Pathfinder Integration Pattern**:

```python
from pyfulmen import fulpack, pathfinder

# Pathfinder delegates to fulpack for archives
entries = fulpack.scan("data.tar.gz")
matches = pathfinder.glob(entries, "**/*.csv")  # Pathfinder applies glob

# Or use Pathfinder's unified API
results = pathfinder.find(
    source="data.tar.gz",
    pattern="**/*.csv"
)
# Pathfinder detects archive, calls fulpack.scan(), applies pattern
```

**Example** (TypeScript):

```typescript
import { fulpack, ArchiveFormat } from "@fulmenhq/tsfulmen/fulpack";

const entries = fulpack.scan("data.tar.gz", {
  includeMetadata: true,
  entryTypes: ["file"], // Only files, no directories
});

// Filter entries by pattern (or let Pathfinder do it)
const csvFiles = entries.filter((e) => e.path.endsWith(".csv"));
console.log(`Found ${csvFiles.length} CSV files`);
```

---

### 4. verify() - Verify Archive Integrity

**Signature**:

```typescript
verify(
  archive: string,
  options?: VerifyOptions
): ValidationResult
```

**Parameters**:

- `archive`: Archive file path
- `options`: Optional verification options (future use)

**Returns**: `ValidationResult` with:

- `valid` - Boolean indicating if archive is intact
- `errors` - Array of validation errors
- `warnings` - Array of warnings (e.g., missing checksums)
- `entry_count` - Number of entries validated
- `checksums_verified` - Count of checksum validations
- `checks_performed` - List of checks (structure_valid, checksums_verified, no_path_traversal, no_decompression_bomb, symlinks_safe)

**Security checks**:

- Archive structure integrity
- Checksum verification (if present)
- Path traversal detection
- Decompression bomb detection (compression ratio, entry count)
- Symlink safety (targets within bounds)

**Example** (Python):

```python
result = fulpack.verify("data.tar.gz")
if not result.valid:
    print(f"Archive validation failed: {result.errors}")
else:
    print(f"Archive valid: {result.entry_count} entries, {result.checksums_verified} checksums verified")
    if result.warnings:
        print(f"Warnings: {result.warnings}")
```

---

### 5. info() - Get Archive Metadata

**Signature**:

```typescript
info(
  archive: string
): ArchiveInfo
```

**Parameters**:

- `archive`: Archive file path

**Returns**: `ArchiveInfo` with:

- `format` - Detected archive format (enum)
- `compression` - Compression algorithm (enum)
- `entry_count` - Total entries
- `total_size` - Uncompressed total size
- `compressed_size` - Archive file size
- `compression_ratio` - Ratio calculation
- `has_checksums` - Boolean
- `checksum_algorithm` - Algorithm used (if has_checksums)
- `created` - Archive creation timestamp (if available)

**Use cases**:

- Quick inspection without extraction
- Format detection
- Size estimation before extraction
- Compression ratio analysis

**Example** (Go):

```go
info, err := fulpack.Info("release.tar.gz")
if err != nil {
    return err
}
fmt.Printf("Format: %s, Entries: %d, Compression: %.1fx\n",
    info.Format, info.EntryCount, info.CompressionRatio)
```

---

## Streaming API (Planned - Implementation Deferred)

**Status**: API reserved, implementation deferred to v0.2.11+

**Why plan now**: Ensures block API doesn't prevent streaming later; reserves method names; defines resource cleanup patterns

**Streaming operations** (future):

- `create_stream()` - Streaming archive creation
- `extract_stream()` - Streaming extraction
- `scan_stream()` - Streaming scan (large archives)

**Language-specific patterns**:

- **Python**: Context managers (`with`), generators, async iterators
- **Go**: `io.Reader`/`io.Writer`, `defer` cleanup, channel-based iteration
- **TypeScript**: Async iterators, `ReadableStream`/`WritableStream`

### Forward Compatibility Confirmation

**Schema Compatibility**: Current schemas are designed to support streaming without breaking changes

**No schema migrations required** when adding streaming in v0.2.11:

1. **Operation schemas remain unchanged**: Streaming variants use same option/result schemas
   - `CreateOptions` works for both `create()` and `create_stream()`
   - `ExtractResult` works for both `extract()` and `extract_stream()`

2. **New method names**: Streaming uses distinct names (`*_stream`)
   - No conflicts with existing methods
   - Both APIs can coexist in same module

3. **Resource cleanup**: Schemas don't dictate cleanup patterns
   - Python: Add context manager protocol to stream objects
   - Go: Add `Close()` method to stream types
   - TypeScript: Add `finally()` to promise chains

4. **Validation unchanged**: Same schemas validate both modes
   - Block API: Full object validation
   - Stream API: Per-entry validation

**Example (forward-compatible implementation)**:

```python
# v1.0.0 (block API)
result = fulpack.extract("archive.tar.gz", "/dest", options)

# v2.0.0 (streaming API added, no schema changes)
with fulpack.extract_stream("archive.tar.gz", "/dest", options) as stream:
    for entry in stream:
        # Process entry with same ArchiveEntry schema
        pass
# Returns same ExtractResult schema
```

**Conclusion**: Current v1.0.0 schemas are streaming-ready. No breaking changes needed for v0.2.11 streaming implementation.

See feature brief for detailed streaming API specification.

---

## Security Model

### Mandatory Protections

**ALL implementations MUST enforce these protections**:

#### 1. Path Traversal Protection

- Normalize all entry paths
- Reject absolute paths (e.g., `/etc/passwd`)
- Reject paths containing `../` (parent directory traversal)
- Validate symlink targets are within archive/destination bounds
- Enforce destination directory bounds during extraction

#### 2. Decompression Bomb Protection

- Max entry size limit (default: 1GB)
- Max total size limit (default: 10GB)
- Max entries limit (default: 100k)
- Monitor compression ratio (warn if >100:1)

#### 3. Checksum Verification

- Verify checksums if present in archive
- Report missing checksums as warnings (not errors)
- Fail extraction on checksum mismatch (unless `verify_checksums: false`)

#### 4. Safe Defaults

- No extraction to CWD (require explicit destination)
- Error on overwrite conflicts (unless `overwrite: "skip"` or `"overwrite"`)
- Preserve permissions only if requested
- Don't follow symlinks unless `follow_symlinks: true`

### Security Test Requirements

All implementations MUST pass security tests for:

- Path traversal attempts (`../../../etc/passwd`)
- Absolute path attacks (`/etc/passwd`)
- Symlink escapes (symlink targeting outside bounds)
- Decompression bombs (10MB → 10GB expansion)
- Checksum tampering detection

---

## Pathfinder Integration Specification

### Integration Pattern

**Pathfinder** provides a unified file discovery API that works seamlessly across:

- Local filesystems
- Archives (via fulpack)
- Network resources (future)

**Workflow**:

1. User calls `pathfinder.find("data.tar.gz", "**/*.csv")`
2. Pathfinder detects archive format
3. Pathfinder calls `fulpack.scan("data.tar.gz", options)`
4. Fulpack returns `ArchiveEntry[]` with normalized paths
5. Pathfinder applies glob pattern to entry paths
6. Pathfinder returns matching entries

**No extraction required**: `scan()` only reads archive TOC/directory

**Performance target**: <1s for TOC read, regardless of archive size

### Edge Case Handling (Cross-Language Determinism)

**CRITICAL**: All implementations MUST handle these edge cases identically to ensure Pathfinder returns the same results across Go/Python/TypeScript:

#### 1. Symlinks in Archives

**Behavior**: `scan()` MUST include symlink entries in results with `type: "symlink"`

- Include `symlink_target` field with original target path (not resolved)
- DO NOT follow symlinks during scan (no recursive resolution)
- DO NOT validate symlink targets during scan (validation happens in `verify()` or `extract()`)
- Pathfinder applies glob to symlink paths, not targets

**Rationale**: Symlink validation is security-critical but separate from discovery. `scan()` provides raw TOC; `verify()`/`extract()` enforce security.

**Example**:

```yaml
# Archive contains: docs/link.md -> ../secret.txt
# scan() returns:
- path: "docs/link.md"
  type: "symlink"
  symlink_target: "../secret.txt" # Original target, not resolved
  size: 0
```

#### 2. Invalid UTF-8 Paths

**Behavior**: `scan()` MUST handle invalid UTF-8 in entry paths deterministically

- **Preferred**: Normalize to UTF-8 using replacement character (U+FFFD)
- **Alternative**: Base64-encode invalid paths with prefix `base64:` (for exact preservation)
- **Document choice**: Each language documents its approach in implementation notes
- **Cross-language tests**: Fixtures MUST include invalid UTF-8 paths to verify consistency

**Rationale**: Archives may contain non-UTF-8 paths (legacy encodings, binary names). Implementations must handle gracefully without crashing.

**Example (replacement character approach)**:

```yaml
# Archive contains: data/file_\xFF\xFE.txt (invalid UTF-8)
# scan() returns:
- path: "data/file_��.txt" # U+FFFD replacement characters
  type: "file"
```

#### 3. Absolute Paths

**Behavior**: `scan()` MUST normalize absolute paths to relative

- Strip leading `/` or drive letters (Windows: `C:/`)
- Emit warning in scan results
- Include in results (don't skip) so Pathfinder can match
- `extract()` and `verify()` MUST reject absolute paths (security)

**Rationale**: `scan()` is discovery; `extract()` is security enforcement. Separation of concerns.

#### 4. Path Traversal Attempts

**Behavior**: `scan()` MUST include `../` paths in results

- DO NOT normalize or reject during scan
- Include in results so Pathfinder can discover them
- `verify()` MUST flag as security warnings
- `extract()` MUST reject with error

**Rationale**: Security tests need to verify that archives with malicious paths are detected and rejected.

### Example Integration

```python
from pyfulmen import pathfinder

# Unified API - pathfinder handles routing
files = pathfinder.find(
    source="data.tar.gz",
    pattern="**/*.csv",
    options={"include_metadata": True}
)

for file in files:
    print(f"{file.path}: {file.size} bytes")
```

---

## Implementation Guidance

### Language-Specific Notes

**Go** (`github.com/fulmenhq/gofulmen/fulpack`):

- Use `archive/tar`, `archive/zip`, `compress/gzip` from stdlib
- Errors returned as `error` type with wrapping
- Enums generated from taxonomy YAML

**Python** (`pyfulmen.fulpack`):

- Use `tarfile`, `zipfile`, `gzip` from stdlib
- Enums generated from taxonomy YAML
- Context managers for file handles
- Exceptions raised for errors

**TypeScript** (`@fulmenhq/tsfulmen/fulpack`):

- Use `tar-stream` and `archiver` for cross-platform compatibility
- Enums generated from taxonomy YAML
- Promises for async operations
- Errors thrown as `Error` instances

### Error Handling

All implementations MUST:

- Use Foundry error schemas for consistency
- Provide clear error messages with context
- Distinguish between validation errors (invalid input) and runtime errors (I/O failures)
- Return structured error envelopes with exit codes

#### Canonical Error Envelope

All fulpack errors MUST use this envelope structure (compatible with Foundry error schemas):

```typescript
interface FulpackError {
  code: string; // Canonical error code (see below)
  message: string; // Human-readable message
  path?: string; // Entry path that caused error (if applicable)
  archive?: string; // Archive file path
  operation: string; // Operation name (create, extract, scan, verify, info)
  details?: {
    // Optional context
    entry_index?: number;
    compression_ratio?: number;
    actual_size?: number;
    max_size?: number;
  };
}
```

#### Canonical Error Codes

**Validation Errors** (invalid input):

- `INVALID_ARCHIVE_FORMAT` - Archive format not recognized
- `INVALID_PATH` - Entry path contains invalid characters
- `INVALID_OPTIONS` - Invalid options passed to operation

**Security Errors** (protection triggered):

- `PATH_TRAVERSAL` - Entry path attempts directory traversal (`../`)
- `ABSOLUTE_PATH` - Entry path is absolute (`/etc/passwd`)
- `SYMLINK_ESCAPE` - Symlink target outside archive/destination bounds
- `DECOMPRESSION_BOMB` - Archive exceeds size/entry limits
- `CHECKSUM_MISMATCH` - Entry checksum verification failed

**Runtime Errors** (I/O failures):

- `ARCHIVE_NOT_FOUND` - Archive file does not exist
- `ARCHIVE_CORRUPT` - Archive structure is invalid/corrupted
- `EXTRACTION_FAILED` - Failed to write extracted file
- `PERMISSION_DENIED` - Insufficient permissions to read/write
- `DISK_FULL` - Insufficient disk space for extraction

**Example Usage**:

```python
# Decompression bomb detection
raise FulpackError(
    code="DECOMPRESSION_BOMB",
    message="Archive exceeds maximum size limit",
    archive="malicious.tar.gz",
    operation="extract",
    details={
        "actual_size": 10737418240,  # 10GB
        "max_size": 1073741824,      # 1GB limit
        "compression_ratio": 1000
    }
)

# Checksum mismatch
raise FulpackError(
    code="CHECKSUM_MISMATCH",
    message="Entry checksum verification failed",
    archive="data.tar.gz",
    path="data/corrupted.csv",
    operation="extract",
    details={"entry_index": 42}
)
```

### Testing Requirements

All implementations MUST provide:

- **Unit tests**: Each operation tested in isolation
- **Security tests**: Path traversal, bombs, checksums
- **Integration tests**: Pathfinder integration
- **Fixture tests**: Using fixtures from `config/library/fulpack/fixtures/`

**Test coverage target**: ≥70%

---

## Test Fixtures

**Location**: `config/library/fulpack/fixtures/`

**Three canonical fixtures**:

1. **basic.tar.gz** - Normal archive structure
   - 10 files in simple directory tree
   - All formats supported (tar.gz, zip versions)
   - Used for basic operation testing

2. **nested.zip** - 3-level directory nesting
   - Tests deep directory traversal
   - Tests path normalization
   - Tests scan with `max_depth` option

3. **pathological.tar.gz** - Security test cases
   - Contains path traversal attempts (`../../../etc/passwd`)
   - Contains absolute paths (`/etc/passwd`)
   - Contains symlink escapes
   - MUST be rejected by extract/scan operations

### Fixture Governance

**Adding New Fixtures**:

1. **Naming Convention**: `{category}-{description}.{format}`
   - Categories: `basic`, `nested`, `pathological`, `utf8`, `symlink`, `large`, `corrupt`
   - Examples: `pathological-traversal.tar.gz`, `utf8-invalid-paths.zip`, `large-10k-entries.tar.gz`

2. **Approval Process**:
   - Create fixture locally and test in your library
   - Document fixture purpose and expected behavior in PR description
   - Add fixture to `config/library/fulpack/fixtures/`
   - Add test cases to validate fixture behavior
   - Request review from Schema Cartographer before merging to Crucible

3. **Documentation**: Each fixture should have an accompanying `.txt` or `.md` file describing:
   - Purpose (what it tests)
   - Expected behavior (pass/fail conditions)
   - Contents summary (number of files, structure)
   - Special characteristics (invalid UTF-8, symlinks, etc.)

4. **Size Limits**:
   - Basic fixtures: <10KB
   - Pathological fixtures: <50KB
   - Large fixtures (if needed): <1MB, must justify in PR

5. **Cross-Language Parity**:
   - New fixtures MUST work identically across all language implementations
   - Include fixture in parity test suites
   - Document any language-specific behavior (e.g., UTF-8 handling differences)

**Example fixture documentation** (`pathological-traversal.tar.gz.txt`):

```
Fixture: pathological-traversal.tar.gz
Purpose: Test path traversal protection
Expected: extract() and verify() MUST reject this archive
Contents:
  - safe.txt (normal file)
  - ../../../etc/passwd (traversal attempt)
  - /root/.ssh/id_rsa (absolute path)
Behavior:
  - scan() MUST list all entries (including malicious paths)
  - verify() MUST return valid=false with PATH_TRAVERSAL errors
  - extract() MUST fail with PATH_TRAVERSAL error
```

---

## Schema References

**Taxonomy schemas**:

- `schemas/taxonomy/library/fulpack/archive-formats/v1.0.0/formats.yaml`
- `schemas/taxonomy/library/fulpack/operations/v1.0.0/operations.yaml`
- `schemas/taxonomy/library/fulpack/entry-types/v1.0.0/types.yaml`

**Data structure schemas**:

- `schemas/library/fulpack/v1.0.0/archive-info.schema.json`
- `schemas/library/fulpack/v1.0.0/archive-entry.schema.json`
- `schemas/library/fulpack/v1.0.0/archive-manifest.schema.json`
- `schemas/library/fulpack/v1.0.0/validation-result.schema.json`
- `schemas/library/fulpack/v1.0.0/create-options.schema.json`
- `schemas/library/fulpack/v1.0.0/extract-options.schema.json`
- `schemas/library/fulpack/v1.0.0/scan-options.schema.json`
- `schemas/library/fulpack/v1.0.0/extract-result.schema.json`

---

## Version History

- **1.0.0** (2025-11-15) - Initial specification
  - 3 archive formats (tar.gz, zip, gzip)
  - 5 operations (create, extract, scan, verify, info)
  - Pathfinder integration
  - Security model defined
  - Streaming API reserved (implementation deferred)

---

## Future Enhancements

**Planned for fulpack v2.0.0+**:

- Streaming API implementation (create_stream, extract_stream, scan_stream)
- Archive composition (nested archives)

**Deferred to fulpack-formats (Specialized tier)**:

- Exotic formats (tar.zst, tar.xz, tar.bz2, 7z, rar)
- Requires heavy external dependencies

**Deferred to fulpack-secure (Specialized tier)**:

- Encryption (zip AES-256, GPG)
- Digital signatures (detached signatures)

---

## Related Standards

- [Fulmen Helper Library Standard](../../../architecture/fulmen-helper-library-standard.md)
- [Canonical Façade Principle](../../../architecture/fulmen-helper-library-standard.md#canonical-façade-principle)
- [Pathfinder Module Standard](pathfinder.md)
- [Module Registry](../../../../config/taxonomy/library/platform-modules/v1.0.0/modules.yaml)

---

**Status**: Active
**Tier**: Common
**Version**: 1.0.0
**Last Updated**: 2025-11-15
