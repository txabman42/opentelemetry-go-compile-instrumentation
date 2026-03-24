// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package semconv

import (
	"net"
	"strconv"

	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

// ClickHouseRequest holds metadata captured at the time of a ClickHouse operation.
type ClickHouseRequest struct {
	Op        string
	Statement string
	DbName    string
	Addr      string
	Params    []any
}

// ClickHouseClientTraceAttrs returns the DB semantic-convention attributes for a ClickHouse operation.
func ClickHouseClientTraceAttrs(req ClickHouseRequest) []attribute.KeyValue {
	host, portStr, err := net.SplitHostPort(req.Addr)
	if err != nil {
		host = req.Addr
	}

	attrs := []attribute.KeyValue{
		semconv.DBSystemNameKey.String("clickhouse"),
		semconv.DBOperationName(req.Op),
		semconv.DBQueryText(req.Statement),
		semconv.DBNamespace(req.DbName),
		semconv.ServerAddress(host),
		semconv.NetworkTransportTCP,
	}

	if err == nil {
		if port, convErr := strconv.Atoi(portStr); convErr == nil && port > 0 {
			attrs = append(attrs, semconv.ServerPort(port))
		}
	}

	return attrs
}
