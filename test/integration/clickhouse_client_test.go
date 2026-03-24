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
	clickhouseImage = "docker.io/clickhouse/clickhouse-server:24.3.4.147"
	clickhousePort  = "9000/tcp"
)

// clickhouseVersions lists every library version that must be tested.
// Corresponds to test/apps/clickhouseclient/<version>/.
var clickhouseVersions = []string{"v2.13.0", "v2.42.0"}

// startClickHouseContainer starts a throwaway ClickHouse server and returns its host:port.
func startClickHouseContainer(t *testing.T) string {
	t.Helper()

	req := testcontainers.ContainerRequest{
		Image:        clickhouseImage,
		ExposedPorts: []string{"8123/tcp", "9000/tcp"},
		WaitingFor:   wait.ForListeningPort("9000/tcp").WithStartupTimeout(2 * time.Minute),
	}
	c, err := testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = c.Terminate(context.Background()) })

	host, err := c.Host(context.Background())
	require.NoError(t, err)
	port, err := c.MappedPort(context.Background(), "9000")
	require.NoError(t, err)

	return fmt.Sprintf("%s:%s", host, port.Port())
}

func TestClickHouseClientCRUD(t *testing.T) {
	for _, ver := range clickhouseVersions {
		t.Run(ver, func(t *testing.T) {
			addr := startClickHouseContainer(t)
			f := testutil.NewTestFixture(t)

			f.BuildAndRun("clickhouseclient/"+ver, "-addr="+addr, "-op=all")

			spans := testutil.AllSpans(f.Traces())
			// all: exec + async_insert + select + query + queryrow + batch + ping = 7 spans
			require.GreaterOrEqual(t, len(spans), 7, "expected at least 7 spans for all operations")

			serverHost, serverPort := splitHostPort(t, addr)

			execSpan := testutil.RequireSpan(t, f.Traces(),
				testutil.IsClient,
				testutil.HasAttribute("db.operation.name", "EXEC"),
			)
			testutil.RequireClickHouseClientSemconv(t, execSpan,
				"EXEC", createTableSQL, serverHost, serverPort, "default")

			insertSpan := testutil.RequireSpan(t, f.Traces(),
				testutil.IsClient,
				testutil.HasAttribute("db.operation.name", "ASYNC_INSERT"),
			)
			testutil.RequireClickHouseClientSemconv(t, insertSpan,
				"ASYNC_INSERT",
				"INSERT INTO otel_test_users (id, name, age) VALUES (?, ?, ?)",
				serverHost, serverPort, "default",
			)

			selectSpan := testutil.RequireSpan(t, f.Traces(),
				testutil.IsClient,
				testutil.HasAttribute("db.operation.name", "SELECT"),
			)
			testutil.RequireClickHouseClientSemconv(t, selectSpan,
				"SELECT",
				"SELECT * FROM otel_test_users WHERE id = ?",
				serverHost, serverPort, "default",
			)

			querySpan := testutil.RequireSpan(t, f.Traces(),
				testutil.IsClient,
				testutil.HasAttribute("db.operation.name", "QUERY"),
			)
			testutil.RequireClickHouseClientSemconv(t, querySpan,
				"QUERY",
				"SELECT * FROM otel_test_users WHERE id = ?",
				serverHost, serverPort, "default",
			)

			pingSpan := testutil.RequireSpan(t, f.Traces(),
				testutil.IsClient,
				testutil.HasAttribute("db.operation.name", "PING"),
			)
			testutil.RequireClickHouseClientSemconv(t, pingSpan,
				"PING", "PING", serverHost, serverPort, "default")
		})
	}
}

func TestClickHouseClientPing(t *testing.T) {
	for _, ver := range clickhouseVersions {
		t.Run(ver, func(t *testing.T) {
			addr := startClickHouseContainer(t)
			f := testutil.NewTestFixture(t)

			f.BuildAndRun("clickhouseclient/"+ver, "-addr="+addr, "-op=ping")

			span := f.RequireSingleSpan()
			require.Equal(t, "PING", span.Name())

			serverHost, serverPort := splitHostPort(t, addr)
			testutil.RequireClickHouseClientSemconv(t, span,
				"PING", "PING", serverHost, serverPort, "default")
		})
	}
}

func TestClickHouseClientExec(t *testing.T) {
	for _, ver := range clickhouseVersions {
		t.Run(ver, func(t *testing.T) {
			addr := startClickHouseContainer(t)
			f := testutil.NewTestFixture(t)

			f.BuildAndRun("clickhouseclient/"+ver, "-addr="+addr, "-op=exec")

			span := f.RequireSingleSpan()
			require.Equal(t, "EXEC", span.Name())

			serverHost, serverPort := splitHostPort(t, addr)
			testutil.RequireClickHouseClientSemconv(t, span,
				"EXEC", createTableSQL, serverHost, serverPort, "default")
		})
	}
}

// createTableSQL is the statement used in the exec operation.
const createTableSQL = `CREATE TABLE IF NOT EXISTS otel_test_users (
		id String,
		name String,
		age Int32
	) ENGINE = MergeTree() ORDER BY id`

// splitHostPort splits an "addr" string into host and port for semconv assertions.
func splitHostPort(t *testing.T, addr string) (string, int64) {
	t.Helper()
	var host string
	var port int
	_, err := fmt.Sscanf(addr, "%s", &addr)
	require.NoError(t, err)
	n, err := fmt.Sscanf(addr, "%[^:]:%d", &host, &port)
	require.NoError(t, err)
	require.Equal(t, 2, n, "expected host:port in address %q", addr)
	return host, int64(port)
}
