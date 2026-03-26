// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package clickhousev2

import (
	"context"
	"errors"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

var errFake = errors.New("fake error")

// fakeConn is a minimal driver.Conn implementation for tests.
// Methods are no-ops that return zero values.
type fakeConn struct{}

func (f *fakeConn) Contributors() []string                        { return nil }
func (f *fakeConn) ServerVersion() (*driver.ServerVersion, error) { return nil, nil }
func (f *fakeConn) Select(_ context.Context, _ any, _ string, _ ...any) error {
	return nil
}
func (f *fakeConn) Query(_ context.Context, _ string, _ ...any) (driver.Rows, error) {
	return nil, nil
}
func (f *fakeConn) QueryRow(_ context.Context, _ string, _ ...any) driver.Row {
	return &fakeRow{}
}
func (f *fakeConn) PrepareBatch(_ context.Context, _ string, _ ...driver.PrepareBatchOption) (driver.Batch, error) {
	return nil, nil
}
func (f *fakeConn) Exec(_ context.Context, _ string, _ ...any) error    { return nil }
func (f *fakeConn) AsyncInsert(_ context.Context, _ string, _ bool, _ ...any) error {
	return nil
}
func (f *fakeConn) Ping(_ context.Context) error { return nil }
func (f *fakeConn) Stats() driver.Stats          { return driver.Stats{} }
func (f *fakeConn) Close() error                 { return nil }

type fakeRow struct{}

func (r *fakeRow) Err() error                       { return nil }
func (r *fakeRow) Scan(_ ...any) error              { return nil }
func (r *fakeRow) ScanStruct(_ any) error           { return nil }
func (r *fakeRow) ColumnTypes() []driver.ColumnType { return nil }
