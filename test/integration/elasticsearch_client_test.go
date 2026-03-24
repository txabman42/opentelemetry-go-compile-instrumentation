// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:build integration

package test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/open-telemetry/opentelemetry-go-compile-instrumentation/test/testutil"
)

const (
	elasticsearchImage = "docker.elastic.co/elasticsearch/elasticsearch:8.4.0"
	elasticsearchPort  = "9200/tcp"
)

// startElasticsearchContainer starts a throwaway Elasticsearch server and returns its base URL.
func startElasticsearchContainer(t *testing.T) string {
	t.Helper()

	req := testcontainers.ContainerRequest{
		Image:        elasticsearchImage,
		ExposedPorts: []string{elasticsearchPort},
		Env: map[string]string{
			"discovery.type":         "single-node",
			"xpack.security.enabled": "false",
			"ES_JAVA_OPTS":           "-Xms512m -Xmx512m",
		},
		WaitingFor: wait.ForHTTP("/_cluster/health").
			WithPort("9200/tcp").
			WithStatusCodeMatcher(func(status int) bool { return status == 200 }).
			WithStartupTimeout(3 * time.Minute),
	}
	c, err := testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = c.Terminate(context.Background()) })

	host, err := c.Host(context.Background())
	require.NoError(t, err)
	port, err := c.MappedPort(context.Background(), "9200")
	require.NoError(t, err)

	return fmt.Sprintf("http://%s:%s", host, port.Port())
}

func TestElasticsearchClient(t *testing.T) {
	testcontainers.SkipIfProviderIsNotHealthy(t)

	versions := []struct {
		name string
		app  string
	}{
		{name: "v8.4.0 (functional client)", app: "elasticsearchclient/v8.4.0"},
		{name: "v8.5.0 (typed client)", app: "elasticsearchclient/v8.5.0"},
	}

	for _, ver := range versions {
		t.Run(ver.name, func(t *testing.T) {
			addr := startElasticsearchContainer(t)
			f := testutil.NewTestFixture(t)

			f.BuildAndRun(ver.app, "-addr="+addr)

			spans := testutil.AllSpans(f.Traces())
			require.GreaterOrEqual(t, len(spans), 4, "expected at least 4 spans (create index, index doc, search, delete index)")

			serverAddress := parseElasticsearchHost(addr)

			createSpan := testutil.RequireSpan(t, f.Traces(),
				testutil.IsClient,
				testutil.HasAttribute("db.operation.name", "put"),
			)
			testutil.RequireElasticsearchClientSemconv(t, createSpan,
				"put", "/my_index", serverAddress)

			searchSpan := testutil.RequireSpan(t, f.Traces(),
				testutil.IsClient,
				testutil.HasAttribute("db.operation.name", "_search"),
			)
			testutil.RequireElasticsearchClientSemconv(t, searchSpan,
				"_search", "/my_index/_search", serverAddress)

			deleteSpan := testutil.RequireSpan(t, f.Traces(),
				testutil.IsClient,
				testutil.HasAttribute("db.operation.name", "delete"),
			)
			testutil.RequireElasticsearchClientSemconv(t, deleteSpan,
				"delete", "/my_index", serverAddress)
		})
	}
}

// parseElasticsearchHost extracts the host (without port) from a URL like
// "http://127.0.0.1:9200".
func parseElasticsearchHost(baseURL string) string {
	s := baseURL
	if len(s) > 7 && s[:7] == "http://" {
		s = s[7:]
	}
	for i, c := range s {
		if c == ':' {
			return s[:i]
		}
	}
	return s
}
