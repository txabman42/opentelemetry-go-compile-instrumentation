// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package main provides a minimal ClickHouse client for integration testing.
// It connects to a real ClickHouse server and performs CRUD operations.
// This client is designed to be instrumented with the otelc compile-time tool.
package main

import (
	"context"
	"flag"
	"log"
	"log/slog"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

var (
	addr = flag.String("addr", "127.0.0.1:9000", "ClickHouse server address (host:port)")
	op   = flag.String("op", "all", "Operation to perform: exec, insert, select, query, queryrow, batch, ping, all")
)

const (
	createTable = `CREATE TABLE IF NOT EXISTS otel_test_users (
		id String,
		name String,
		age Int32
	) ENGINE = MergeTree() ORDER BY id`

	insertSQL = `INSERT INTO otel_test_users (id, name, age) VALUES (?, ?, ?)`
)

func main() {
	flag.Parse()

	con, err := clickhouse.Open(&clickhouse.Options{
		Addr:     []string{*addr},
		Protocol: clickhouse.Native,
	})
	if err != nil {
		log.Fatalf("failed to open clickhouse connection: %v", err)
	}
	defer con.Close()

	ctx := context.Background()

	switch *op {
	case "exec":
		doExec(ctx, con)
	case "insert":
		doAsyncInsert(ctx, con)
	case "select":
		doSelect(ctx, con)
	case "query":
		doQuery(ctx, con)
	case "queryrow":
		doQueryRow(ctx, con)
	case "batch":
		doBatch(ctx, con)
	case "ping":
		doPing(ctx, con)
	case "all":
		doExec(ctx, con)
		doAsyncInsert(ctx, con)
		doSelect(ctx, con)
		doQuery(ctx, con)
		doQueryRow(ctx, con)
		doBatch(ctx, con)
		doPing(ctx, con)
	default:
		log.Fatalf("unknown operation: %s", *op)
	}

	slog.Info("clickhouse operations completed successfully")
}

func doExec(ctx context.Context, con driver.Conn) {
	if err := con.Exec(ctx, createTable); err != nil {
		log.Fatalf("exec failed: %v", err)
	}
	slog.Info("exec succeeded", "op", "CREATE TABLE")
}

func doAsyncInsert(ctx context.Context, con driver.Conn) {
	if err := con.AsyncInsert(ctx, insertSQL, false, "1", "Alice", int32(30)); err != nil {
		log.Fatalf("async insert failed: %v", err)
	}
	slog.Info("async insert succeeded")
}

func doSelect(ctx context.Context, con driver.Conn) {
	type User struct {
		ID   string `ch:"id"`
		Name string `ch:"name"`
		Age  int32  `ch:"age"`
	}
	var users []User
	if err := con.Select(ctx, &users, `SELECT * FROM otel_test_users WHERE id = ?`, "1"); err != nil {
		log.Fatalf("select failed: %v", err)
	}
	slog.Info("select succeeded", "count", len(users))
}

func doQuery(ctx context.Context, con driver.Conn) {
	rows, err := con.Query(ctx, `SELECT * FROM otel_test_users WHERE id = ?`, "1")
	if err != nil {
		log.Fatalf("query failed: %v", err)
	}
	defer rows.Close()
	slog.Info("query succeeded")
}

func doQueryRow(ctx context.Context, con driver.Conn) {
	row := con.QueryRow(ctx, `SELECT * FROM otel_test_users WHERE id = ?`, "1")
	if row.Err() != nil {
		log.Fatalf("queryrow failed: %v", row.Err())
	}
	slog.Info("queryrow succeeded")
}

func doBatch(ctx context.Context, con driver.Conn) {
	batch, err := con.PrepareBatch(ctx, `INSERT INTO otel_test_users (id, name, age)`)
	if err != nil {
		log.Fatalf("prepare batch failed: %v", err)
	}
	if err := batch.Append("2", "Bob", int32(25)); err != nil {
		log.Fatalf("batch append failed: %v", err)
	}
	if err := batch.Send(); err != nil {
		log.Fatalf("batch send failed: %v", err)
	}
	slog.Info("batch insert succeeded")
}

func doPing(ctx context.Context, con driver.Conn) {
	if err := con.Ping(ctx); err != nil {
		log.Fatalf("ping failed: %v", err)
	}
	slog.Info("ping succeeded")
}
