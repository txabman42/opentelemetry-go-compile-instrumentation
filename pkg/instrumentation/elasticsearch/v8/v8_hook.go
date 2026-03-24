// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package v8

import (
	"net/http"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/elastic/elastic-transport-go/v8/elastictransport"
	elasticsearch "github.com/elastic/go-elasticsearch/v8"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/open-telemetry/opentelemetry-go-compile-instrumentation/pkg/inst"
	"github.com/open-telemetry/opentelemetry-go-compile-instrumentation/pkg/instrumentation/shared"
)

const (
	instrumentationName = "github.com/open-telemetry/opentelemetry-go-compile-instrumentation/pkg/instrumentation/elasticsearch/v8"
	instrumentationKey  = "ELASTICSEARCH"
)

var (
	logger   = shared.Logger()
	tracer   trace.Tracer
	initOnce sync.Once
)

type elasticsearchEnabler struct{}

func (e elasticsearchEnabler) Enable() bool { return shared.Instrumented(instrumentationKey) }

var enabler = elasticsearchEnabler{}

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
		if err := shared.SetupOTelSDK("go.opentelemetry.io/compile-instrumentation/elasticsearch/v8", version); err != nil {
			logger.Error("failed to setup OTel SDK", "error", err)
		}
		tracer = otel.GetTracerProvider().Tracer(instrumentationName, trace.WithInstrumentationVersion(version))
		if err := shared.StartRuntimeMetrics(); err != nil {
			logger.Error("failed to start runtime metrics", "error", err)
		}
	})
}

// BeforePerform is called before (*BaseClient).Perform and starts a DB client span.
func BeforePerform(ictx inst.HookContext, client *elasticsearch.BaseClient, request *http.Request) {
	if !enabler.Enable() {
		return
	}
	initInstrumentation()

	var addresses []string
	if client != nil {
		if t, ok := client.Transport.(*elastictransport.Client); ok {
			for _, u := range t.URLs() {
				addresses = append(addresses, u.Hostname())
			}
		}
	}
	serverAddress := strings.Join(addresses, ",")

	op, urlPath, indexName := parseEsRequest(request)
	attrs := []attribute.KeyValue{
		semconv.DBSystemNameKey.String("elasticsearch"),
		semconv.DBOperationNameKey.String(op),
		semconv.DBQueryTextKey.String(urlPath),
		semconv.ServerAddressKey.String(serverAddress),
	}
	if indexName != "" {
		attrs = append(attrs, semconv.DBNamespaceKey.String(indexName))
	}

	parentCtx := request.Context()
	ctx, span := tracer.Start(parentCtx, op,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(attrs...),
	)
	ictx.SetKeyData("ctx", ctx)
	ictx.SetKeyData("span", span)
}

// AfterPerform is called after (*BaseClient).Perform and ends the DB client span.
func AfterPerform(ictx inst.HookContext, response *http.Response, err error) {
	span, ok := ictx.GetKeyData("span").(trace.Span)
	if !ok || span == nil {
		return
	}
	defer span.End()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}

// parseEsRequest extracts the operation name, URL path, and index name from an
// HTTP request to Elasticsearch. The URL structure is:
//
//	/<index>/<sub-resource>/<id>
//
// The operation is the sub-resource segment (e.g. _doc, _search, _update) when
// present, or the lower-cased HTTP method for top-level index operations (e.g.
// PUT /my_index, DELETE /my_index). The index name is returned as the namespace;
// it is empty for cluster-level endpoints (e.g. /_cluster/health).
func parseEsRequest(req *http.Request) (op, urlPath, indexName string) {
	if req == nil || req.URL == nil {
		return "UNKNOWN", "", ""
	}
	urlPath = req.URL.Path
	parts := strings.Split(strings.TrimPrefix(urlPath, "/"), "/")
	// parts[0] = index name (or system endpoint starting with _), parts[1] = sub-resource
	switch len(parts) {
	case 0:
		return strings.ToLower(req.Method), urlPath, ""
	case 1:
		index := parts[0]
		if strings.HasPrefix(index, "_") {
			index = "" // cluster-level endpoint, no index namespace
		}
		return strings.ToLower(req.Method), urlPath, index
	default:
		index := parts[0]
		if strings.HasPrefix(index, "_") {
			index = ""
		}
		if parts[1] != "" {
			return parts[1], urlPath, index
		}
		return strings.ToLower(req.Method), urlPath, index
	}
}
