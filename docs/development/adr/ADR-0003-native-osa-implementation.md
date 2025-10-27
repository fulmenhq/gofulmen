---
id: "ADR-0003"
title: "Native Go OSA Implementation to Replace matchr"
status: "accepted"
date: "2025-10-27"
deciders: ["@3leapsdave", "Foundation Forge"]
scope: "gofulmen"
tags: ["similarity", "osa", "algorithms", "performance", "foundry"]
related_adrs: ["ADR-0002"]
---

# ADR-0003: Native Go OSA Implementation to Replace matchr

## Status

**Current Status**: Accepted

**Supersedes**: ADR-0002 section on using matchr for OSA

## Context

During Phase 1b implementation of Crucible similarity v2.0.0 fixtures, we discovered a bug in `github.com/antzucaro/matchr` library's OSA (Optimal String Alignment) implementation.

### Bug Discovery

Two OSA test cases failed:

1. `"hello"/"ehllo"`: matchr returned distance=2, fixture expected 1
2. `"algorithm"/"lagorithm"`: matchr returned distance=2, fixture expected 1

Both cases involve **start-of-string adjacent transpositions** (swapping first two characters).

### PyFulmen Validation

PyFulmen validation using `rapidfuzz.distance.OSA` (Python bindings to Rust `strsim-rs 0.11.x`) confirmed:

- ✅ Crucible fixtures are **correct** (distance=1 for both cases)
- ❌ matchr library has a **bug** (returns distance=2)
- ✅ Critical distinction test passes in both (CA/ABC: OSA=3, Unrestricted=2)

**Reference**: `pyfulmen/.plans/memos/20251027-osa-validation-response.md`

### Why This Matters

- **Fixture pass rate**: Blocks 100% pass rate (stuck at 28/30 = 93.3%)
- **Cross-language consistency**: gofulmen diverges from PyFulmen/TSFulmen behavior
- **Production correctness**: Start-of-string transpositions are common typos (e.g., "hte" for "the")
- **Ecosystem alignment**: All foundations should use canonical algorithm implementations

## Decision

**Implement OSA algorithm natively in Go**, replacing matchr.OSA() dependency.

**Implementation approach**:

1. Base on our proven optimized Levenshtein implementation
2. Reference rapidfuzz-cpp OSA implementation as primary source
3. Add transposition detection with three-row DP approach
4. Maintain our performance advantages (1.24-1.76x faster, 3-326x less memory)

**File**: `foundry/similarity/osa.go`

## Rationale

### Why Native Implementation

#### 1. Proven Performance Pattern

Our Phase 1a benchmarks demonstrated native Go advantages over matchr for Levenshtein:

| Metric   | Native vs matchr       | Advantage                            |
| -------- | ---------------------- | ------------------------------------ |
| Speed    | 1.24-1.76x faster      | Consistent across all cases          |
| Memory   | 3-326x less allocation | Especially critical for long strings |
| Accuracy | 100% fixture match     | Both implementations correct         |

**Conclusion**: Native Go outperforms external libraries while maintaining correctness.

#### 2. OSA Algorithm Simplicity

OSA is a **small modification** to Levenshtein:

- Levenshtein: 3 operations (insert, delete, substitute)
- OSA: 4 operations (insert, delete, substitute, **adjacent transpose**)

**Key insight**: We already have optimized Levenshtein. Adding OSA is ~30 lines of code.

#### 3. Foundation for string-metrics-go

This implementation serves as:

- **Validation** of pure Go approach (vs CGO/WASM)
- **Performance baseline** for future benchmarks
- **Reference code** for algorithm porting decisions
- **Proof of concept** for native Go string metrics viability

When string-metrics-go research project proceeds (ADR-0002), this code provides:

- Working algorithm to compare against
- Performance data for decision-making
- Validation that pure Go is competitive

#### 4. Elimination of External Bug

matchr bug demonstrates risk of external dependencies:

- Start-of-string transposition bug existed undetected
- No way to fix without upstream PR or forking
- Blocks our fixture validation and release

