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

**Read loongsuite test scenarios first** — `loongsuite-go-agent/test/<name>_tests.go` lists all scenarios (basic, HTTP/2, HTTPS, metrics, etc.) and `loongsuite-go-agent/test/<name>/test_*.go` shows the exact attributes asserted for each. Each registered `NewGeneralTestCase` should map to a `t.Run` sub-test here. The `verifier.Verify*Attributes` calls in the loongsuite test apps are the ground truth for which semconv attributes are expected per scenario.

Read an existing test for style: `test/integration/redis_client_test.go` (simple) or `test/integration/db_client_test.go` (Complex-DB).

**Pattern:**
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

**Available semconv helpers** (`test/testutil/semconv.go`):
- `RequireHTTPClientSemconv` / `RequireHTTPServerSemconv`
- `RequireGRPCClientSemconv` / `RequireGRPCServerSemconv`
- `RequireDBClientSemconv`
- `RequireRedisClientSemconv`
- `RequireMessagingSemconv`

If none fits, add `RequireXxxSemconv(t, span, ...)` to `test/testutil/semconv.go`. New helpers must only assert attributes that are defined in the [OTel Semantic Conventions](https://opentelemetry.io/docs/specs/semconv/) for that signal (e.g. [Messaging spans](https://opentelemetry.io/docs/specs/semconv/messaging/messaging-spans/), [Database spans](https://opentelemetry.io/docs/specs/semconv/database/database-spans/), [RPC spans](https://opentelemetry.io/docs/specs/semconv/rpc/rpc-spans/)). Do not invent custom attribute names — check the spec first.

## E2E Test

Write `test/e2e/<name>_test.go` with `//go:build e2e` **only** when multiple instrumented processes interact and trace context propagation must be validated (e.g. messaging producer → consumer). Read `test/e2e/http_test.go` for the pattern.

## Quality Gates

All must pass. Fix failures before marking migration complete.

```bash
go test -C pkg/instrumentation/<name> ./...  # unit
make format/license                           # license headers
make go-mod-tidy && make crosslink            # module consistency
make build                                    # packages hooks into binary
make test-integration                         # integration
make lint/go                                  # linter — no suppressions
```

## Common Pitfalls

- **Missing replace directives**: run `make crosslink` to fix
- **Wrong param index**: for methods, index 0 is the receiver; double-check against the target function signature
- **initOnce not resettable in tests**: declare `var initOnce sync.Once` at package level (not inside a function)
- **Semconv version**: use `go.opentelemetry.io/otel/semconv/v1.37.0`; check `.semconv-version` file
- **Test app location**: `TestFixture` resolves apps from `test/apps/<name>/`; directory name must match `BuildAndRun()` argument
- **Struct-injection modules excluded from make test-unit/pkg**: if your instrumentation uses struct injection, `go test -C pkg/instrumentation/<name> ./...` directly instead of relying on the make target
