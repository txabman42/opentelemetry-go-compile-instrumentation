# Translation Patterns

## 1. JSON to YAML Field Mapping

| Loongsuite JSON field | OTel YAML field | Notes |
|-----------------------|-----------------|-------|
| `ImportPath` | `target` | |
| `Function` | `func` | |
| `ReceiverType` | `recv` | e.g. `"*Transport"` |
| `OnEnter` | `before` | Export and rename to `BeforeX` |
| `OnExit` | `after` | Export and rename to `AfterX` |
| `Path` | `path` | Replace with otel module path |
| `Version` `[v1,v2)` | `version` `v1,v2` | Drop brackets; ensure `v` prefix |
| `StructType` | `struct` | Struct injection rule |
| `FieldName` + `FieldType` | `new_field: [{name, type}]` | |
| `Dependencies` | (not supported) | Drop — no equivalent |

Version: `[1.3.0,1.7.4)` → `v1.3.0,v1.7.4`. Single minimum: `[v1.2.0,)` → `v1.2.0`.

## 2. go.mod Template

```
module github.com/open-telemetry/opentelemetry-go-compile-instrumentation/pkg/instrumentation/<name>

go 1.25.0

replace github.com/open-telemetry/opentelemetry-go-compile-instrumentation/pkg => <depth to pkg/>

replace github.com/open-telemetry/opentelemetry-go-compile-instrumentation/pkg/instrumentation/shared => <depth to shared/>

require (
    github.com/open-telemetry/opentelemetry-go-compile-instrumentation/pkg v0.0.0
    github.com/open-telemetry/opentelemetry-go-compile-instrumentation/pkg/instrumentation/shared v0.0.0
    github.com/stretchr/testify v1.11.1
    go.opentelemetry.io/otel v1.40.0
    go.opentelemetry.io/otel/sdk v1.40.0
    go.opentelemetry.io/otel/trace v1.40.0
    <target-library> <version>
)
```

Replace directive depth: at `pkg/instrumentation/<name>/` → `../..` for `pkg`, `../shared` for `shared`. At `pkg/instrumentation/<name>/client/` → `../../..` and `../../shared`.

Check an existing `go.mod` at the same depth for the exact require versions to copy.

## 3. Hook Boilerplate Template

```go
// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package <name>

import (
    "runtime/debug"
    "sync"

    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/codes"
    "go.opentelemetry.io/otel/propagation"
    "go.opentelemetry.io/otel/trace"

    "github.com/open-telemetry/opentelemetry-go-compile-instrumentation/pkg/inst"
    "github.com/open-telemetry/opentelemetry-go-compile-instrumentation/pkg/instrumentation/shared"
)

const (
    instrumentationName = "github.com/open-telemetry/opentelemetry-go-compile-instrumentation/pkg/instrumentation/<name>"
    instrumentationKey  = "<NAME>"  // used in OTEL_GO_ENABLED/DISABLED_INSTRUMENTATIONS
)

var (
    logger     = shared.Logger()
    tracer     trace.Tracer
    propagator propagation.TextMapPropagator
    initOnce   sync.Once
)

func moduleVersion() string {
    bi, ok := debug.ReadBuildInfo()
    if !ok { return "dev" }
    if bi.Main.Version != "" && bi.Main.Version != "(devel)" { return bi.Main.Version }
    return "dev"
}

func initInstrumentation() {
    initOnce.Do(func() {
        version := moduleVersion()
        if err := shared.SetupOTelSDK("go.opentelemetry.io/compile-instrumentation/<name>", version); err != nil {
            logger.Error("failed to setup OTel SDK", "error", err)
        }
        tracer = otel.GetTracerProvider().Tracer(instrumentationName, trace.WithInstrumentationVersion(version))
        propagator = otel.GetTextMapPropagator()
        if err := shared.StartRuntimeMetrics(); err != nil {
            logger.Error("failed to start runtime metrics", "error", err)
        }
    })
}

type <name>Enabler struct{}
func (e <name>Enabler) Enable() bool { return shared.Instrumented(instrumentationKey) }
var enabler = <name>Enabler{}
```

Simple tier: omit `initInstrumentation`, `tracer`, `propagator`, `initOnce`.

## 4. api.CallContext → inst.HookContext

Direct rename — all methods are identical:

```
call api.CallContext  →  ictx inst.HookContext
call.SetData(v)       →  ictx.SetData(v)
call.GetData()        →  ictx.GetData()
call.SetKeyData(k,v)  →  ictx.SetKeyData(k,v)
call.GetKeyData(k)    →  ictx.GetKeyData(k)
call.SetParam(i,v)    →  ictx.SetParam(i,v)
call.GetParam(i)      →  ictx.GetParam(i)
call.SetReturnVal(i,v)→  ictx.SetReturnVal(i,v)
```

## 5. Instrumenter Builder → Direct SDK

```go
// Loongsuite:
ctx = inst.Start(parentCtx, req)
inst.End(ctx, req, resp, err)

// OTel:
// Before hook:
ctx, span := tracer.Start(ctx, spanName,
    trace.WithSpanKind(trace.SpanKindClient),
    trace.WithAttributes(attrs...),
)
ictx.SetKeyData("span", span)

// After hook:
span := ictx.GetKeyData("span").(trace.Span)
defer span.End()
if err != nil { span.RecordError(err); span.SetStatus(codes.Error, err.Error()) }
```

For `StartAndEnd(parentCtx, req, nil, err, startTime, endTime)`:
```go
_, span := tracer.Start(ctx, spanName, ..., trace.WithTimestamp(startTime))
// set attributes
span.End(trace.WithTimestamp(endTime))
```

## 6. AttrsGetter → Direct Attributes

Replace `AttrsGetter` interface implementations with direct `[]attribute.KeyValue`:

```go
// Instead of implementing GetSystem(), GetDestination(), etc:
attrs := []attribute.KeyValue{
    semconv.MessagingSystemKey.String("kafka"),
    semconv.MessagingDestinationNameKey.String(topic),
    semconv.MessagingOperationTypeKey.String("publish"),
}
```

Use `go.opentelemetry.io/otel/semconv/v1.37.0` (check `.semconv-version` file).

## 7. GLS: LocalRootSpanFromGLS → runtime package

```go
// Loongsuite:
lcs := trace.LocalRootSpanFromGLS()
if lcs != nil { lcs.SetName(newName) }

// OTel — import "github.com/open-telemetry/opentelemetry-go-compile-instrumentation/pkg/instrumentation/runtime":
traceCtx := runtime.GetTraceContextFromGLS()
if traceCtx != nil {
    if ctx, ok := traceCtx.(context.Context); ok {
        span := trace.SpanFromContext(ctx)
        if span.IsRecording() { span.SetName(newName) }
    }
}
```

## 8. //go:linkname → Exported Functions

```go
// Loongsuite:
//go:linkname clientOnEnter net/http.clientOnEnter
func clientOnEnter(call api.CallContext, ...) { ... }

// OTel — just export the function; the tool generates linkname from YAML:
func BeforeRoundTrip(ictx inst.HookContext, ...) { ... }
```

## 9. Propagation Carriers

Keep carrier structs unchanged — they implement `propagation.TextMapCarrier`. Just change how they're used:

```go
// Producer (inject):
propagator.Inject(ctx, &myProducerCarrier{...})  // after tracer.Start

// Consumer (extract):
ctx = propagator.Extract(parentCtx, myConsumerCarrier{...})  // before tracer.Start
```
