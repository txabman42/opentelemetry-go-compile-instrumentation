// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package gin

import (
	"github.com/gin-gonic/gin"
	"github.com/open-telemetry/opentelemetry-go-compile-instrumentation/pkg/inst"
	"github.com/open-telemetry/opentelemetry-go-compile-instrumentation/pkg/instrumentation/shared"
	"go.opentelemetry.io/otel/trace"
)

const instrumentationKey = "gin"

type ginEnabler struct{}

func (ginEnabler) Enable() bool { return shared.Instrumented(instrumentationKey) }

var enabler = ginEnabler{}

// renameRootSpan renames the active span for this request to the gin route
// pattern when it differs from the raw request URL path. This avoids
// high-cardinality span names (e.g. "/user/alice" → "/user/:name").
//
// The span is retrieved from c.Request.Context(), which is populated by the
// nethttp server instrumentation before gin's handler chain runs.
func renameRootSpan(c *gin.Context) {
	if c == nil {
		return
	}
	fullPath := c.FullPath()
	if fullPath == "" {
		return
	}
	if c.Request == nil || c.Request.URL == nil {
		return
	}
	if fullPath == c.Request.URL.Path {
		return
	}
	span := trace.SpanFromContext(c.Request.Context())
	if span.IsRecording() {
		span.SetName(fullPath)
	}
}

// BeforeNext is called before (*gin.Context).Next(). It renames the active
// span to the matched route pattern (e.g. "/user/:name") so that HTTP server
// spans are grouped by route rather than by the raw request URL.
func BeforeNext(_ inst.HookContext, c *gin.Context) {
	if !enabler.Enable() {
		return
	}
	renameRootSpan(c)
}

// BeforeHTML is called before (*gin.Context).HTML(). It applies the same
// route-pattern rename as BeforeNext for HTML template responses.
func BeforeHTML(_ inst.HookContext, c *gin.Context, _ int, _ string, _ any) {
	if !enabler.Enable() {
		return
	}
	renameRootSpan(c)
}
