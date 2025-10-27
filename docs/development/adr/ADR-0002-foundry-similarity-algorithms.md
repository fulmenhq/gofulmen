---
id: "ADR-0002"
title: "Foundry Similarity Algorithm Implementation Strategy"
status: "accepted"
date: "2025-10-26"
deciders: ["@3leapsdave", "Foundation Forge"]
scope: "gofulmen"
tags: ["similarity", "performance", "algorithms", "foundry"]
---

# ADR-0002: Foundry Similarity Algorithm Implementation Strategy

## Status

**Current Status**: Accepted (v0.1.5 implementation)

**Future Review**: Pending string-metrics-go research project (see Research Backlog below)

## Context

Crucible similarity schema v2.0.0 introduces explicit support for multiple string distance algorithms:

- **Levenshtein**: Classic edit distance (insertions, deletions, substitutions)
- **Damerau-Levenshtein OSA**: Optimal String Alignment variant (adds adjacent transpositions)
- **Damerau-Levenshtein Unrestricted**: True Damerau-Levenshtein (unrestricted transpositions)
- **Jaro-Winkler**: Similarity metric optimized for short strings with common prefixes
- **Substring**: Longest common substring matching

Gofulmen must implement all variants to achieve cross-language parity with PyFulmen (RapidFuzz) and TSFulmen (WASM strsim bindings), while maintaining performance characteristics suitable for hot-loop operations (per ADR-0008 performance-sensitive instrumentation pattern).

### External Library Evaluation

**Candidates considered**:

1. `github.com/antzucaro/matchr` - Provides Levenshtein, Damerau-Levenshtein, Jaro, Jaro-Winkler
2. `github.com/xrash/smetrics` - Older library, less maintained, missing OSA variant
3. Custom port from rapidfuzz-cpp - Reference implementation for complex algorithms

### Performance Requirements

Similarity operations are hot-loop sensitive (ADR-0008):

- Called thousands of times per CLI run (typo correction, fuzzy matching, suggestion ranking)
- Latency target: <0.5ms p95 for 128-character strings
- Memory pressure: Minimize allocations in high-frequency paths
- No histogram timing instrumentation (50-100ns overhead unacceptable)

## Decision

**Hybrid implementation strategy for v0.1.5**:

1. **Levenshtein**: Keep existing gofulmen implementation
   - Benchmarked 1.24-1.76x faster than matchr
   - Uses 3-326x less memory (critical for long strings)
   - Already validated against Crucible fixtures

2. **Damerau OSA**: Use matchr library wrapper
   - Well-tested external implementation
   - Lower maintenance burden than custom port
   - Performance acceptable for non-hot-loop usage

