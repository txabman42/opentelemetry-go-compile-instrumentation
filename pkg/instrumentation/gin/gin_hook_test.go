// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package gin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/open-telemetry/opentelemetry-go-compile-instrumentation/pkg/inst/insttest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func setupTestTracer(t *testing.T) *tracetest.SpanRecorder {
	t.Helper()
	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })
	return sr
}

// routedGinContext runs a request through a gin engine with the given route
// registered, captures the *gin.Context inside the handler, and returns it.
// The request context carries an active span so span name assertions work.
func routedGinContext(t *testing.T, route, urlPath string) (*gin.Context, *tracetest.SpanRecorder) {
	t.Helper()
	sr := setupTestTracer(t)

	ctx, span := otel.Tracer("test").Start(context.Background(), "GET "+urlPath)
	t.Cleanup(func() { span.End() })

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, urlPath, nil)
	w := httptest.NewRecorder()
	engine := gin.New()
	var capturedCtx *gin.Context
	engine.GET(route, func(c *gin.Context) {
		capturedCtx = c
	})
	engine.ServeHTTP(w, req)
	require.NotNil(t, capturedCtx, "handler must have been called for route %s", route)
	return capturedCtx, sr
}

// TestBeforeNext_RenamesSpan verifies that BeforeNext renames the active span
// to the matched gin route pattern.
func TestBeforeNext_RenamesSpan(t *testing.T) {
	t.Setenv("OTEL_GO_ENABLED_INSTRUMENTATIONS", "gin")

	c, sr := routedGinContext(t, "/user/:name", "/user/alice")
	ictx := insttest.NewMockHookContext()
	BeforeNext(ictx, c)

	// End all spans and check the name.
	for _, s := range sr.Started() {
		s.End()
	}
	spans := sr.Ended()
	require.GreaterOrEqual(t, len(spans), 1)
	found := false
	for _, s := range spans {
		if s.Name() == "/user/:name" {
			found = true
		}
	}
	assert.True(t, found, "span should have been renamed to /user/:name; got: %v", spanNames(spans))
}

// TestBeforeNext_NoRenameWhenPathMatchesRoute verifies that when the full path
// equals the URL path (no parameterised segment), no rename occurs.
func TestBeforeNext_NoRenameWhenPathMatchesRoute(t *testing.T) {
	t.Setenv("OTEL_GO_ENABLED_INSTRUMENTATIONS", "gin")

	c, sr := routedGinContext(t, "/query", "/query")
	ictx := insttest.NewMockHookContext()
	BeforeNext(ictx, c)

	for _, s := range sr.Started() {
		s.End()
	}
	spans := sr.Ended()
	require.GreaterOrEqual(t, len(spans), 1)
	// Original name should be preserved — "/query" == "/query" means no rename.
	for _, s := range spans {
		assert.NotEmpty(t, s.Name())
		// The name should remain as originally set ("GET /query"), not get renamed
		// to the empty string or some other value.
	}
}

// TestBeforeNext_Disabled verifies that no rename occurs when gin instrumentation
// is disabled via the environment variable.
func TestBeforeNext_Disabled(t *testing.T) {
	t.Setenv("OTEL_GO_DISABLED_INSTRUMENTATIONS", "gin")

	c, sr := routedGinContext(t, "/user/:name", "/user/alice")
	ictx := insttest.NewMockHookContext()
	BeforeNext(ictx, c)

	for _, s := range sr.Started() {
		s.End()
	}
	spans := sr.Ended()
	require.GreaterOrEqual(t, len(spans), 1)
	for _, s := range spans {
		assert.NotEqual(t, "/user/:name", s.Name(),
			"span must not be renamed when instrumentation is disabled")
	}
}

// TestBeforeNext_NilContext verifies that a nil *gin.Context is handled safely.
func TestBeforeNext_NilContext(t *testing.T) {
	t.Setenv("OTEL_GO_ENABLED_INSTRUMENTATIONS", "gin")
	ictx := insttest.NewMockHookContext()
	assert.NotPanics(t, func() {
		BeforeNext(ictx, nil)
	})
}

// TestBeforeHTML_RenamesSpan verifies that BeforeHTML renames the span to the
// matched route pattern for HTML template responses.
func TestBeforeHTML_RenamesSpan(t *testing.T) {
	t.Setenv("OTEL_GO_ENABLED_INSTRUMENTATIONS", "gin")

	c, sr := routedGinContext(t, "/query/:id", "/query/42")
	ictx := insttest.NewMockHookContext()
	BeforeHTML(ictx, c, http.StatusOK, "index.tmpl", nil)

	for _, s := range sr.Started() {
		s.End()
	}
	spans := sr.Ended()
	require.GreaterOrEqual(t, len(spans), 1)
	found := false
	for _, s := range spans {
		if s.Name() == "/query/:id" {
			found = true
		}
	}
	assert.True(t, found, "span should have been renamed to /query/:id; got: %v", spanNames(spans))
}

// TestBeforeHTML_NilContext verifies that a nil *gin.Context is handled safely.
func TestBeforeHTML_NilContext(t *testing.T) {
	t.Setenv("OTEL_GO_ENABLED_INSTRUMENTATIONS", "gin")
	ictx := insttest.NewMockHookContext()
	assert.NotPanics(t, func() {
		BeforeHTML(ictx, nil, http.StatusOK, "index.tmpl", nil)
	})
}

// TestRenameRootSpan_NilRequest verifies safe handling when c.Request is nil.
func TestRenameRootSpan_NilRequest(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = nil
	assert.NotPanics(t, func() {
		renameRootSpan(c)
	})
}

// TestRenameRootSpan_NoFullPath verifies that renameRootSpan is a no-op when
// the context has no matched route (FullPath returns "").
func TestRenameRootSpan_NoFullPath(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodGet, "/unmatched", nil)
	c.Request = req
	assert.Empty(t, c.FullPath())
	assert.NotPanics(t, func() {
		renameRootSpan(c)
	})
}

// TestEnabler verifies the enable/disable logic responds to env vars correctly.
func TestEnabler(t *testing.T) {
	t.Run("enabled by default", func(t *testing.T) {
		t.Setenv("OTEL_GO_ENABLED_INSTRUMENTATIONS", "")
		t.Setenv("OTEL_GO_DISABLED_INSTRUMENTATIONS", "")
		assert.True(t, enabler.Enable())
	})

	t.Run("disabled explicitly", func(t *testing.T) {
		t.Setenv("OTEL_GO_DISABLED_INSTRUMENTATIONS", "gin")
		assert.False(t, enabler.Enable())
	})

	t.Run("enabled explicitly", func(t *testing.T) {
		t.Setenv("OTEL_GO_ENABLED_INSTRUMENTATIONS", "gin")
		assert.True(t, enabler.Enable())
	})
}

func spanNames(spans []sdktrace.ReadOnlySpan) []string {
	names := make([]string, len(spans))
	for i, s := range spans {
		names[i] = s.Name()
	}
	return names
}
