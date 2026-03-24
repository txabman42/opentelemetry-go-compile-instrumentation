---
name: migrate-loongsuite-rule
description: >-
  Migrate an instrumentation rule from loongsuite-go-agent/pkg/rules/ to
  opentelemetry-go-compile-instrumentation/pkg/instrumentation/. Use when
  the user provides a loongsuite rule name (e.g., "gin", "kafka", "gorm")
  and wants to create the equivalent instrumentation with hooks, YAML rules,
  tests, and test app in the target repository.
---

# Migrate Loongsuite Rule

Both repositories must be available in the workspace. Loongsuite is read-only (source); all writes go to the otel repo.

**Terminology**: rule = YAML declaration, hook = Go Before/After function, instrumentation = complete package (hooks + rules + tests + go.mod).

## Step 0: Read Source and Classify

**1. Check if already migrated.** List `pkg/instrumentation/` in the otel repo. If a directory matching `<name>` exists, STOP and tell the user.

**2. Read all loongsuite source files:**
- JSON rule(s): `loongsuite-go-agent/tool/data/rules/*.json` — search for entries where `Path` contains `<name>` or `ImportPath` matches the target library
- Hook code: all `.go` files under `loongsuite-go-agent/pkg/rules/<name>/`

**3. Detect multi-version support.** List subdirectories of `loongsuite-go-agent/test/<name>/`. If it contains version subdirectories (e.g. `v2.13.0/`, `v2.42.0/`), record ALL of them as the **version list**. Many loongsuite rules test against multiple library versions — all must be migrated. Read the `go.mod` in each version subdirectory to identify the pinned library version. If there is only one version subdirectory or no subdirectories (flat layout), record a single version.

**4. Classify the tier** by reading the hook code:

| Signal in hook code | Tier |
|---------------------|------|
| JSON has `StructType` entries | **Complex-DB** |
| Imports `instrumenter.Builder` AND carrier for kafka/amqp/rocketmq headers | **Messaging** |
| Imports `instrumenter.Builder` (HTTP/RPC/DB pattern) | **Standard** |
| Only calls `LocalRootSpanFromGLS()`, no span creation | **Simple** |

Also identify: target package, number of hooks, which functions/methods are hooked, whether propagation carriers exist.

## Step 1: Create Package Structure

Look at a similar existing instrumentation for layout guidance:
- Simple tier: no close equivalent exists — create a minimal single-file layout
- Standard (client/server): `pkg/instrumentation/nethttp/{client,server}/` or `pkg/instrumentation/grpc/{client,server}/`
- Standard (single package): `pkg/instrumentation/redis/v9/` or `pkg/instrumentation/databasesql/`
- Complex-DB: `pkg/instrumentation/databasesql/` (canonical reference — read it)
- Messaging: no existing example — use Standard single-package layout with `semconv/` subpackage

Create the directory and files. For `go.mod`, read [reference/translation-patterns.md](reference/translation-patterns.md) section 2 for the template and replace directive depths.

After creating: `make go-mod-tidy && make crosslink`.

## Step 2: Translate Rules to YAML

Read [reference/translation-patterns.md](reference/translation-patterns.md) section 1 for the field mapping table (JSON field → YAML field).

Create one YAML block per JSON rule entry. For struct injection entries, create `struct` + `new_field` YAML blocks. Version range syntax: drop brackets, ensure `v` prefix — e.g. `[1.3.0,1.7.4)` → `v1.3.0,v1.7.4`.

## Step 3: Translate Hook Code

Read [reference/translation-patterns.md](reference/translation-patterns.md) for the boilerplate template and translation patterns.

**All tiers:**
- Apache 2.0 license header on every `.go` file
- `//go:linkname` unexported functions → exported `BeforeX`/`AfterX` (the tool generates linkname automatically)
- `api.CallContext` → `inst.HookContext` (1:1 rename, same methods)

**Simple tier** — GLS span-rename only:
- No `initInstrumentation()`, no tracer, no span creation
- Translate `trace.LocalRootSpanFromGLS()` → read GLS via `runtime.GetTraceContextFromGLS()`, cast to `context.Context`, call `trace.SpanFromContext(ctx)` (see translation pattern 7)
- Enable check: `shared.Instrumented("KEY")`

**Standard tier** — Instrumenter flatten:
- Full boilerplate: enabler + `initInstrumentation()` + tracer + propagator
- Flatten `instrumenter.Builder` + `AttrsGetter` → direct `tracer.Start()` / `span.SetAttributes()` / `span.End()` (see patterns 5 and 6)
- Propagation: keep carrier structs, call `propagator.Inject`/`Extract` directly (see pattern 9)
- Semconv keys: use `go.opentelemetry.io/otel/semconv/v1.37.0`
- Read an existing similar hook as a style reference before writing

