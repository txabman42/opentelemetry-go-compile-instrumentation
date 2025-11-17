//go:build integration

// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/open-telemetry/opentelemetry-go-compile-instrumentation/test/app"
)

func TestBasic(t *testing.T) {
	appDir := filepath.Join("..", "..", "demo", "basic")

	app.Build(t, appDir, "go", "build", "-a")
	output := app.Run(t, appDir)
	expect := []string{
		"Every1",
		"Every3",
		"MyStruct.Example",
		"traceID: 123, spanID: 456",
		"[MyHook]",
		"=setupOpenTelemetry=",
		"RawCode",
	}
	for _, e := range expect {
		require.Contains(t, output, e)
	}
	require.Contains(t, output, "GenericExample before hook")
	require.Contains(t, output, "Hello, Generic World!")
	require.Contains(t, output, "GenericExample after hook")
	require.Contains(t, output, "GenericLookupTableExample before hook")
	require.Contains(t, output, "GenericLookupTableExample after hook")
}
