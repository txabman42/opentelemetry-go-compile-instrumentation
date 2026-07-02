// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"
	"log/slog"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func TestLogLevel(t *testing.T) {
	tests := []struct {
		envValue string
		expected slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"error", slog.LevelError},
		{"", slog.LevelInfo},        // unset defaults to info
		{"unknown", slog.LevelInfo}, // unrecognized value defaults to info
	}

	for _, tt := range tests {
		name := tt.envValue
		if name == "" {
			name = "unset"
		}
		t.Run(name, func(t *testing.T) {
			t.Setenv("OTEL_LOG_LEVEL", tt.envValue)
			assert.Equal(t, tt.expected, logLevel())
		})
	}
}

func TestShutdownNilProviders(t *testing.T) {
	// The providers are package globals shared across tests, so save and restore
	// them to keep this test order-independent.
	origTracer, origMeter := tracerProvider, meterProvider
	t.Cleanup(func() { tracerProvider, meterProvider = origTracer, origMeter })

	// With no providers configured, Shutdown is a no-op that returns nil.
	tracerProvider = nil
	meterProvider = nil
	require.NoError(t, Shutdown(context.Background()))
}

func TestShutdownProviders(t *testing.T) {
	origTracer, origMeter := tracerProvider, meterProvider
	t.Cleanup(func() { tracerProvider, meterProvider = origTracer, origMeter })

	// Providers with no exporters shut down cleanly, exercising both shutdown
	// branches without any network dependency.
	tracerProvider = sdktrace.NewTracerProvider()
	meterProvider = sdkmetric.NewMeterProvider()
	require.NoError(t, Shutdown(context.Background()))
}

// restoreProviders resets the global providers after a test that configures them.
func restoreProviders(t *testing.T) {
	origTracer, origMeter := tracerProvider, meterProvider
	t.Cleanup(func() { tracerProvider, meterProvider = origTracer, origMeter })
}

func TestSetupOpenTelemetry(t *testing.T) {
	// A configured OTLP endpoint drives the full trace-provider setup path.
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4317")
	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "grpc")
	restoreProviders(t)

	err := setupOpenTelemetry(Config{
		ServiceName:            "test-service",
		ServiceVersion:         "1.0.0",
		InstrumentationName:    "test-inst",
		InstrumentationVersion: "2.0.0",
	})
	require.NoError(t, err)
}

func TestSetupOpenTelemetryExporterError(t *testing.T) {
	// An invalid protocol makes the exporter fail to build, but setupOpenTelemetry
	// logs and swallows the error rather than propagating it.
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4317")
	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "invalid-protocol")
	restoreProviders(t)

	err := setupOpenTelemetry(Config{ServiceName: "test-service"})
	require.NoError(t, err)
}

func TestInitializePanicRecovery(t *testing.T) {
	// Initialize must never crash the host application. A nil logger forces a
	// panic inside setup, which the deferred recover in Initialize must absorb.
	origLogger := logger
	t.Cleanup(func() {
		logger = origLogger
		initOnce = sync.Once{}
	})
	restoreProviders(t)

	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4317")
	logger = nil
	initOnce = sync.Once{}

	assert.NotPanics(t, func() {
		Initialize(Config{ServiceName: "test-service"})
	})
}