**Complex-DB tier:**
- Add `struct` rules in YAML for every injected field
- Hook the constructor to store connection metadata into injected fields
- Hook operations to read injected fields and build DB semconv spans
- Reference `pkg/instrumentation/databasesql/client.go` as the canonical pattern
- Hooks that read injected fields cannot be unit tested — document this limitation inline

**Messaging tier:**
- SpanKind: Producer for publish, Consumer for receive/process
- Producer: `propagator.Inject(ctx, carrier)` after `tracer.Start`
- Consumer: `ctx = propagator.Extract(parentCtx, carrier)` before `tracer.Start`
- Keep carrier structs from loongsuite unchanged
- Semconv: `MessagingSystemKey`, `MessagingDestinationNameKey`, `MessagingOperationTypeKey`
- Span name: `"<topic> publish"` / `"<topic> process"` etc.

Read [reference/tier-examples.md](reference/tier-examples.md) for annotated before/after snippets for your tier.

## Step 4: Identify Test Scenarios from Loongsuite

Before writing any tests, read the loongsuite test files to discover what scenarios were considered important:

1. **`loongsuite-go-agent/test/<name>_tests.go`** — lists every test case registered via `NewGeneralTestCase`. The number and names of cases tell you how many distinct scenarios exist (e.g. for nethttp: basic, HTTP/2, HTTPS, metrics).
2. **`loongsuite-go-agent/test/<name>/<version>/test_*.go`** — each file is one scenario. Read the `verifier.Verify*Attributes` calls to see exactly which semconv attributes were validated and which parent-child span relationships were asserted.

**Multi-version rules:** If Step 0 identified multiple versions, read the test files from **every** version subdirectory (e.g. `test/<name>/v2.13.0/test_*.go` AND `test/<name>/v2.42.0/test_*.go`). Compare them — often the test code is identical or nearly identical across versions, but sometimes newer versions add scenarios or change APIs. Note any differences; each version needs its own test app in Step 6.

Use these scenarios as the **test coverage target** for both unit and integration tests. Do not limit yourself to a single happy-path test if loongsuite validated multiple distinct protocol or transport variants.

## Step 5: Unit Tests

Read [reference/testing-checklist.md](reference/testing-checklist.md) for required test cases and the mock setup pattern.

Look at an existing test file for style: `pkg/instrumentation/nethttp/client/client_hook_test.go` (Standard) or `pkg/instrumentation/redis/v9/hook_test.go` (simpler Standard).

Use `insttest.NewMockHookContext(...)` from `pkg/inst/insttest` and `tracetest.SpanRecorder` for span capture. Write table-driven tests.

## Step 6: Test App

**Single-version rules:** Create `test/apps/<name>/main.go` — standalone Go module (`go.mod` + `go.sum`). It must accept flags (`-addr`, `-op`, `-scheme`, etc.) and perform the operations identified in Step 4.

**Multi-version rules:** Create one subdirectory per version under `test/apps/<name>/`:

```
test/apps/<name>/
├── <version-a>/       # e.g. v2.13.0
│   ├── go.mod         # pins <library> <version-a>
│   ├── go.sum
│   └── main.go
└── <version-b>/       # e.g. v2.42.0
    ├── go.mod         # pins <library> <version-b>
    ├── go.sum
    └── main.go
```

Each version subdirectory is a standalone Go module. The `go.mod` in each pins the corresponding library version (read from the loongsuite `test/<name>/<version>/go.mod`). The `main.go` is adapted from the corresponding loongsuite `test/<name>/<version>/test_*.go`. Often `main.go` is identical across versions — if so, write it once and copy, adjusting only if the API changed between versions.

The `TestFixture.resolveAppPath` already supports path-like names via `filepath.Join`, so `f.BuildAndRun("<name>/<version>")` resolves to `test/apps/<name>/<version>/`. The `make tidy/test-apps` target uses `find test/apps -name "go.mod"` which automatically discovers nested modules.

Run `make tidy/test-apps` after creating all version directories.

Look at `test/apps/redisclient/` or `test/apps/dbclient/` as style references for the `main.go` content.

## Step 7: Integration Test

Create `test/integration/<name>_test.go` with `//go:build integration`.

**Docker-based tests (testcontainers):** Integration tests run in CI on **macOS, Windows, and Ubuntu** (see `.github/workflows/test-integration.yaml`). Only Ubuntu runners have Docker. When a test requires Docker (e.g. Elasticsearch, Kafka, ClickHouse), use `testcontainers-go` with `testcontainers.SkipIfProviderIsNotHealthy(t)` as the first line in the test function. This gracefully skips the test on runners without Docker instead of panicking.

