// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package mux

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/open-telemetry/opentelemetry-go-compile-instrumentation/pkg/inst/insttest"
)

func setupTestTracer(t *testing.T) *tracetest.SpanRecorder {
	t.Helper()
	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))
	otel.SetTracerProvider(tp)
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })
	return sr
}

// newSpanRequest creates an *http.Request with an active recording span in its
// context, simulating what the net/http server hook provides before routing.
func newSpanRequest(t *testing.T, path, spanName string, sr *tracetest.SpanRecorder) (*http.Request, func()) {
	t.Helper()
	ctx, span := otel.Tracer("test").Start(context.Background(), spanName)
	req := &http.Request{
		Method: "GET",
		URL:    &url.URL{Path: path},
		Header: make(http.Header),
	}
	req = req.WithContext(ctx)
	return req, func() { span.End() }
}

// ---- Enabler tests ----

func TestMuxEnabler_EnabledByDefault(t *testing.T) {
	assert.True(t, enabler.Enable())
}

func TestMuxEnabler_DisabledViaEnv(t *testing.T) {
	t.Setenv("OTEL_GO_DISABLED_INSTRUMENTATIONS", "gorilla/mux")
	assert.False(t, enabler.Enable())
}

func TestMuxEnabler_EnabledViaEnvList(t *testing.T) {
	t.Setenv("OTEL_GO_ENABLED_INSTRUMENTATIONS", "gorilla/mux")
	assert.True(t, enabler.Enable())
}

func TestMuxEnabler_ExcludedFromEnvList(t *testing.T) {
	t.Setenv("OTEL_GO_ENABLED_INSTRUMENTATIONS", "redis,grpc")
	assert.False(t, enabler.Enable())
}

// ---- BeforeSetCurrentRoute (v1.3.0) ----

func TestBeforeSetCurrentRoute_Disabled(t *testing.T) {
	t.Setenv("OTEL_GO_DISABLED_INSTRUMENTATIONS", "gorilla/mux")
	sr := setupTestTracer(t)
	req, end := newSpanRequest(t, "/test/123", "original", sr)
	defer end()

	ictx := insttest.NewMockHookContext()
	BeforeSetCurrentRoute(ictx, req, buildRoute(t, "/test/{id}"))

	end()
	spans := sr.Ended()
	require.Len(t, spans, 1)
	assert.Equal(t, "original", spans[0].Name())
}

func TestBeforeSetCurrentRoute_NilRequest(t *testing.T) {
	t.Setenv("OTEL_GO_ENABLED_INSTRUMENTATIONS", "gorilla/mux")
	ictx := insttest.NewMockHookContext()
	// Must not panic.
	BeforeSetCurrentRoute(ictx, nil, buildRoute(t, "/test/{id}"))
}

func TestBeforeSetCurrentRoute_NilRoute(t *testing.T) {
	t.Setenv("OTEL_GO_ENABLED_INSTRUMENTATIONS", "gorilla/mux")
	sr := setupTestTracer(t)
	req, end := newSpanRequest(t, "/test/123", "original", sr)
	defer end()

	ictx := insttest.NewMockHookContext()
	BeforeSetCurrentRoute(ictx, req, nil)

	end()
	spans := sr.Ended()
	require.Len(t, spans, 1)
	assert.Equal(t, "original", spans[0].Name())
}

func TestBeforeSetCurrentRoute_WrongRouteType(t *testing.T) {
	t.Setenv("OTEL_GO_ENABLED_INSTRUMENTATIONS", "gorilla/mux")
	sr := setupTestTracer(t)
	req, end := newSpanRequest(t, "/test/123", "original", sr)
	defer end()

	ictx := insttest.NewMockHookContext()
	BeforeSetCurrentRoute(ictx, req, "not-a-mux-route")

	end()
	spans := sr.Ended()
	require.Len(t, spans, 1)
	assert.Equal(t, "original", spans[0].Name())
}

func TestBeforeSetCurrentRoute_RenamesSpan(t *testing.T) {
	t.Setenv("OTEL_GO_ENABLED_INSTRUMENTATIONS", "gorilla/mux")
	sr := setupTestTracer(t)
	req, end := newSpanRequest(t, "/test/123", "original", sr)

	ictx := insttest.NewMockHookContext()
	BeforeSetCurrentRoute(ictx, req, buildRoute(t, "/test/{id}"))

	end()
	spans := sr.Ended()
	require.Len(t, spans, 1)
	assert.Equal(t, "/test/{id}", spans[0].Name())
}

func TestBeforeSetCurrentRoute_SamePathNoRename(t *testing.T) {
	t.Setenv("OTEL_GO_ENABLED_INSTRUMENTATIONS", "gorilla/mux")
	sr := setupTestTracer(t)
	req, end := newSpanRequest(t, "/test/path", "original", sr)
	defer end()

	ictx := insttest.NewMockHookContext()
	// Template equals the raw URL path — skip rename.
	BeforeSetCurrentRoute(ictx, req, buildRoute(t, "/test/path"))

	end()
	spans := sr.Ended()
	require.Len(t, spans, 1)
	assert.Equal(t, "original", spans[0].Name())
}

func TestBeforeSetCurrentRoute_NilURL(t *testing.T) {
	t.Setenv("OTEL_GO_ENABLED_INSTRUMENTATIONS", "gorilla/mux")
	sr := setupTestTracer(t)
	ctx, span := otel.Tracer("test").Start(context.Background(), "original")
	defer span.End()

	req := &http.Request{
		Method: "GET",
		Header: make(http.Header),
		// URL is nil
	}
	req = req.WithContext(ctx)

	ictx := insttest.NewMockHookContext()
	BeforeSetCurrentRoute(ictx, req, buildRoute(t, "/test/{id}"))

	span.End()
	spans := sr.Ended()
	require.Len(t, spans, 1)
	// Span must NOT be renamed because URL is nil.
	assert.Equal(t, "original", spans[0].Name())
}

