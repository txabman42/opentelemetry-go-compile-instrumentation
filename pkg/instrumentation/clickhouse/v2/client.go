// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package clickhousev2

import (
	"context"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/open-telemetry/opentelemetry-go-compile-instrumentation/pkg/inst"
	"github.com/open-telemetry/opentelemetry-go-compile-instrumentation/pkg/instrumentation/clickhouse/v2/semconv"
	"github.com/open-telemetry/opentelemetry-go-compile-instrumentation/pkg/instrumentation/shared"
)

const (
	instrumentationName = "github.com/open-telemetry/opentelemetry-go-compile-instrumentation/pkg/instrumentation/clickhouse/v2"
	instrumentationKey  = "CLICKHOUSE"
)

var (
	logger   = shared.Logger()
	tracer   trace.Tracer
	initOnce sync.Once
)

type clickhouseEnabler struct{}

func (e clickhouseEnabler) Enable() bool { return shared.Instrumented(instrumentationKey) }

var enabler = clickhouseEnabler{}

func moduleVersion() string {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return "dev"
	}
	if bi.Main.Version != "" && bi.Main.Version != "(devel)" {
		return bi.Main.Version
	}
	return "dev"
}

func initInstrumentation() {
	initOnce.Do(func() {
		version := moduleVersion()
		if err := shared.SetupOTelSDK(
			"go.opentelemetry.io/compile-instrumentation/clickhouse/v2",
			version,
		); err != nil {
			logger.Error("failed to setup OTel SDK", "error", err)
		}
		tracer = otel.GetTracerProvider().Tracer(instrumentationName, trace.WithInstrumentationVersion(version))
		if err := shared.StartRuntimeMetrics(); err != nil {
			logger.Error("failed to start runtime metrics", "error", err)
		}
	})
}

// BeforeOpen stores the *clickhouse.Options so AfterOpen can access them.
func BeforeOpen(ictx inst.HookContext, options *clickhouse.Options) {
	if !enabler.Enable() {
		return
	}
	ictx.SetData(options)
}

// AfterOpen wraps the returned driver.Conn with an OtelConn that auto-traces every operation.
func AfterOpen(ictx inst.HookContext, con driver.Conn, err error) {
	if err != nil || con == nil {
		return
	}
	if !enabler.Enable() {
		return
	}
	opts, ok := ictx.GetData().(*clickhouse.Options)
	if !ok || opts == nil {
		return
	}
	ictx.SetReturnVal(0, newOtelConn(con, opts))
}

// OtelConn wraps a driver.Conn and records a span for every operation.
type OtelConn struct {
	conn driver.Conn
	opts *clickhouse.Options
}

func newOtelConn(conn driver.Conn, opts *clickhouse.Options) *OtelConn {
	return &OtelConn{conn: conn, opts: opts}
}

func (c *OtelConn) addr() string {
	return strings.Join(c.opts.Addr, ",")
}

func (c *OtelConn) record(op, statement string, params []any, fn func() error) error {
	if !enabler.Enable() {
		return fn()
	}
	initInstrumentation()

	req := semconv.ClickHouseRequest{
		Op:        op,
		Statement: statement,
		DbName:    c.opts.Auth.Database,
		Addr:      c.addr(),
		Params:    params,
	}
	attrs := semconv.ClickHouseClientTraceAttrs(req)

	start := time.Now()
	ctx := context.Background()
	_, span := tracer.Start(ctx, op,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(attrs...),
		trace.WithTimestamp(start),
	)

	runErr := fn()

	if runErr != nil {
		span.RecordError(runErr)
		span.SetStatus(codes.Error, runErr.Error())
	}
	span.End(trace.WithTimestamp(time.Now()))
	return runErr
}

func (c *OtelConn) Contributors() []string {
	return c.conn.Contributors()
}

func (c *OtelConn) ServerVersion() (*driver.ServerVersion, error) {
	var sv *driver.ServerVersion
	err := c.record("SERVER_VERSION", "SERVER_VERSION", nil, func() error {
		var e error
		sv, e = c.conn.ServerVersion()
		return e
	})
	return sv, err
}

func (c *OtelConn) Select(ctx context.Context, dest any, query string, args ...any) error {
	return c.record("SELECT", query, args, func() error {
		return c.conn.Select(ctx, dest, query, args...)
	})
}

func (c *OtelConn) Query(ctx context.Context, query string, args ...any) (driver.Rows, error) {
	var rows driver.Rows
	err := c.record("QUERY", query, args, func() error {
		var e error
		rows, e = c.conn.Query(ctx, query, args...)
		return e
	})
	return rows, err
}

func (c *OtelConn) QueryRow(ctx context.Context, query string, args ...any) driver.Row {
	row := c.conn.QueryRow(ctx, query, args...)
	_ = c.record("QUERY_ROW", query, args, func() error {
		return row.Err()
	})
	return row
}

func (c *OtelConn) PrepareBatch(
	ctx context.Context,
	query string,
	opts ...driver.PrepareBatchOption,
) (driver.Batch, error) {
	var batch driver.Batch
	err := c.record("PREPARE_BATCH", query, nil, func() error {
		var e error
		batch, e = c.conn.PrepareBatch(ctx, query, opts...)
		return e
	})
	return batch, err
}

func (c *OtelConn) Exec(ctx context.Context, query string, args ...any) error {
	return c.record("EXEC", query, args, func() error {
		return c.conn.Exec(ctx, query, args...)
	})
}

func (c *OtelConn) AsyncInsert(ctx context.Context, query string, wait bool, args ...any) error {
	return c.record("ASYNC_INSERT", query, args, func() error {
		return c.conn.AsyncInsert(ctx, query, wait, args...)
	})
}

func (c *OtelConn) Ping(ctx context.Context) error {
	return c.record("PING", "PING", nil, func() error {
		return c.conn.Ping(ctx)
	})
}

func (c *OtelConn) Stats() driver.Stats {
	return c.conn.Stats()
}

func (c *OtelConn) Close() error {
	return c.conn.Close()
}