```go
func TestElasticsearch(t *testing.T) {
    testcontainers.SkipIfProviderIsNotHealthy(t)  // skips on macOS/Windows, runs on Ubuntu
    // ... testcontainers setup ...
}
```

For services that have lightweight in-process alternatives, prefer those (e.g. `miniredis` for Redis, `httptest.Server` for simple HTTP, fake SQL drivers for `database/sql`).

Read an existing test for the full pattern: `test/integration/redis_client_test.go` (simple) or `test/integration/db_client_test.go` (Complex-DB). Use `testutil.NewTestFixture`, `f.BuildAndRun`, and `testutil.Require*Semconv` helpers. Add a new `RequireXxxSemconv` helper in `test/testutil/semconv.go` if no existing one fits.

**CRITICAL — test module API divergence:** The `test/testutil/semconv.go` helpers have DIFFERENT signatures from the identically-named file under `pkg/instrumentation/`. In particular, `RequireDBClientSemconv` in the `test` module does NOT accept a variadic `opts ...DBClientSemconvOptions` parameter. Always **read `test/testutil/semconv.go`** before writing a new `Require*Semconv` helper to match its actual API. When adding a system-specific helper (e.g. `RequireElasticsearchClientSemconv`), assert `db.system.name` directly via `RequireAttribute` instead of relying on options structs that don't exist in the test module.

Write one sub-test (`t.Run`) per scenario identified in Step 4. For example, if loongsuite tests HTTP/1.1, HTTP/2, and HTTPS, add three sub-tests. The loongsuite `verifier.Verify*Attributes` calls tell you exactly which attributes to assert in each scenario.

**Hook coverage rule:** Every hook function declared in the YAML (`before`/`after`) must be exercised by at least one integration sub-test. A hook is only exercised if the test app actually calls the hooked method — for example, `BeforeHTML` is only triggered when the handler calls `c.HTML()`, not `c.String()` or `c.JSON()`. Check the hooked function names and ensure the test app has a route that calls each one directly. Use `engine.SetHTMLTemplate(template.Must(...))` to register inline templates if file templates would add unnecessary complexity.

**Multi-version rules:** Wrap all scenario sub-tests in a per-version `t.Run` block. Use the path-like app name to target each version's test app:

```go
func TestClickhouse(t *testing.T) {
    versions := []string{"v2.13.0", "v2.42.0"}
    for _, ver := range versions {
        t.Run(ver, func(t *testing.T) {
            f := testutil.NewTestFixture(t)
            dep := startClickhouse(t)
            f.BuildAndRun("clickhousev2/"+ver, "-addr="+dep.Addr())
            // assert spans...
        })
    }
}
```

This ensures every library version covered by loongsuite is validated in the otel repo. The `resolveAppPath` in `TestFixture` supports path-like names (`"<name>/<version>"` resolves to `test/apps/<name>/<version>/`).

For Messaging: write an E2E test (`test/e2e/<name>_test.go`) if context propagation across producer/consumer must be validated.

## Step 8: Validate — Do Not Skip

Run every command. Fix failures before considering the migration done:

```bash
go test -C pkg/instrumentation/<name> ./...  # unit tests (run directly — see pitfall below)
make lint/license-header/fix                  # add missing license headers
make go-mod-tidy && make crosslink            # module consistency
make test-integration                         # end-to-end (includes build internally)
make lint/go                                  # no lint suppression allowed
```

> `make test-integration` already depends on `make build` and `make build-demo` — do not run `make build` separately first.

<<<<<<< Updated upstream
=======
> **`make test-integration` is a BLOCKING gate.** Do NOT mark the migration as complete if it fails. Read the failure output carefully. Fix the hook code and re-run until all tests pass.

>>>>>>> Stashed changes
**Final checklist:**
- [ ] Apache 2.0 header on every `.go` file
- [ ] Every hook function exported and named in YAML `before`/`after`
- [ ] Every hook has ≥1 unit test asserting span attributes
<<<<<<< Updated upstream
- [ ] Integration test asserts semconv attributes
=======
- [ ] Every hook declared in YAML has ≥1 integration test sub-case that exercises it directly (e.g. a route using `c.HTML()` to trigger `BeforeHTML`, not just routes using `c.String()`/`c.JSON()`)
- [ ] Integration test asserts semconv attributes
- [ ] `make test-integration` passes (all tests green)
>>>>>>> Stashed changes
- [ ] Multi-version: test app exists under `test/apps/<name>/<version>/` for EVERY version from Step 0
- [ ] Multi-version: integration test has a `t.Run` sub-test for EVERY version
- [ ] `make all` passes clean