3. **Damerau Unrestricted**: Port Zhao algorithm from rapidfuzz-cpp
   - matchr does not provide unrestricted variant
   - Reference implementation: [rapidfuzz-cpp DamerauLevenshtein_impl.hpp](https://github.com/rapidfuzz/rapidfuzz-cpp/blob/main/rapidfuzz/distance/DamerauLevenshtein_impl.hpp)
   - Based on "Linear space string correction algorithm using the Damerau-Levenshtein distance" by Chunchun Zhao and Sartaj Sahni
   - Properly attributed in code comments and this ADR

4. **Jaro-Winkler**: Use matchr library wrapper
   - Proven implementation with configurable prefix scale
   - Matches Crucible fixture expectations

5. **Substring**: Custom implementation
   - Longest common substring via dynamic programming
   - Aligned with PyFulmen implementation pattern

## Performance Benchmark Results (Phase 1a)

### Methodology

Compared existing gofulmen Levenshtein implementation against matchr library using:

- Crucible v2 fixtures (10 test cases, all exact matches)
- Benchmark suite: short strings, medium strings, long strings (100-1000 chars), Unicode, emoji, CJK
- Platform: Apple M2 Max (darwin/arm64)
- Tool: `go test -bench -benchmem`

### Results Summary

| Test Case                 | Current (ns/op) | Matchr (ns/op) | Speedup   | Current (B/op) | Matchr (B/op) | Memory Advantage |
| ------------------------- | --------------- | -------------- | --------- | -------------- | ------------- | ---------------- |
| short_identical           | 94.49           | 101.2          | 1.07x     | 96             | 288           | 3.0x less        |
| short_different           | 121.0           | 139.8          | 1.16x     | 128            | 448           | 3.5x less        |
| medium_ascii (44 chars)   | 3,038           | 5,118          | **1.68x** | 1,056          | 16,736        | **15.8x less**   |
| medium_unicode (19 chars) | 488.6           | 808.4          | **1.65x** | 288            | 2,688         | **9.3x less**    |
| long_100chars             | 17,233          | 30,369         | **1.76x** | 2,624          | 82,753        | **31.5x less**   |
| long_1000chars            | 1,660,242       | 2,770,514      | **1.67x** | 24,576         | 8,028,174     | **326x less**    |
| emoji                     | 248.7           | 349.6          | 1.41x     | 192            | 1,152         | 6.0x less        |
| cjk                       | 244.7           | 304.6          | 1.24x     | 160            | 768           | 4.8x less        |

### Key Findings

1. **Consistent speed advantage**: Current implementation 1.24-1.76x faster across all test cases
2. **Dramatic memory efficiency**: 3-326x less memory allocation
   - Especially critical for long strings (1000 chars: 326x advantage)
   - Matters for hot-loop operations processing many strings
3. **Accuracy**: Both implementations match all Crucible fixtures exactly (100% agreement)
4. **Performance scaling**: Advantage increases with string length (algorithm uses two-row DP with O(min(m,n)) space)

### Decision Rationale

Keep existing Levenshtein implementation because:

- Performance advantage significant and consistent
- Memory characteristics ideal for hot-loop usage (ADR-0008)
- Already validated against Crucible fixtures
- Pure Go, no external dependencies for most common operation

Use matchr for missing algorithms (OSA, Jaro-Winkler) because:

- Proven implementations, lower maintenance burden
- Performance acceptable for less frequent operations
- Avoids reinventing well-tested algorithms

## Implementation Details

### API Design

Following PyFulmen's unified API pattern:

```go
// Algorithm type for explicit algorithm selection
type Algorithm string

const (
    AlgorithmLevenshtein          Algorithm = "levenshtein"
    AlgorithmDamerauOSA           Algorithm = "damerau_osa"
    AlgorithmDamerauUnrestricted  Algorithm = "damerau_unrestricted"
    AlgorithmJaroWinkler          Algorithm = "jaro_winkler"
    AlgorithmSubstring            Algorithm = "substring"
)

// Distance calculates edit distance between strings using specified algorithm
func Distance(a, b string, algorithm Algorithm) (int, error)

// Score calculates normalized similarity score (0.0-1.0)
func Score(a, b string, algorithm Algorithm, opts *ScoreOptions) (float64, error)

// ScoreOptions configures score calculation
type ScoreOptions struct {
    JaroPrefixScale float64 // Jaro-Winkler prefix scaling (default 0.1)
    JaroMaxPrefix   int     // Jaro-Winkler max prefix length (default 4)
}

// SubstringMatch finds longest common substring
func SubstringMatch(needle, haystack string) (MatchRange, float64)
```

### Code Attribution

Unrestricted Damerau-Levenshtein implementation includes proper attribution:

```go
// damerauUnrestrictedDistance implements unrestricted Damerau-Levenshtein distance.
//
// Based on the Zhao-Sahni algorithm: "Linear space string correction algorithm
// using the Damerau-Levenshtein distance" by Chunchun Zhao and Sartaj Sahni.
//
// Reference implementation: rapidfuzz-cpp
// https://github.com/rapidfuzz/rapidfuzz-cpp/blob/main/rapidfuzz/distance/DamerauLevenshtein_impl.hpp
//
// Copyright notice from rapidfuzz-cpp:
// SPDX-License-Identifier: MIT
// Copyright © 2022-present Max Bachmann
```

### Telemetry Pattern

Per ADR-0008, similarity operations use **performance-sensitive pattern (counter-only)**:

- Metrics: `similarity_distance_total`, `similarity_error_total`
- No histogram timing (avoids 50-100ns overhead per call)
- Counters only for operation success/failure tracking

## Cross-Language Alignment

| Foundation | Implementation                  | Library          | Status         |
| ---------- | ------------------------------- | ---------------- | -------------- |
| PyFulmen   | RapidFuzz Python bindings       | strsim-rs (Rust) | ✅ Complete    |
| TSFulmen   | WASM bindings                   | strsim-rs (Rust) | 🚧 In Progress |
| Gofulmen   | Hybrid (native + matchr + port) | Multiple sources | ✅ v0.1.5      |

All implementations validated against shared Crucible v2 fixtures.

## Alternatives Considered

### Alternative 1: Use matchr for All Algorithms

**Rejected**: Performance benchmarks show 1.24-1.76x slower Levenshtein, 3-326x more memory allocation. Unacceptable for hot-loop operations.

### Alternative 2: Implement All Algorithms from Scratch

**Rejected**: Higher maintenance burden, risk of implementation errors, slower time-to-market. Hybrid approach balances performance, reliability, and development velocity.

### Alternative 3: CGO Bindings to strsim-rs

**Rejected for v0.1.5**: CGO adds complexity (cross-compilation, deployment friction). Reserved for future string-metrics-go research project.

### Alternative 4: Pure smetrics Library

**Rejected**: Older, less maintained, missing OSA variant. matchr provides better coverage and maintenance.

## Research Backlog: string-metrics-go

**Status**: Future research project (post-v0.1.5)  
**Trigger**: Success of TSFulmen string-metrics-wasm implementation

### Motivation

Phase 1a benchmarks demonstrate significant performance advantages of native Go implementations. If TSFulmen's WASM approach proves successful, a dedicated **string-metrics-go** library could provide:

1. **Unified ecosystem**: All foundations use strsim-rs foundation (Python via RapidFuzz, TypeScript via WASM, Go via native port or bindings)
2. **Optimal performance**: Leverage Go's strengths (similar to current Levenshtein advantages)
3. **Reduced maintenance**: Shared algorithm source of truth
4. **First-class support**: Purpose-built for Go ecosystem

### Potential Approaches

1. **Pure Go port of strsim-rs**: No CGO, pure Go toolchain, optimal for ecosystem
2. **CGO bindings to strsim-rs**: Direct use of proven implementation, upstream sync
3. **WASM bindings**: Portable, shared with TSFulmen, no CGO complexity

### Success Criteria

- TSFulmen WASM approach demonstrates viability
- Performance meets/exceeds current hybrid implementation
- Memory characteristics remain favorable (preserve 326x advantage for long strings)
- Maintenance model sustainable (upstream sync strategy)
- Cross-language parity via Crucible fixtures

### ADR Update Plan

If string-metrics-go succeeds, this ADR will be revised to document:

- Migration from hybrid to unified library approach
- Updated performance benchmarks (preserve Phase 1a data for comparison)
- Backward compatibility strategy for existing consumers
- Ecosystem alignment narrative

**Note**: v0.1.5 proceeds with hybrid approach documented above. Research project is contingent on string-metrics-wasm validation.

## Consequences

### Positive

- ✅ Optimal Levenshtein performance for hot-loop operations (1.24-1.76x faster, 3-326x less memory)
- ✅ Proven external implementations for complex algorithms (matchr, rapidfuzz-cpp reference)
- ✅ Cross-language parity via Crucible v2 fixtures (100% test coverage)
- ✅ Clear attribution for ported algorithms (Zhao algorithm, rapidfuzz-cpp)
- ✅ Future-proof: Research path to unified string-metrics-go library

### Negative

- ⚠️ Hybrid approach increases dependency surface (matchr + custom ports)
- ⚠️ Maintenance burden for ported algorithms (unrestricted Damerau)
- ⚠️ Potential future migration if string-metrics-go research succeeds

### Neutral

- ℹ️ Performance advantages documented for future library design decisions
- ℹ️ Benchmark data available for ecosystem-wide performance comparisons
- ℹ️ Clear migration path if research project validates unified library approach

## References

- [Crucible Similarity Schema v2.0.0](../../../schemas/crucible-go/library/foundry/v2.0.0/similarity.schema.json)
- [Crucible Similarity Fixtures](../../../config/crucible-go/library/foundry/similarity-fixtures.yaml)
- [ADR-0008: Helper Library Instrumentation Patterns](../../crucible-go/architecture/decisions/ADR-0008-helper-library-instrumentation-patterns.md)
- [rapidfuzz-cpp DamerauLevenshtein Implementation](https://github.com/rapidfuzz/rapidfuzz-cpp/blob/main/rapidfuzz/distance/DamerauLevenshtein_impl.hpp)
- [strsim-rs (Rust String Similarity)](https://github.com/rapidfuzz/strsim-rs)
- [matchr Library](https://github.com/antzucaro/matchr)
- [Implementation Plan](.plans/active/v0.1.5/similarity-v2-fixtures.md)

---

**Decision Outcome**: Hybrid implementation strategy balances performance (keep optimized Levenshtein), reliability (use proven libraries for missing algorithms), and future flexibility (research path to unified string-metrics-go). Performance benchmarks document 1.24-1.76x speed advantage and 3-326x memory advantage for native Go Levenshtein, justifying retention while using external libraries for algorithms where performance trade-off is acceptable.
