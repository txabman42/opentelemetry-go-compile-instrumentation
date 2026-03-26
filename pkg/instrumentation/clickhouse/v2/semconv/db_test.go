// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package semconv

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClickHouseClientTraceAttrs(t *testing.T) {
	tests := []struct {
		name     string
		req      ClickHouseRequest
		expected map[string]interface{}
	}{
		{
			name: "exec with host:port address",
			req: ClickHouseRequest{
				Op:        "EXEC",
				Statement: "CREATE TABLE IF NOT EXISTS users (id String)",
				DbName:    "default",
				Addr:      "127.0.0.1:9000",
			},
			expected: map[string]interface{}{
				"db.system.name":    "clickhouse",
				"db.operation.name": "EXEC",
				"db.query.text":     "CREATE TABLE IF NOT EXISTS users (id String)",
				"db.namespace":      "default",
				"server.address":    "127.0.0.1",
				"server.port":       int64(9000),
				"network.transport": "tcp",
			},
		},
		{
			name: "select with multi-host comma address (no port split)",
			req: ClickHouseRequest{
				Op:        "SELECT",
				Statement: "SELECT * FROM users WHERE id = ?",
				DbName:    "mydb",
				Addr:      "10.0.0.1:9000,10.0.0.2:9000",
				Params:    []any{"1"},
			},
			expected: map[string]interface{}{
				"db.system.name":    "clickhouse",
				"db.operation.name": "SELECT",
				"db.query.text":     "SELECT * FROM users WHERE id = ?",
				"db.namespace":      "mydb",
				"server.address":    "10.0.0.1:9000,10.0.0.2:9000",
				"network.transport": "tcp",
			},
		},
		{
			name: "ping operation",
			req: ClickHouseRequest{
				Op:        "PING",
				Statement: "PING",
				DbName:    "default",
				Addr:      "localhost:9000",
			},
			expected: map[string]interface{}{
				"db.system.name":    "clickhouse",
				"db.operation.name": "PING",
				"db.query.text":     "PING",
				"db.namespace":      "default",
				"server.address":    "localhost",
				"server.port":       int64(9000),
				"network.transport": "tcp",
			},
		},
		{
			name: "async insert",
			req: ClickHouseRequest{
				Op:        "ASYNC_INSERT",
				Statement: "INSERT INTO users (id, name, age) VALUES (?, ?, ?)",
				DbName:    "default",
				Addr:      "127.0.0.1:9000",
				Params:    []any{"1", "Alice", 30},
			},
			expected: map[string]interface{}{
				"db.system.name":    "clickhouse",
				"db.operation.name": "ASYNC_INSERT",
				"db.query.text":     "INSERT INTO users (id, name, age) VALUES (?, ?, ?)",
				"db.namespace":      "default",
				"server.address":    "127.0.0.1",
				"server.port":       int64(9000),
				"network.transport": "tcp",
			},
		},
		{
			name: "server version operation",
			req: ClickHouseRequest{
				Op:        "SERVER_VERSION",
				Statement: "SERVER_VERSION",
				DbName:    "",
				Addr:      "127.0.0.1:9000",
			},
			expected: map[string]interface{}{
				"db.system.name":    "clickhouse",
				"db.operation.name": "SERVER_VERSION",
				"db.query.text":     "SERVER_VERSION",
				"db.namespace":      "",
				"server.address":    "127.0.0.1",
				"server.port":       int64(9000),
				"network.transport": "tcp",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attrs := ClickHouseClientTraceAttrs(tt.req)
			require.NotNil(t, attrs)
			assert.Greater(t, len(attrs), 0, "should return attributes")

			attrMap := make(map[string]interface{})
			for _, attr := range attrs {
				attrMap[string(attr.Key)] = attr.Value.AsInterface()
			}

			for key, expectedVal := range tt.expected {
				actualVal, ok := attrMap[key]
				require.True(t, ok, "expected attribute %q not found, got: %v", key, attrMap)
				assert.Equal(t, expectedVal, actualVal, "attribute %q value mismatch", key)
			}
		})
	}
}
