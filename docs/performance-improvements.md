# Performance Improvements

[![Audience: Developers](https://img.shields.io/badge/Audience-Developers-blue)](../README.md)

This document outlines performance improvements for the OpenTelemetry Go Compile-Time Instrumentation Tool, focusing on optimizations for large projects.

## Table of Contents

- [Overview](#overview)
- [Priority 1: Quick Wins](#priority-1-quick-wins)
- [Priority 2: Caching Layer](#priority-2-caching-layer)
- [Priority 3: Advanced Optimizations](#priority-3-advanced-optimizations)
- [Implementation Status](#implementation-status)
- [Expected Performance Gains](#expected-performance-gains)

## Overview

The tool currently has several performance bottlenecks that become significant with large projects:

1. **Forced full rebuild** (`-a` flag) - Rebuilds all packages even if unchanged
2. **Setup runs every time** - No caching mechanism
3. **File I/O on every toolexec invocation** - Each package reads matched rules from disk
4. **No caching of instrumented files** - Re-instruments unchanged packages
5. **Sequential processing** - Some operations could be parallelized
6. **Excessive logging** - Per-package logging adds overhead

## Priority 1: Quick Wins

High impact, low effort improvements that can be implemented quickly.

### 1.1 Setup Caching ✅ IMPLEMENTED

**Status**: ✅ Completed

**Implementation**: See `tool/internal/setup/cache.go`

**What it does**:
- Caches setup completion state using a marker file
- Tracks SHA256 hashes of dependency files (`go.mod`, `go.sum`, `go.work`, `go.work.sum`)
- Skips setup phase if dependencies haven't changed

**Benefits**:
- 80-95% faster on subsequent builds
- Eliminates expensive dependency analysis when not needed

**How it works**:
1. After successful setup, saves `.otel-build/setup.marker` with:
   - Completion timestamp
   - File hashes of dependency files
2. On next run, checks:
   - Marker file exists
   - Setup artifacts exist (`matched.json`, extracted modules)
   - Dependency file hashes match
3. If all checks pass, skips setup entirely

### 1.2 Incremental Build Support ✅ IMPLEMENTED

**Status**: ✅ Completed

**Implementation**: See `tool/internal/setup/cache.go` and `tool/internal/setup/setup.go`

**What it does**:
- Adds `--force-rebuild` flag for explicit control
- Only adds `-a` flag if:
  - Setup phase ran (dependencies changed)
  - User explicitly requests it via `--force-rebuild` flag
  - Cache is invalid
- Leverages Go's build cache for incremental builds by default

**Benefits**:
- 50-90% faster for unchanged code
- Leverages Go's built-in build cache
- Reduces compilation time significantly

**How it works**:
1. `ShouldForceRebuild()` checks:
   - User's `--force-rebuild` flag
   - Setup marker to see if setup ran
   - Cache validity (file hashes, artifacts existence)
2. If all checks pass, skips `-a` flag
3. Go's build cache handles incremental compilation

**Usage**:
```bash
# Default: incremental build (if cache valid)
otel go build

# Force full rebuild
otel go --force-rebuild build
```

### 1.3 Rule Indexing

**Status**: ⏳ Pending

**Current Issue**: Each package does linear search through all rules:
```go
// tool/internal/instrument/match.go:47
for _, rset := range allSet {
    if rset.ModulePath == importPath {
        return rset
    }
}
```

**Proposed Solution**:
```go
// In setup phase, create index
type RuleIndex struct {
    ByPackage map[string]*rule.InstRuleSet
}

// In toolexec phase, use index
func (ip *InstrumentPhase) match(index *RuleIndex, args []string) *rule.InstRuleSet {
    importPath := util.FindFlagValue(args, "-p")
    return index.ByPackage[importPath]
}
```

**Implementation Steps**:
1. Create `RuleIndex` struct in setup phase
2. Build index map during `store()` operation
3. Save index alongside matched rules
4. Load index in toolexec phase
5. Use O(1) lookup instead of O(n) search

**Benefits**:
- 10-30% faster per-package processing
- Scales better with many rules
- Reduces CPU usage

**Files to Modify**:
- `tool/internal/setup/store.go` - Create and save index
- `tool/internal/instrument/match.go` - Use index for lookup

### 1.4 Reduce Logging Overhead

**Status**: ⏳ Pending

**Current Issue**: Every package logs multiple messages:
- `cmd_toolexec started/completed`
- `Toolexec started/completed`
- `interceptCompile started/completed`
- `load started/completed`
- `instrument started/completed`
- `parseFile started/completed`
- `writeInstrumented started/completed`

**Proposed Solution**:
1. Use debug level for per-package logs
2. Only log at info level for:
   - Setup phase start/end
   - Build phase start/end
   - Significant events (errors, warnings)
3. Add log sampling for high-frequency events
4. Batch log writes

**Implementation Steps**:
1. Change per-package logs to debug level
2. Add structured logging with sampling
3. Batch log writes in toolexec phase

**Benefits**:
- 5-15% faster overall
- Reduced I/O overhead
- Cleaner logs for users

**Files to Modify**:
- `tool/internal/instrument/toolexec.go`
- `tool/internal/instrument/instrument.go`
- `tool/cmd/cmd_toolexec.go`

## Priority 2: Caching Layer

Medium effort, high impact improvements requiring more implementation work.

### 2.1 File-Based Cache for Instrumented Files

**Status**: ⏳ Pending

**Current Issue**: Every package re-instruments even if source files haven't changed.

**Proposed Solution**:
```go
type InstrumentCache struct {
    SourceHash    string // Hash of source files + rules
    Instrumented  string // Path to cached instrumented file
    Timestamp     time.Time
}

func (ip *InstrumentPhase) getCachedInstrumented(rset *rule.InstRuleSet) (string, bool) {
    cacheKey := computeCacheKey(rset)
    cacheFile := getCachePath(cacheKey)
    
    if !util.PathExists(cacheFile) {
        return "", false
    }
    
    // Verify cache is still valid
    if !isCacheValid(cacheFile, rset) {
        return "", false
    }
    
    return cacheFile, true
}
```

**Implementation Steps**:
1. Compute cache key from:
   - Source file hashes
   - Matched rules
   - Tool version
2. Store instrumented files in `.otel-build/cache/`
3. Check cache before instrumentation
4. Invalidate cache when:
   - Source files change
   - Rules change
   - Tool version changes

**Benefits**:
- 60-80% faster for unchanged packages
- Significant time savings on large projects
- Reduces AST parsing and manipulation

**Files to Create**:
- `tool/internal/instrument/cache.go`

**Files to Modify**:
- `tool/internal/instrument/instrument.go` - Check cache before instrumenting

### 2.2 Shared Memory for Matched Rules

**Status**: ⏳ Pending

**Current Issue**: Each toolexec invocation reads and parses JSON file:
```go
// tool/internal/instrument/match.go:27
content, err := os.ReadFile(f)
// ... JSON unmarshal
```

**Proposed Solution**:
1. Use memory-mapped file for matched rules
2. Or use a daemon process that keeps rules in memory
3. Or use a shared cache file with locking

**Implementation Options**:

**Option A: Memory-Mapped File**
```go
import "golang.org/x/sys/unix"

func loadRulesMMAP() ([]*rule.InstRuleSet, error) {
    fd, err := unix.Open(matchedFile, unix.O_RDONLY, 0)
    // ... mmap implementation
}
```

**Option B: Daemon Process**
- Start a daemon that loads rules once
- toolexec processes communicate via IPC
- More complex but most efficient

**Option C: Shared Cache with Locking**
- Use file locking for concurrent access
- Cache parsed rules in memory per process
- Simpler but less efficient than daemon

**Benefits**:
- 20-40% faster per-package processing
- Reduces file I/O significantly
- Better scalability

**Files to Create**:
- `tool/internal/instrument/shared_cache.go`

**Files to Modify**:
- `tool/internal/instrument/match.go` - Use shared cache

### 2.3 AST Parsing Cache

**Status**: ⏳ Pending

**Current Issue**: Files are parsed every time even if unchanged.

**Proposed Solution**:
```go
type ASTCache struct {
    FileHash   string
    ParsedAST  *dst.File
    Timestamp  time.Time
}

func (ip *InstrumentPhase) getCachedAST(file string) (*dst.File, bool) {
    fileHash := computeFileHash(file)
    cacheKey := fmt.Sprintf("%s:%s", file, fileHash)
    
    if cached, exists := astCache[cacheKey]; exists {
        return cached.ParsedAST, true
    }
    return nil, false
}
```

**Implementation Steps**:
1. Cache parsed ASTs in memory
2. Key by file path + file hash
3. Invalidate on file change
4. Limit cache size to prevent memory issues

**Benefits**:
- 15-25% faster for files with multiple rules
- Reduces CPU usage
- Better for large files

**Files to Modify**:
- `tool/internal/instrument/instrument.go` - Cache ASTs

## Priority 3: Advanced Optimizations

Higher effort optimizations requiring architectural changes.

### 3.1 Parallel File Operations

**Status**: ⏳ Pending

**Current Issue**: Some operations are sequential that could be parallel.

**Areas for Parallelization**:
1. `addDeps()` - Can process multiple packages concurrently
2. `extract()` - Can extract multiple modules in parallel
3. File I/O operations - Batch and parallelize

**Proposed Solution**:
```go
func (sp *SetupPhase) addDepsParallel(matched []*rule.InstRuleSet) error {
    g, ctx := errgroup.WithContext(ctx)
    g.SetLimit(runtime.NumCPU())
    
    for _, m := range matched {
        m := m // capture loop variable
        g.Go(func() error {
            return sp.addDepsForPackage(m)
        })
    }
    
    return g.Wait()
}
```

**Benefits**:
- 30-50% faster setup phase
- Better CPU utilization
- Scales with number of cores

**Files to Modify**:
- `tool/internal/setup/add.go`
- `tool/internal/setup/extract.go`

### 3.2 Lazy AST Parsing

**Status**: ⏳ Pending

**Current Issue**: Files are parsed even if no rules match.

**Proposed Solution**:
1. Check if file needs instrumentation before parsing
2. Only parse files that have matching rules
3. Skip parsing entirely for non-instrumented packages

**Benefits**:
- 10-20% faster overall
- Reduces memory usage
- Faster for projects with few instrumented packages

**Files to Modify**:
- `tool/internal/instrument/instrument.go`

### 3.3 Incremental Dependency Analysis

**Status**: ⏳ Pending

**Current Issue**: Full dependency analysis runs every time setup is needed.

**Proposed Solution**:
1. Cache dependency graph
2. Only re-analyze changed modules
3. Use Go's build cache information

**Benefits**:
- 40-60% faster setup phase
- Better for large dependency trees

**Files to Modify**:
- `tool/internal/setup/find.go`

## Implementation Status

| Improvement | Priority | Status | Estimated Effort | Impact |
|------------|----------|--------|------------------|--------|
| Setup Caching | 1 | ✅ Done | Low | High |
| Incremental Build | 1 | ✅ Done | Low | High |
| Rule Indexing | 1 | ⏳ Pending | Low | Medium |
| Reduce Logging | 1 | ⏳ Pending | Low | Low |
| File Cache | 2 | ⏳ Pending | Medium | High |
| Shared Memory Rules | 2 | ⏳ Pending | Medium | Medium |
| AST Cache | 2 | ⏳ Pending | Medium | Medium |
| Parallel Operations | 3 | ⏳ Pending | High | Medium |
| Lazy Parsing | 3 | ⏳ Pending | Medium | Low |
| Incremental Deps | 3 | ⏳ Pending | High | Medium |

## Expected Performance Gains

### Current Performance (Baseline)
For a project with **1000 packages**:
- First build: ~5-10 minutes
- Subsequent builds: ~5-10 minutes (no caching)

### With Priority 1 Improvements
- First build: ~5-10 minutes (unchanged)
- Subsequent builds: ~30 seconds - 2 minutes
  - Setup caching: 80-95% faster
  - Incremental build: 50-90% faster
  - Rule indexing: 10-30% faster

### With Priority 1 + 2 Improvements
- First build: ~3-6 minutes
- Subsequent builds: ~10-30 seconds
  - File caching: 60-80% faster
  - Shared memory: 20-40% faster
  - AST caching: 15-25% faster

### With All Improvements
- First build: ~2-4 minutes
- Subsequent builds: ~5-15 seconds
  - Parallel operations: 30-50% faster
  - Lazy parsing: 10-20% faster
  - Incremental deps: 40-60% faster

## Implementation Roadmap

### Phase 1: Quick Wins (1-2 weeks)
1. ✅ Setup caching
2. ✅ Incremental build support
3. Rule indexing
4. Reduce logging overhead

### Phase 2: Caching Layer (2-4 weeks)
1. File-based cache for instrumented files
2. Shared memory for matched rules
3. AST parsing cache

### Phase 3: Advanced Optimizations (4-6 weeks)
1. Parallel file operations
2. Lazy AST parsing
3. Incremental dependency analysis

## Testing Recommendations

For each improvement:
1. **Benchmark before/after** - Measure actual performance gains
2. **Test with large projects** - Verify improvements scale
3. **Test cache invalidation** - Ensure correctness
4. **Test edge cases** - Handle failures gracefully

## Notes

- All improvements maintain backward compatibility
- Cache invalidation is critical for correctness
- Performance gains are estimates and may vary by project
- Some improvements may require Go version constraints

