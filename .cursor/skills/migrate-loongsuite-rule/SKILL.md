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

**3. Classify the tier** by reading the hook code:

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
2. **`loongsuite-go-agent/test/<name>/test_*.go`** — each file is one scenario. Read the `verifier.Verify*Attributes` calls to see exactly which semconv attributes were validated and which parent-child span relationships were asserted.

Use these scenarios as the **test coverage target** for both unit and integration tests. Do not limit yourself to a single happy-path test if loongsuite validated multiple distinct protocol or transport variants.

## Step 5: Unit Tests

Read [reference/testing-checklist.md](reference/testing-checklist.md) for required test cases and the mock setup pattern.

Look at an existing test file for style: `pkg/instrumentation/nethttp/client/client_hook_test.go` (Standard) or `pkg/instrumentation/redis/v9/hook_test.go` (simpler Standard).

Use `insttest.NewMockHookContext(...)` from `pkg/inst/insttest` and `tracetest.SpanRecorder` for span capture. Write table-driven tests.

## Step 6: Test App

Create `test/apps/<name>/main.go` — standalone Go module (`go.mod` + `go.sum`). It must accept flags (`-addr`, `-op`, `-scheme`, etc.) and perform the operations identified in Step 4. Run `make tidy/test-apps` after.

Look at `test/apps/redisclient/` or `test/apps/dbclient/` as style references.

## Step 7: Integration Test

Create `test/integration/<name>_test.go` with `//go:build integration`.

Read an existing test for the full pattern: `test/integration/redis_client_test.go` (simple) or `test/integration/db_client_test.go` (Complex-DB). Use `testutil.NewTestFixture`, `f.BuildAndRun`, and `testutil.Require*Semconv` helpers. Add a new `RequireXxxSemconv` helper in `test/testutil/semconv.go` if no existing one fits.

Write one sub-test (`t.Run`) per scenario identified in Step 4. For example, if loongsuite tests HTTP/1.1, HTTP/2, and HTTPS, add three sub-tests. The loongsuite `verifier.Verify*Attributes` calls tell you exactly which attributes to assert in each scenario.

For Messaging: write an E2E test (`test/e2e/<name>_test.go`) if context propagation across producer/consumer must be validated.

## Step 8: Validate — Do Not Skip

Run every command. Fix failures before considering the migration done:

```bash
go test -C pkg/instrumentation/<name> ./...  # unit tests
make format/license                           # license headers
make go-mod-tidy && make crosslink            # module consistency
make build                                    # packages hooks into binary
make test-integration                         # end-to-end validation
make lint/go                                  # no lint suppression allowed
```

**Final checklist:**
- [ ] Apache 2.0 header on every `.go` file
- [ ] Every hook function exported and named in YAML `before`/`after`
- [ ] Every hook has ≥1 unit test asserting span attributes
- [ ] Integration test asserts semconv attributes
- [ ] `make all` passes clean