func TestBeforeSetCurrentRoute_EmptyRoute(t *testing.T) {
	t.Setenv("OTEL_GO_ENABLED_INSTRUMENTATIONS", "gorilla/mux")
	ictx := insttest.NewMockHookContext()
	// Empty route has no path template — GetPathTemplate returns an error.
	r := mux.NewRouter()
	BeforeSetCurrentRoute(ictx, newTestRequest("/test"), r.NewRoute())
}

// ---- BeforeRequestWithRoute (v1.7.4) ----

func TestBeforeRequestWithRoute_Disabled(t *testing.T) {
	t.Setenv("OTEL_GO_DISABLED_INSTRUMENTATIONS", "gorilla/mux")
	sr := setupTestTracer(t)
	req, end := newSpanRequest(t, "/test/123", "original", sr)
	defer end()

	ictx := insttest.NewMockHookContext()
	BeforeRequestWithRoute(ictx, req, buildRoute(t, "/test/{id}"))

	end()
	spans := sr.Ended()
	require.Len(t, spans, 1)
	assert.Equal(t, "original", spans[0].Name())
}

func TestBeforeRequestWithRoute_NilRoute(t *testing.T) {
	t.Setenv("OTEL_GO_ENABLED_INSTRUMENTATIONS", "gorilla/mux")
	sr := setupTestTracer(t)
	req, end := newSpanRequest(t, "/test/123", "original", sr)
	defer end()

	ictx := insttest.NewMockHookContext()
	BeforeRequestWithRoute(ictx, req, nil)

	end()
	spans := sr.Ended()
	require.Len(t, spans, 1)
	assert.Equal(t, "original", spans[0].Name())
}

func TestBeforeRequestWithRoute_NilRequest(t *testing.T) {
	t.Setenv("OTEL_GO_ENABLED_INSTRUMENTATIONS", "gorilla/mux")
	ictx := insttest.NewMockHookContext()
	// Must not panic.
	BeforeRequestWithRoute(ictx, nil, buildRoute(t, "/test/{id}"))
}

func TestBeforeRequestWithRoute_RenamesSpan(t *testing.T) {
	t.Setenv("OTEL_GO_ENABLED_INSTRUMENTATIONS", "gorilla/mux")
	sr := setupTestTracer(t)
	req, end := newSpanRequest(t, "/api/users/42", "original", sr)

	ictx := insttest.NewMockHookContext()
	BeforeRequestWithRoute(ictx, req, buildRoute(t, "/api/users/{id}"))

	end()
	spans := sr.Ended()
	require.Len(t, spans, 1)
	assert.Equal(t, "/api/users/{id}", spans[0].Name())
}

func TestBeforeRequestWithRoute_PrefixPattern(t *testing.T) {
	t.Setenv("OTEL_GO_ENABLED_INSTRUMENTATIONS", "gorilla/mux")
	sr := setupTestTracer(t)
	req, end := newSpanRequest(t, "/1/countries/2", "original", sr)

	ictx := insttest.NewMockHookContext()
	BeforeRequestWithRoute(ictx, req, buildRoute(t, "/{name}/countries/{country}"))

	end()
	spans := sr.Ended()
	require.Len(t, spans, 1)
	assert.Equal(t, "/{name}/countries/{country}", spans[0].Name())
}

func TestBeforeRequestWithRoute_SamePathNoRename(t *testing.T) {
	t.Setenv("OTEL_GO_ENABLED_INSTRUMENTATIONS", "gorilla/mux")
	sr := setupTestTracer(t)
	req, end := newSpanRequest(t, "/test/path", "original", sr)
	defer end()

	ictx := insttest.NewMockHookContext()
	BeforeRequestWithRoute(ictx, req, buildRoute(t, "/test/path"))

	end()
	spans := sr.Ended()
	require.Len(t, spans, 1)
	assert.Equal(t, "original", spans[0].Name())
}

func TestBeforeRequestWithRoute_NilURL(t *testing.T) {
	t.Setenv("OTEL_GO_ENABLED_INSTRUMENTATIONS", "gorilla/mux")
	sr := setupTestTracer(t)
	ctx, span := otel.Tracer("test").Start(context.Background(), "original")
	defer span.End()

	req := &http.Request{Method: "GET", Header: make(http.Header)}
	req = req.WithContext(ctx)

	ictx := insttest.NewMockHookContext()
	BeforeRequestWithRoute(ictx, req, buildRoute(t, "/test/{id}"))

	span.End()
	spans := sr.Ended()
	require.Len(t, spans, 1)
	assert.Equal(t, "original", spans[0].Name())
}

func TestBeforeRequestWithRoute_EmptyRoute(t *testing.T) {
	t.Setenv("OTEL_GO_ENABLED_INSTRUMENTATIONS", "gorilla/mux")
	ictx := insttest.NewMockHookContext()
	r := mux.NewRouter()
	BeforeRequestWithRoute(ictx, newTestRequest("/test"), r.NewRoute())
}

// ---- Helpers ----

func newTestRequest(path string) *http.Request {
	return &http.Request{
		Method: "GET",
		URL:    &url.URL{Path: path},
		Header: make(http.Header),
	}
}

func buildRoute(t *testing.T, tmpl string) *mux.Route {
	t.Helper()
	r := mux.NewRouter()
	return r.NewRoute().Path(tmpl)
}
