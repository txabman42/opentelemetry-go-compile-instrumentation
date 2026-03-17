# Tier Examples

Instead of duplicating code here, this file points to the canonical source and target files to read for each tier. Read both sides before writing any code.

## Simple Tier — GLS span-rename

**Loongsuite source to read:**
- `loongsuite-go-agent/pkg/rules/mux/mux_server_setup.go` — two `OnEnter` hooks that call `trace.LocalRootSpanFromGLS().SetName(tmpl)`
- `loongsuite-go-agent/tool/data/rules/mux.json` — two JSON entries with version ranges `[1.3.0,1.7.4)` and `[1.7.4,1.8.2)`

**Key translation deltas:**
1. `trace.LocalRootSpanFromGLS()` has no direct equivalent — see translation pattern 7 in `translation-patterns.md`
2. No tracer, no span creation, no `initInstrumentation()` — only `shared.Instrumented("KEY")` guard
3. Version `[1.3.0,1.7.4)` → YAML `version: v1.3.0,v1.7.4`
4. `//go:linkname muxRoute130OnEnter github.com/gorilla/mux.muxRoute130OnEnter` → exported `func BeforeRouteMatch130(ictx inst.HookContext, ...)` named in YAML `before:`

**No OTel equivalent exists yet** — this tier has not been migrated. Use the translation patterns and write from scratch.

---

## Standard Tier — Instrumenter flatten

**Loongsuite source to read:**
- `loongsuite-go-agent/pkg/rules/goredis/` — hooks `afterNew*` calling `goredisInstrumenter.Start` pattern; simpler than kafka
- `loongsuite-go-agent/pkg/rules/http/client_setup.go` + `http/net_http_otel_instrumenter.go` — full `BuildPropagatingToDownstreamInstrumenter` pattern

**OTel reference to read (the target pattern):**
- `pkg/instrumentation/nethttp/client/client_hook.go` — canonical `BeforeRoundTrip`/`AfterRoundTrip`: tracer.Start, propagator.Inject, ictx.SetKeyData("span"), span.End
- `pkg/instrumentation/redis/v9/client_hook.go` — simpler: hook on constructor, add OTel hook to client
- `pkg/instrumentation/grpc/client/client_hook.go` — stats handler injection pattern (for frameworks that use middleware/stats handlers)

**Key translation deltas:**
1. `builder.Init().SetSpanNameExtractor(...).AddAttributesExtractor(...)` → determine span name and attributes inline from request fields
2. `instrumenter.Start(ctx, req)` → `tracer.Start(ctx, spanName, trace.WithSpanKind(...), trace.WithAttributes(attrs...))`
3. `instrumenter.End(ctx, req, resp, err)` → `span.SetAttributes(respAttrs...); span.SetStatus(...); span.End()`
4. `BuildPropagatingToDownstreamInstrumenter(carrierFn, prop)` → `propagator.Inject(ctx, carrier)` after `tracer.Start`
5. `BuildPropagatingFromUpstreamInstrumenter(carrierFn, prop)` → `ctx = propagator.Extract(ctx, carrier)` before `tracer.Start`

---

## Complex-DB Tier — Struct injection

**Loongsuite source to read:**
- `loongsuite-go-agent/pkg/rules/gorm/setup.go` — `afterGormOpen` OnExit hook that registers GORM callbacks
- `loongsuite-go-agent/tool/data/rules/gorm.json` — struct injection entry for `gorm.io/driver/mysql.Dialector`

**OTel reference to read (the target pattern):**
- `pkg/instrumentation/databasesql/client.go` — canonical Complex-DB: Open hook stores metadata in injected fields; every operation reads them for semconv
- `pkg/instrumentation/databasesql/db.yaml` — canonical struct injection YAML with `new_field` list

**Key translation deltas:**
1. JSON `StructType`/`FieldName`/`FieldType` → YAML `struct`/`new_field` (see translation-patterns.md section 1)
2. Hooks reading injected fields (e.g. `db.Endpoint`) cannot be unit tested — annotate with comment
3. GORM callback registration stays as Go code; replace `gormInstrumenter.Start/End` with `tracer.Start`/`span.End`

---

## Messaging Tier — Propagation carriers

**Loongsuite source to read:**
- `loongsuite-go-agent/pkg/rules/segmentio-kafka-go/kafka_producer_setup.go` — `producerWriteMessagesOnEnter/Exit`
- `loongsuite-go-agent/pkg/rules/segmentio-kafka-go/kafka_consumer_setup.go` — `consumerReadMessageOnEnter/Exit`
- `loongsuite-go-agent/pkg/rules/segmentio-kafka-go/kafka_otel_instrumenter.go` — carrier structs `kafkaProducerCarrier`, `kafkaConsumerCarrier` and their builder setup

**No OTel equivalent exists yet.** The carrier structs are pure Go (`propagation.TextMapCarrier`) — copy them unchanged.

**Key translation deltas:**
1. Producer: copy carrier struct → `propagator.Inject(ctx, &kafkaProducerCarrier{messages: msgPtrs})` after `tracer.Start`
2. Consumer: copy carrier struct → `ctx = propagator.Extract(parentCtx, kafkaConsumerCarrier{message: msg})` before `tracer.Start`
3. `StartAndEnd(parentCtx, req, nil, err, startTime, endTime)` → `tracer.Start` with `trace.WithTimestamp(startTime)` then `span.End(trace.WithTimestamp(endTime))`
4. `AlwaysProducerExtractor` → `trace.WithSpanKind(trace.SpanKindProducer)`
5. `AlwaysConsumerExtractor` → `trace.WithSpanKind(trace.SpanKindConsumer)`
6. `MessageSpanNameExtractor{OperationName: "publish"}` → `topic + " publish"`
7. Semconv: `MessagingSystemKey.String("kafka")`, `MessagingDestinationNameKey.String(topic)`, `MessagingOperationTypeKey.String("publish"/"process")`
