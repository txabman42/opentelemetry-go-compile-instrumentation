// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package testutil

import (
	"testing"

	"go.opentelemetry.io/collector/pdata/ptrace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

func TestRequireMessagingSemconv(t *testing.T) {
	t.Run("kafka publish", func(t *testing.T) {
		s := ptrace.NewSpan()
		s.SetName("my-topic publish")
		s.SetKind(ptrace.SpanKindProducer)
		s.Attributes().PutStr(string(semconv.MessagingSystemKey), "kafka")
		s.Attributes().PutStr(string(semconv.MessagingDestinationNameKey), "my-topic")
		s.Attributes().PutStr(string(semconv.MessagingOperationTypeKey), "publish")

		RequireMessagingSemconv(t, s, "kafka", "my-topic", "publish")
	})

	t.Run("rabbitmq receive", func(t *testing.T) {
		s := ptrace.NewSpan()
		s.SetName("my-queue receive")
		s.SetKind(ptrace.SpanKindConsumer)
		s.Attributes().PutStr(string(semconv.MessagingSystemKey), "rabbitmq")
		s.Attributes().PutStr(string(semconv.MessagingDestinationNameKey), "my-queue")
		s.Attributes().PutStr(string(semconv.MessagingOperationTypeKey), "receive")

		RequireMessagingSemconv(t, s, "rabbitmq", "my-queue", "receive")
	})

	t.Run("rocketmq process", func(t *testing.T) {
		s := ptrace.NewSpan()
		s.SetName("orders process")
		s.SetKind(ptrace.SpanKindConsumer)
		s.Attributes().PutStr(string(semconv.MessagingSystemKey), "rocketmq")
		s.Attributes().PutStr(string(semconv.MessagingDestinationNameKey), "orders")
		s.Attributes().PutStr(string(semconv.MessagingOperationTypeKey), "process")

		RequireMessagingSemconv(t, s, "rocketmq", "orders", "process")
	})
}

func TestRequireHTTPClientSemconv(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		s := ptrace.NewSpan()
		s.SetName("GET")
		s.SetKind(ptrace.SpanKindClient)
		s.Attributes().PutStr(string(semconv.HTTPRequestMethodKey), "GET")
		s.Attributes().PutStr(string(semconv.ServerAddressKey), "example.com")
		s.Attributes().PutStr(string(semconv.URLFullKey), "https://example.com/api/v1/users")
		s.Attributes().PutInt(string(semconv.HTTPResponseStatusCodeKey), 200)
		s.Attributes().PutStr(string(semconv.NetworkProtocolVersionKey), "1.1")
		s.Attributes().PutStr(string(semconv.URLSchemeKey), "https")
		s.Attributes().PutInt(string(semconv.ServerPortKey), 443)

		RequireHTTPClientSemconv(t, s, "GET", "https://example.com/api/v1/users", "example.com", 200, 443, "1.1", "https")
	})
}
