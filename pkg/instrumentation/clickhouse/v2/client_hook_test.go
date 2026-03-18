// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package clickhousev2

import (
	"testing"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/open-telemetry/opentelemetry-go-compile-instrumentation/pkg/inst/insttest"
)

func TestBeforeOpen_StoresOptions(t *testing.T) {
	opts := &clickhouse.Options{
		Addr: []string{"127.0.0.1:9000"},
		Auth: clickhouse.Auth{Database: "testdb", Username: "default"},
	}
	ictx := insttest.NewMockHookContext()

	BeforeOpen(ictx, opts)

	got, ok := ictx.GetData().(*clickhouse.Options)
	require.True(t, ok, "expected *clickhouse.Options stored in hook context")
	assert.Equal(t, opts, got)
}

func TestBeforeOpen_DisabledDoesNotStore(t *testing.T) {
	t.Setenv("OTEL_GO_DISABLED_INSTRUMENTATIONS", "CLICKHOUSE")
	opts := &clickhouse.Options{
		Addr: []string{"127.0.0.1:9000"},
	}
	ictx := insttest.NewMockHookContext()

	BeforeOpen(ictx, opts)

	// When disabled, no data should be stored.
	assert.Nil(t, ictx.GetData(), "expected no data stored when disabled")
}

func TestAfterOpen_NilConnIsIgnored(t *testing.T) {
	ictx := insttest.NewMockHookContext()

	// Should not panic when conn is nil.
	AfterOpen(ictx, nil, nil)

	require.Equal(t, 0, ictx.GetReturnValCount(), "no return val should be set for nil conn")
}

func TestAfterOpen_ErrorIsIgnored(t *testing.T) {
	opts := &clickhouse.Options{
		Addr: []string{"127.0.0.1:9000"},
	}
	ictx := insttest.NewMockHookContext()
	BeforeOpen(ictx, opts)

	// When err != nil the hook should bail without wrapping.
	AfterOpen(ictx, &fakeConn{}, errFake)

	require.Equal(t, 0, ictx.GetReturnValCount(), "no return val should be set on error")
}

func TestOtelConn_AddrSingleHost(t *testing.T) {
	opts := &clickhouse.Options{
		Addr: []string{"127.0.0.1:9000"},
		Auth: clickhouse.Auth{Database: "mydb"},
	}
	oc := newOtelConn(&fakeConn{}, opts)
	assert.Equal(t, "127.0.0.1:9000", oc.addr())
}

func TestOtelConn_AddrMultipleHosts(t *testing.T) {
	opts := &clickhouse.Options{
		Addr: []string{"host1:9000", "host2:9000"},
	}
	oc := newOtelConn(&fakeConn{}, opts)
	assert.Equal(t, "host1:9000,host2:9000", oc.addr())
}
