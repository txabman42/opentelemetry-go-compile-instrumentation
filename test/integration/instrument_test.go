//go:build integration

// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/open-telemetry/opentelemetry-go-compile-instrumentation/test/app"
	"github.com/stretchr/testify/require"
	"gotest.tools/v3/golden"
)

func TestInstrumentWithDifferentRuleTypes_Integration(t *testing.T) {
	appDir := filepath.Join("..", "..", "demo", "basic")

	// Copy integration.yaml as rules.yaml to the app directory.
	pwd, err := os.Getwd()
	require.NoError(t, err)
	integrationYamlPath := filepath.Join(pwd, "..", "..", "tool", "data", "integration.yaml")
	rulesPath := filepath.Join(appDir, "rules.yaml")
	copyFile(t, integrationYamlPath, rulesPath)
	defer os.Remove(rulesPath) // Clean up after test

	// Build the application with instrumentation.
	app.Build(t, appDir, "go", "build", "-a")

	// Verify generated files match golden files.
	verifyGoldenFiles(t, appDir)
}

func verifyGoldenFiles(t *testing.T, appDir string) {
	t.Helper()

	// List of files to verify against golden files.
	filesToVerify := []string{
		"main.go",
		"otel.globals.go",
	}

	for _, fileName := range filesToVerify {
		actualPath := filepath.Join(appDir, fileName)
		goldenPath := filepath.Join("testdata", "combined_rules", fileName+".golden")

		// Skip if golden file doesn't exist.
		if _, err := os.Stat(goldenPath); os.IsNotExist(err) {
			continue
		}

		// Check that the actual file exists.
		require.FileExists(t, actualPath, "file should exist: %s", actualPath)

		// Read the actual content.
		actualContent, err := os.ReadFile(actualPath)
		require.NoError(t, err, "failed to read file: %s", actualPath)

		// Normalize line endings and remove trailing whitespace.
		normalizedContent := normalizeContent(string(actualContent))

		// Compare with golden file.
		golden.Assert(t, normalizedContent, goldenPath)
	}
}

func normalizeContent(content string) string {
	// Split into lines and normalize each line.
	lines := strings.Split(content, "\n")
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], " \t\r")
	}

	// Join back and ensure single trailing newline.
	normalized := strings.Join(lines, "\n")
	normalized = strings.TrimRight(normalized, "\n") + "\n"
	return normalized
}

func copyFile(t *testing.T, src, dst string) {
	t.Helper()
	data, err := os.ReadFile(src)
	require.NoError(t, err, "failed to read file: %s", src)
	err = os.WriteFile(dst, data, 0o644)
	require.NoError(t, err, "failed to write file: %s", dst)
}
