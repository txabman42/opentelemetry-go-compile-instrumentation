// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package mux

import (
	"net/http"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel/trace"

	"github.com/open-telemetry/opentelemetry-go-compile-instrumentation/pkg/inst"
	"github.com/open-telemetry/opentelemetry-go-compile-instrumentation/pkg/instrumentation/shared"
)

const instrumentationKey = "gorilla/mux"

type muxEnabler struct{}

func (e muxEnabler) Enable() bool { return shared.Instrumented(instrumentationKey) }

var enabler = muxEnabler{}

// BeforeSetCurrentRoute hooks github.com/gorilla/mux.setCurrentRoute (v1.3.0–v1.7.4).
// The route argument is untyped (interface{}) in this version range — we perform
// a runtime type assertion to *mux.Route before calling GetPathTemplate.
//
// The span is retrieved from req.Context(), which the net/http server hook
// populates before calling the handler and mux's route resolution.
func BeforeSetCurrentRoute(ictx inst.HookContext, req *http.Request, route interface{}) {
	if !enabler.Enable() {
		return
	}
	if req == nil || route == nil {
		return
	}
	r, ok := route.(*mux.Route)
	if !ok {
		return
	}
	tmpl, err := r.GetPathTemplate()
	if err != nil || req.URL == nil || tmpl == req.URL.Path {
		return
	}
	renameSpan(req, tmpl)
}

// BeforeRequestWithRoute hooks github.com/gorilla/mux.requestWithRoute (v1.7.4–v1.8.2).
// Since v1.7.4, the route argument is already typed as *mux.Route.
func BeforeRequestWithRoute(ictx inst.HookContext, req *http.Request, route *mux.Route) {
	if !enabler.Enable() {
		return
	}
	if req == nil || route == nil {
		return
	}
	tmpl, err := route.GetPathTemplate()
	if err != nil || req.URL == nil || tmpl == req.URL.Path {
		return
	}
	renameSpan(req, tmpl)
}

// renameSpan renames the span stored in the request context to the matched
// route template. The net/http server instrumentation stores the active span
// in req.Context() before the handler chain runs, so it is available here.
func renameSpan(req *http.Request, name string) {
	span := trace.SpanFromContext(req.Context())
	if span.IsRecording() {
		span.SetName(name)
	}
}