**Native implementation**: Full control, no upstream dependency risk.

### Why Not Alternative Libraries

**Evaluated in ADR-0002**:

- `github.com/xrash/smetrics`: Older, less maintained
- `github.com/agext/levenshtein`: Unknown reliability

**Risk**: Spend time evaluating only to discover similar bugs. Better to implement once correctly.

### Why Not Wait for string-metrics-go

**Timeline consideration**:

- string-metrics-go: Post-v0.1.5 research project (contingent on TSFulmen WASM success)
- Native OSA: 1-2 hours implementation

**Benefits of implementing now**:

- ✅ Achieves 100% fixture pass rate for v0.1.5
- ✅ Validates approach for string-metrics-go
- ✅ Removes blocking issue for release
- ✅ Provides performance comparison data

**Waiting would**:

- ❌ Block v0.1.5 at 93.3% fixture pass
- ❌ Delay validation of native Go approach
- ❌ Miss opportunity to use working Levenshtein base

## Algorithm Details

### OSA (Optimal String Alignment)

OSA is a variant of Damerau-Levenshtein distance with a restriction:
**Cannot edit the same substring more than once**.

**Allowed operations**:

1. Insertion: Insert a character
2. Deletion: Delete a character
3. Substitution: Replace a character
4. **Adjacent Transposition**: Swap two adjacent characters

**Key difference from Unrestricted Damerau-Levenshtein**:

- **OSA**: Single-pass DP matrix (restriction automatically enforced)
- **Unrestricted**: Complex Zhao-Sahni algorithm tracking character histories

### Implementation Approach

**Base**: Our existing Levenshtein implementation

```go
// Current Levenshtein uses 2 rows:
prevRow := make([]int, lenA+1)
currRow := make([]int, lenA+1)
```

**OSA modification**: Add third row for transposition detection

```go
// OSA needs 3 rows:
prevPrevRow := make([]int, lenA+1)
prevRow := make([]int, lenA+1)
currRow := make([]int, lenA+1)

// In DP loop, after calculating insert/delete/substitute:
if i > 1 && j > 1 &&
   runesA[i-1] == runesB[j-2] &&
   runesA[i-2] == runesB[j-1] {
    // Adjacent transposition detected
    transpose := prevPrevRow[i-2] + 1
    currRow[i] = min(currRow[i], transpose)
}

// Rotate rows at end of iteration:
prevPrevRow, prevRow, currRow = prevRow, currRow, prevPrevRow
```

**Space complexity**: O(min(m,n)) - three rows instead of two

- Still maintains our memory efficiency advantage
- Much better than unrestricted's O(m×n)

**Time complexity**: O(m×n) - same as Levenshtein

### Reference Implementation

**Primary Source**: rapidfuzz-cpp OSA.hpp

- URL: https://github.com/rapidfuzz/rapidfuzz-cpp/blob/main/rapidfuzz/distance/OSA.hpp
- License: MIT (compatible with gofulmen)
- Rationale: Clean, modern C++, production-proven, part of rapidfuzz ecosystem

**Secondary Source**: strsim-rs OSA

- URL: https://github.com/rapidfuzz/strsim-rs
- Rationale: Canonical Rust implementation (what PyFulmen uses)
- Use: Validation of algorithm correctness

**Tertiary Source**: MWH Python Blog (archived)

- URL: https://web.archive.org/web/20150909134357/http://mwh.geek.nz:80/2009/04/26/python-damerau-levenshtein-distance/
- Rationale: Historical reference with clear explanation
- Note: Less useful for production code (older Python style)

**Note**: GeeksforGeeks article is derivative of Wikipedia; prefer primary sources.

### Code Attribution

Implementation will include proper attribution in comments:

```go
// osaDistance implements Optimal String Alignment distance.
//
// Based on rapidfuzz-cpp OSA implementation:
// https://github.com/rapidfuzz/rapidfuzz-cpp/blob/main/rapidfuzz/distance/OSA.hpp
//
// Copyright notice from rapidfuzz-cpp:
// SPDX-License-Identifier: MIT
// Copyright © 2022-present Max Bachmann
//
// Ported to Go with Unicode (rune) handling and optimizations from
// gofulmen Levenshtein implementation (see similarity.go).
```

## Expected Performance

### Based on Levenshtein Benchmarks (Phase 1a)

Our Levenshtein showed consistent advantages over matchr:

| Test Case                     | Speed Advantage | Memory Advantage |
| ----------------------------- | --------------- | ---------------- |
| Short strings (5-7 chars)     | 1.07-1.16x      | 3-3.5x less      |
| Medium strings (40-50 chars)  | 1.65-1.68x      | 9-16x less       |
| Long strings (100-1000 chars) | 1.67-1.76x      | 31-326x less     |

**OSA expectation**: Similar or better performance

- Same base algorithm (2-row DP → 3-row DP)
- Adds one extra check per cell (minimal overhead)
- Maintains memory efficiency pattern

**Validation**: Benchmark OSA against matchr.OSA() to confirm

## Testing Strategy

### Fixture Validation

All 5 OSA fixtures must pass:

| Test Case            | Input                   | Expected | matchr (Bug) | Native (Target) |
| -------------------- | ----------------------- | -------- | ------------ | --------------- |
| Basic transposition  | "abcd"/"abdc"           | 1        | 1 ✅         | 1 ✅            |
| Start transposition  | "hello"/"ehllo"         | 1        | 2 ❌         | 1 ✅            |
| Prefix transposition | "algorithm"/"lagorithm" | 1        | 2 ❌         | 1 ✅            |
| Non-adjacent         | "abcd"/"acbd"           | 1        | 1 ✅         | 1 ✅            |
| OSA distinction      | "CA"/"ABC"              | 3        | 3 ✅         | 3 ✅            |

**Critical test**: "CA"/"ABC" must remain 3 (proves OSA restriction works)

### Cross-Language Validation

Validate against PyFulmen for additional test cases:

```bash
# gofulmen
go test -v -run TestDistanceWithAlgorithm_DamerauOSA

# pyfulmen (for comparison)
uv run python -c "
from pyfulmen.foundry.similarity import distance
print(distance('hello', 'ehllo', metric='damerau_osa'))  # expect: 1
"
```

### Edge Cases

- Empty strings: `"" vs ""`
- Identical strings: `"test" vs "test"`
- Single character: `"a" vs "b"`
- Unicode: `"café" vs "cfae"` (accent transposition)
- Emoji: `"🔥🎉" vs "🎉🔥"` (emoji transposition)
- CJK: `"日本" vs "本日"` (CJK transposition)

### Performance Benchmark

Compare native OSA against matchr.OSA():

```go
BenchmarkOSA_Native
BenchmarkOSA_Matchr
```

**Target**: Match or exceed Levenshtein performance advantages (1.24-1.76x faster)

## Implementation Plan

### File Structure

**New file**: `foundry/similarity/osa.go`

```go
package similarity

// osaDistance implements OSA algorithm
func osaDistance(a, b string) int { ... }
```

**Update**: `foundry/similarity/distance_v2.go`

```go
// Replace matchr wrapper:
func damerauOSADistance(a, b string) int {
    return osaDistance(a, b)  // Use native implementation
}
```

**Update**: Remove matchr OSA dependency from imports

### Timeline

| Task                        | Effort       | Status      |
| --------------------------- | ------------ | ----------- |
| Study rapidfuzz-cpp OSA.hpp | 30 min       | Pending     |
| Implement osa.go            | 45 min       | Pending     |
| Test all fixtures           | 30 min       | Pending     |
| Documentation updates       | 15 min       | Pending     |
| **Total**                   | **~2 hours** | **Pending** |

### Success Criteria

**Functional**:

- [ ] All 5 OSA fixtures pass (100%)
- [ ] Critical distinction remains correct (CA/ABC = 3)
- [ ] Unicode edge cases handled correctly
- [ ] Score calculations correct from distances

