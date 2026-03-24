# Testing Checklist

## Unit Test Requirements

Every hook file needs a `_hook_test.go`. Look at `pkg/instrumentation/nethttp/client/client_hook_test.go` for the full pattern.

**Required test cases:**
1. Normal call → span created with correct name, kind, and semconv attributes on the finished span
2. `OTEL_GO_DISABLED_INSTRUMENTATIONS=<KEY>` → no span created
3. `OTEL_GO_ENABLED_INSTRUMENTATIONS=<KEY>` → span is created
4. Before hook stores data that After hook retrieves (span lifecycle works)
5. After hook with error → `span.SetStatus(codes.Error, ...)` and `span.RecordError`

**Setup pattern:**
```go
func setupTestTracer(t *testing.T) (*tracetest.SpanRecorder, *sdktrace.TracerProvider) {
    t.Helper()
    sr := tracetest.NewSpanRecorder()
    tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))
    otel.SetTracerProvider(tp)
    otel.SetTextMapPropagator(propagation.TraceContext{})
    initOnce = sync.Once{} // reset between tests
    return sr, tp
}
```

**Mock context:**
```go
import "github.com/open-telemetry/opentelemetry-go-compile-instrumentation/pkg/inst/insttest"
ictx := insttest.NewMockHookContext(param0, param1)
```

**Struct injection limitation:** hooks that read injected fields (e.g. `db.Endpoint`) will fail to compile in unit tests — those fields don't exist at normal build time. Skip unit tests for that code path; rely on integration tests. Add this comment:
```go
// Integration tested only: accesses compile-time injected struct fields.
```

## Integration Test Requirements

Create `test/integration/<name>_test.go` with `//go:build integration`.

<<<<<<< Updated upstream
=======
**Docker-based tests (testcontainers):** Integration tests run on macOS, Windows, and Ubuntu in CI (see `.github/workflows/test-integration.yaml`). Only Ubuntu has Docker. When a test needs Docker, use `testcontainers-go` but ALWAYS call `testcontainers.SkipIfProviderIsNotHealthy(t)` as the first line. This skips the test gracefully on runners without Docker. For services with lightweight in-process alternatives, prefer those (`miniredis`, `httptest.Server`, fake DB drivers).

>>>>>>> Stashed changes
**Read loongsuite test scenarios first** — `loongsuite-go-agent/test/<name>_tests.go` lists all scenarios (basic, HTTP/2, HTTPS, metrics, etc.) and `loongsuite-go-agent/test/<name>/<version>/test_*.go` shows the exact attributes asserted for each. Each registered `NewGeneralTestCase` should map to a `t.Run` sub-test here. The `verifier.Verify*Attributes` calls in the loongsuite test apps are the ground truth for which semconv attributes are expected per scenario.

Read an existing test for style: `test/integration/redis_client_test.go` (simple) or `test/integration/db_client_test.go` (Complex-DB).

**Single-version pattern:**
```go
func Test<Name>(t *testing.T) {
    f := testutil.NewTestFixture(t)
    dep := startInProcessDependency(t)  // httptest.Server, miniredis, testdb driver, etc.
    f.BuildAndRun("<name>", "-addr="+dep.Addr())
    span := f.RequireSingleSpan()
    testutil.RequireAttribute(t, span, "attr.key", expectedValue)
    // or use a Require*Semconv helper
}
```

**Multi-version pattern** (when loongsuite has multiple version subdirectories):
```go
func Test<Name>(t *testing.T) {
    versions := []string{"v2.13.0", "v2.42.0"}
    for _, ver := range versions {
        t.Run(ver, func(t *testing.T) {
            f := testutil.NewTestFixture(t)
            dep := startInProcessDependency(t)
            f.BuildAndRun("<name>/"+ver, "-addr="+dep.Addr())
            span := f.RequireSingleSpan()
            testutil.RequireAttribute(t, span, "attr.key", expectedValue)
        })
    }
}
```

**Available semconv helpers** (`test/testutil/semconv.go`):
- `RequireHTTPClientSemconv` / `RequireHTTPServerSemconv`
- `RequireGRPCClientSemconv` / `RequireGRPCServerSemconv`
- `RequireDBClientSemconv`
- `RequireRedisClientSemconv`
- `RequireMessagingSemconv`

