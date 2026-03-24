// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package v8

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/open-telemetry/opentelemetry-go-compile-instrumentation/pkg/inst/insttest"
)

func setupTestTracer(t *testing.T) (*tracetest.SpanRecorder, *sdktrace.TracerProvider) {
	t.Helper()
	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))
	otel.SetTracerProvider(tp)
	initOnce = sync.Once{}
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })
	return sr, tp
}

func newHTTPRequest(t *testing.T, method, urlStr string) *http.Request {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), method, urlStr, nil)
	require.NoError(t, err)
	return req
}

func TestParseEsRequest(t *testing.T) {
	tests := []struct {
		name          string
		method        string
		url           string
		wantOp        string
		wantPath      string
		wantNamespace string
	}{
		{
			name:          "index create",
			method:        "PUT",
			url:           "http://localhost:9200/my_index",
			wantOp:        "put",
			wantPath:      "/my_index",
			wantNamespace: "my_index",
		},
		{
			name:          "index delete",
			method:        "DELETE",
			url:           "http://localhost:9200/my_index",
			wantOp:        "delete",
			wantPath:      "/my_index",
			wantNamespace: "my_index",
		},
		{
			name:          "doc index",
			method:        "POST",
			url:           "http://localhost:9200/my_index/_doc",
			wantOp:        "_doc",
			wantPath:      "/my_index/_doc",
			wantNamespace: "my_index",
		},
		{
			name:          "doc get",
			method:        "GET",
			url:           "http://localhost:9200/my_index/_doc/id",
			wantOp:        "_doc",
			wantPath:      "/my_index/_doc/id",
			wantNamespace: "my_index",
		},
		{
			name:          "search",
			method:        "GET",
			url:           "http://localhost:9200/my_index/_search",
			wantOp:        "_search",
			wantPath:      "/my_index/_search",
			wantNamespace: "my_index",
		},
		{
			name:          "update",
			method:        "POST",
			url:           "http://localhost:9200/my_index/_update/id",
			wantOp:        "_update",
			wantPath:      "/my_index/_update/id",
			wantNamespace: "my_index",
		},
		{
			name:          "cluster health (no index)",
			method:        "GET",
			url:           "http://localhost:9200/_cluster/health",
			wantOp:        "health",
			wantPath:      "/_cluster/health",
			wantNamespace: "",
		},
		{
			name:          "nil request",
			method:        "",
			url:           "",
			wantOp:        "UNKNOWN",
			wantPath:      "",
			wantNamespace: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.url != "" {
				req = newHTTPRequest(t, tt.method, tt.url)
			}
			op, path, namespace := parseEsRequest(req)
			assert.Equal(t, tt.wantOp, op)
			assert.Equal(t, tt.wantPath, path)
			assert.Equal(t, tt.wantNamespace, namespace)
		})
	}
}

func TestBeforePerform_Disabled(t *testing.T) {
	t.Setenv("OTEL_GO_DISABLED_INSTRUMENTATIONS", "ELASTICSEARCH")
	sr, _ := setupTestTracer(t)

	req := newHTTPRequest(t, "GET", "http://localhost:9200/my_index/_search")
	ictx := insttest.NewMockHookContext()

	// BeforePerform with nil client — disabled path returns early
	BeforePerform(ictx, nil, req)

	assert.Empty(t, sr.Ended(), "no spans should be created when instrumentation is disabled")
	assert.Nil(t, ictx.GetKeyData("span"), "span key must not be set when disabled")
}

func TestBeforeAndAfterPerform_CreatesSpan(t *testing.T) {
	t.Setenv("OTEL_GO_ENABLED_INSTRUMENTATIONS", "ELASTICSEARCH")
	sr, _ := setupTestTracer(t)

	req := newHTTPRequest(t, "GET", "http://localhost:9200/my_index/_search")
	ictx := insttest.NewMockHookContext()

	// BeforePerform with nil client — no panic; addresses will be empty
	BeforePerform(ictx, nil, req)

	// Span is started but not yet ended
	assert.Empty(t, sr.Ended())
	span := ictx.GetKeyData("span")
	require.NotNil(t, span, "span key must be set by BeforePerform")

	// AfterPerform ends the span
	AfterPerform(ictx, nil, nil)

	spans := sr.Ended()
	require.Len(t, spans, 1)
	assert.Equal(t, "_search", spans[0].Name())
	assert.Equal(t, "client", spans[0].SpanKind().String())

	attrMap := make(map[string]interface{})
	for _, a := range spans[0].Attributes() {
		attrMap[string(a.Key)] = a.Value.AsInterface()
	}
	assert.Equal(t, "elasticsearch", attrMap["db.system.name"])
	assert.Equal(t, "_search", attrMap["db.operation.name"])
	assert.Equal(t, "/my_index/_search", attrMap["db.query.text"])
	assert.Equal(t, "my_index", attrMap["db.namespace"])
}

func TestAfterPerform_RecordsError(t *testing.T) {
	t.Setenv("OTEL_GO_ENABLED_INSTRUMENTATIONS", "ELASTICSEARCH")
	sr, _ := setupTestTracer(t)

	req := newHTTPRequest(t, "GET", "http://localhost:9200/my_index/_doc/id")
	ictx := insttest.NewMockHookContext()

	BeforePerform(ictx, nil, req)

	expectedErr := errors.New("connection refused")
	AfterPerform(ictx, nil, expectedErr)

	spans := sr.Ended()
	require.Len(t, spans, 1)
	assert.Equal(t, codes.Error, spans[0].Status().Code)
	assert.Contains(t, spans[0].Status().Description, "connection refused")
}

func TestAfterPerform_NoSpan_NoPanic(t *testing.T) {
	ictx := insttest.NewMockHookContext()
	// AfterPerform with no span stored must not panic
	assert.NotPanics(t, func() {
		AfterPerform(ictx, nil, nil)
	})
}

func TestElasticsearchEnabler(t *testing.T) {
	tests := []struct {
		name     string
		setupEnv func(t *testing.T)
		expected bool
	}{
		{
			name: "enabled explicitly",
			setupEnv: func(t *testing.T) {
				t.Setenv("OTEL_GO_ENABLED_INSTRUMENTATIONS", "ELASTICSEARCH")
			},
			expected: true,
		},
		{
			name: "disabled explicitly",
			setupEnv: func(t *testing.T) {
				t.Setenv("OTEL_GO_DISABLED_INSTRUMENTATIONS", "ELASTICSEARCH")
			},
			expected: false,
		},
		{
			name: "default enabled when no env set",
			setupEnv: func(t *testing.T) {
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupEnv(t)
			e := elasticsearchEnabler{}
			assert.Equal(t, tt.expected, e.Enable())
		})
	}
}

func TestModuleVersion(t *testing.T) {
	v := moduleVersion()
	assert.NotEmpty(t, v)
}