**Non-Functional**:

- [ ] Performance ≥ matchr.OSA() (ideally match Levenshtein advantages)
- [ ] Memory allocation ≤ our Levenshtein pattern
- [ ] Code properly attributed to rapidfuzz-cpp

**Documentation**:

- [ ] This ADR completed
- [ ] Code comments reference rapidfuzz-cpp
- [ ] Discrepancy memo updated with RESOLVED status
- [ ] ADR-0002 updated noting matchr removal

## Consequences

### Positive

- ✅ **100% fixture pass rate** (30/30) for v0.1.5 release
- ✅ **Eliminates external bug** dependency
- ✅ **Maintains performance advantages** (1.24-1.76x faster pattern)
- ✅ **Cross-language consistency** with PyFulmen/TSFulmen
- ✅ **Foundation for string-metrics-go** research project
- ✅ **Full algorithm control** for future optimizations
- ✅ **Validates native Go approach** for similarity algorithms

### Negative

- ⚠️ **Maintenance responsibility** shifts to gofulmen team
- ⚠️ **Implementation effort** (~2 hours one-time cost)
- ⚠️ **Algorithm complexity** slightly higher (3 rows vs 2)

### Neutral

- ℹ️ **Reduces external dependencies** (one less library)
- ℹ️ **Increases codebase size** (~100 lines for osa.go)
- ℹ️ **Sets precedent** for native implementations over external libraries

## Future Considerations

### string-metrics-go Research Project

This implementation provides valuable data for string-metrics-go (ADR-0002):

**Performance comparison**:

- Native Go OSA performance data
- Memory allocation patterns
- Benchmark comparisons vs external libraries

**Implementation validation**:

- Proves pure Go approach is competitive
- Demonstrates algorithm porting feasibility
- Validates cross-language fixture methodology

**Decision input**:

- If native Go performs well → supports pure Go port approach
- If CGO/WASM significantly faster → informs architecture choice
- Provides baseline for "good enough" performance threshold

### Algorithm Evolution

Once string-metrics-go exists:

- **Option A**: Replace with string-metrics-go binding
- **Option B**: Keep native if performance superior
- **Option C**: Maintain both with configuration option

**Decision deferred** until string-metrics-go research concludes.

## Related Work

### ADR-0002: Similarity Algorithm Strategy

ADR-0002 documents hybrid approach:

- Keep optimized Levenshtein (native)
- Use matchr for OSA (external) ← **This decision supersedes this**
- Port Zhao algorithm for unrestricted (custom)

**Update**: ADR-0002 now reflects matchr removal for OSA, replaced with native.

### Discrepancy Memo

`.plans/memos/libraries/20251027-similarity-discrepancies.md` documents:

- matchr bug discovery
- PyFulmen validation
- Fixture analysis

**Status update**: Marked RESOLVED after native OSA implementation complete.

## References

- **Feature Brief**: `.plans/active/v0.1.5/native-osa-implementation.md`
- **Discrepancy Analysis**: `.plans/memos/libraries/20251027-similarity-discrepancies.md`
- **PyFulmen Validation**: `pyfulmen/.plans/memos/20251027-osa-validation-response.md`
- **ADR-0002**: Foundry Similarity Algorithm Implementation Strategy
- **rapidfuzz-cpp OSA**: https://github.com/rapidfuzz/rapidfuzz-cpp/blob/main/rapidfuzz/distance/OSA.hpp
- **strsim-rs**: https://github.com/rapidfuzz/strsim-rs
- **MWH Python Blog**: https://web.archive.org/web/20150909134357/http://mwh.geek.nz:80/2009/04/26/python-damerau-levenshtein-distance/

---

**Decision Outcome**: Implement OSA algorithm natively in Go to replace matchr, achieving 100% fixture pass rate and maintaining our performance advantages. This validates the native approach for future string-metrics-go research while unblocking v0.1.5 release.