If none fits, add `RequireXxxSemconv(t, span, ...)` to `test/testutil/semconv.go`. New helpers must only assert attributes that are defined in the [OTel Semantic Conventions](https://opentelemetry.io/docs/specs/semconv/) for that signal (e.g. [Messaging spans](https://opentelemetry.io/docs/specs/semconv/messaging/messaging-spans/), [Database spans](https://opentelemetry.io/docs/specs/semconv/database/database-spans/), [RPC spans](https://opentelemetry.io/docs/specs/semconv/rpc/rpc-spans/)). Do not invent custom attribute names — check the spec first.

<<<<<<< Updated upstream
=======
**IMPORTANT — `test/testutil/semconv.go` vs `pkg/` copy:** The `test` module has its OWN `semconv.go` with DIFFERENT function signatures. In particular, `RequireDBClientSemconv` in `test/testutil/` does NOT accept variadic `opts ...DBClientSemconvOptions`. Always READ `test/testutil/semconv.go` before writing helpers. For system-specific attributes, call `RequireAttribute` directly (e.g. `RequireAttribute(t, span, "db.system.name", "elasticsearch")`) instead of relying on options structs from the `pkg/` copy.

>>>>>>> Stashed changes
## E2E Test

Write `test/e2e/<name>_test.go` with `//go:build e2e` **only** when multiple instrumented processes interact and trace context propagation must be validated (e.g. messaging producer → consumer). Read `test/e2e/http_test.go` for the pattern.

## Quality Gates

All must pass. Fix failures before marking migration complete.

```bash
go test -C pkg/instrumentation/<name> ./...  # unit tests (run directly — see note below)
make lint/license-header/fix                  # add missing license headers
make go-mod-tidy && make crosslink            # module consistency
make test-integration                         # end-to-end validation (includes build step)
make lint/go                                  # linter — no suppressions
```

> `make test-integration` depends on `make build` and `make build-demo` internally — do not run `make build` separately before it; just call `make test-integration` directly.

> The convenience alias `make format` runs `format/go`, `format/yaml`, and `lint/license-header/fix` together and can replace the standalone license step above.

## Common Pitfalls

- **`test/go.mod` not synced (CI-breaking)**: integration tests live in the `test/` Go module, separate from `pkg/`. When a new import is added to a file under `test/integration/`, you MUST run `cd test && go mod tidy` so the dependency is added to `test/go.mod`. The `make go-mod-tidy` target includes the `test/` module, but if the file was created AFTER the tidy step ran, the dep will be missing. A missing dep in `test/go.mod` causes `gotestfmt` panics ("Empty package name") in CI even for unrelated test targets (e.g. `test-e2e`). Always verify with `cd test && go vet -tags integration ./integration/... && go vet -tags e2e ./e2e/...` after making changes.
- **Missing `SkipIfProviderIsNotHealthy` with testcontainers (CI-breaking)**: integration tests run on macOS/Windows runners without Docker. Calling `testcontainers.GenericContainer()` without a prior `testcontainers.SkipIfProviderIsNotHealthy(t)` panics with "rootless Docker not found". Always add `testcontainers.SkipIfProviderIsNotHealthy(t)` as the FIRST line in any test function that uses testcontainers.
- **`server.address` includes port (Go `url.Host` trap)**: Go's `url.URL.Host` returns `host:port` (e.g. `localhost:9200`). OTel semconv requires `server.address` to be the hostname ONLY. Always use `url.URL.Hostname()` for `semconv.ServerAddressKey`. Use `url.URL.Port()` (convert to int) for `semconv.ServerPortKey`. This mismatch silently passes unit tests (where client is nil) but fails integration tests with real containers using random ports.
- **`RequireDBClientSemconv` always asserts `db.namespace`**: passing `""` as `dbNamespace` still asserts the attribute exists with an empty value, which fails if the instrumentation doesn't set `db.namespace` (e.g. Elasticsearch). `RequireDBClientSemconv` guards `db.namespace` with `if dbNamespace != ""` — pass `""` to skip the assertion. When writing a new `Require*Semconv` helper, do NOT call `RequireDBClientSemconv` with a dummy namespace if the library doesn't produce one; call `RequireAttribute` directly for only the attributes the hook actually sets.
- **Loongsuite `AttrsGetter` returning `""` does not mean omit the attribute**: methods like `GetDbNamespace()` or `GetCollection()` returning `""` in loongsuite are often gaps in loongsuite's implementation, not a spec signal that the attribute should be absent. Always check the [OTel semconv spec](https://opentelemetry.io/docs/specs/semconv/) for the signal type. If the attribute is defined in the spec and the data is derivable from the request (e.g. index name from the URL path → `db.namespace`), implement it in the hook. Write an integration test that asserts the attribute — this forces correct implementation at authoring time.
- **Missing replace directives**: run `make crosslink` to fix
- **Wrong param index**: for methods, index 0 is the receiver; double-check against the target function signature
- **initOnce not resettable in tests**: declare `var initOnce sync.Once` at package level (not inside a function)
- **Semconv version**: use `go.opentelemetry.io/otel/semconv/v1.37.0`; check `.semconv-version` file
- **Test app location**: `TestFixture` resolves apps from `test/apps/<name>/`; directory name must match `BuildAndRun()` argument. For multi-version apps, use path-like names: `f.BuildAndRun("<name>/<version>")` resolves to `test/apps/<name>/<version>/`
- **Multi-version: only one version migrated**: always check `loongsuite-go-agent/test/<name>/` for ALL version subdirectories in Step 0. Create a test app and integration sub-test for EACH version, not just the first or minimum
- **Multi-version: `make tidy/test-apps`**: uses `find test/apps -name "go.mod"` so nested version directories are auto-discovered — no Makefile changes needed
- **Struct-injection modules excluded from `make test-unit/pkg`**: the Makefile excludes `runtime`, `databasesql`, and the root `pkg` module from `make test-unit/pkg`. Any new instrumentation that uses struct injection (Complex-DB tier) must also be excluded there, or run unit tests directly via `go test -C pkg/instrumentation/<name> ./...` instead of relying on the make target
