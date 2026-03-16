// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

func makeSpan(name string, kind ptrace.SpanKind) ptrace.Span {
	s := ptrace.NewSpan()
	s.SetName(name)
	s.SetKind(kind)
	return s
}

func TestIsClient(t *testing.T) {
	tests := []struct {
		name     string
		kind     ptrace.SpanKind
		expected bool
	}{
		{"client", ptrace.SpanKindClient, true},
		{"server", ptrace.SpanKindServer, false},
		{"producer", ptrace.SpanKindProducer, false},
		{"consumer", ptrace.SpanKindConsumer, false},
		{"internal", ptrace.SpanKindInternal, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := makeSpan("test", tt.kind)
			assert.Equal(t, tt.expected, IsClient(s))
		})
	}
}

func TestIsServer(t *testing.T) {
	tests := []struct {
		name     string
		kind     ptrace.SpanKind
		expected bool
	}{
		{"server", ptrace.SpanKindServer, true},
		{"client", ptrace.SpanKindClient, false},
		{"producer", ptrace.SpanKindProducer, false},
		{"consumer", ptrace.SpanKindConsumer, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := makeSpan("test", tt.kind)
			assert.Equal(t, tt.expected, IsServer(s))
		})
	}
}

func TestIsProducer(t *testing.T) {
	tests := []struct {
		name     string
		kind     ptrace.SpanKind
		expected bool
	}{
		{"producer", ptrace.SpanKindProducer, true},
		{"consumer", ptrace.SpanKindConsumer, false},
		{"client", ptrace.SpanKindClient, false},
		{"server", ptrace.SpanKindServer, false},
		{"internal", ptrace.SpanKindInternal, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := makeSpan("test", tt.kind)
			assert.Equal(t, tt.expected, IsProducer(s))
		})
	}
}

func TestIsConsumer(t *testing.T) {
	tests := []struct {
		name     string
		kind     ptrace.SpanKind
		expected bool
	}{
		{"consumer", ptrace.SpanKindConsumer, true},
		{"producer", ptrace.SpanKindProducer, false},
		{"client", ptrace.SpanKindClient, false},
		{"server", ptrace.SpanKindServer, false},
		{"internal", ptrace.SpanKindInternal, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := makeSpan("test", tt.kind)
			assert.Equal(t, tt.expected, IsConsumer(s))
		})
	}
}

func TestHasSpanName(t *testing.T) {
	tests := []struct {
		spanName    string
		matchName   string
		expected    bool
	}{
		{"GET /foo", "GET /foo", true},
		{"GET /foo", "POST /foo", false},
		{"", "", true},
		{"some-span", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.spanName+"_vs_"+tt.matchName, func(t *testing.T) {
			s := makeSpan(tt.spanName, ptrace.SpanKindInternal)
			matcher := HasSpanName(tt.matchName)
			assert.Equal(t, tt.expected, matcher(s))
		})
	}
}

func TestHasStatusCode(t *testing.T) {
	tests := []struct {
		name       string
		setCode    ptrace.StatusCode
		matchCode  ptrace.StatusCode
		expected   bool
	}{
		{"unset matches unset", ptrace.StatusCodeUnset, ptrace.StatusCodeUnset, true},
		{"ok matches ok", ptrace.StatusCodeOk, ptrace.StatusCodeOk, true},
		{"error matches error", ptrace.StatusCodeError, ptrace.StatusCodeError, true},
		{"ok does not match error", ptrace.StatusCodeOk, ptrace.StatusCodeError, false},
		{"error does not match unset", ptrace.StatusCodeError, ptrace.StatusCodeUnset, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := ptrace.NewSpan()
			s.Status().SetCode(tt.setCode)
			matcher := HasStatusCode(tt.matchCode)
			assert.Equal(t, tt.expected, matcher(s))
		})
	}
}

func TestHasAttribute(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    any
		expected bool
	}{
		{"string match", "http.method", "GET", true},
		{"string mismatch", "http.method", "POST", false},
		{"int match", "http.status_code", int64(200), true},
		{"bool match", "error", true, true},
		{"missing key", "nonexistent", "val", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := ptrace.NewSpan()
			s.Attributes().PutStr("http.method", "GET")
			s.Attributes().PutInt("http.status_code", 200)
			s.Attributes().PutBool("error", true)
			matcher := HasAttribute(tt.key, tt.value)
			assert.Equal(t, tt.expected, matcher(s))
		})
	}
}

func TestHasAttributeContaining(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		substr   string
		expected bool
	}{
		{"contains prefix", "db.statement", "SELECT", true},
		{"contains suffix", "db.statement", "users", true},
		{"no match", "db.statement", "DELETE", false},
		{"missing key", "nonexistent", "val", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := ptrace.NewSpan()
			s.Attributes().PutStr("db.statement", "SELECT * FROM users")
			matcher := HasAttributeContaining(tt.key, tt.substr)
			assert.Equal(t, tt.expected, matcher(s))
		})
	}
}

func TestRequireSpan(t *testing.T) {
	td := ptrace.NewTraces()
	rs := td.ResourceSpans().AppendEmpty()
	ss := rs.ScopeSpans().AppendEmpty()

	s1 := ss.Spans().AppendEmpty()
	s1.SetName("client-span")
	s1.SetKind(ptrace.SpanKindClient)
	s1.Attributes().PutStr("http.method", "GET")

	s2 := ss.Spans().AppendEmpty()
	s2.SetName("server-span")
	s2.SetKind(ptrace.SpanKindServer)
	s2.Attributes().PutStr("http.method", "POST")

	t.Run("find client span by kind", func(t *testing.T) {
		found := RequireSpan(t, td, IsClient)
		assert.Equal(t, "client-span", found.Name())
	})

	t.Run("find server span by kind", func(t *testing.T) {
		found := RequireSpan(t, td, IsServer)
		assert.Equal(t, "server-span", found.Name())
	})

	t.Run("find span by name and attribute", func(t *testing.T) {
		found := RequireSpan(t, td, HasSpanName("server-span"), HasAttribute("http.method", "POST"))
		assert.Equal(t, "server-span", found.Name())
	})
}

func TestRequireAttribute(t *testing.T) {
	s := ptrace.NewSpan()
	s.SetName("test-span")
	s.Attributes().PutStr("http.method", "GET")
	s.Attributes().PutInt("http.status_code", 200)

	t.Run("string attribute matches", func(t *testing.T) {
		RequireAttribute(t, s, "http.method", "GET")
	})

	t.Run("int attribute matches", func(t *testing.T) {
		RequireAttribute(t, s, "http.status_code", int64(200))
	})
}

func TestRequireAttributeExists(t *testing.T) {
	s := ptrace.NewSpan()
	s.SetName("test-span")
	s.Attributes().PutStr("http.method", "GET")

	t.Run("existing attribute passes", func(t *testing.T) {
		RequireAttributeExists(t, s, "http.method")
	})
}

func TestRequireAttributeExistsHelper(t *testing.T) {
	s := ptrace.NewSpan()
	s.Attributes().PutStr("key", "value")
	require.NotNil(t, s)
	RequireAttributeExists(t, s, "key")
}
