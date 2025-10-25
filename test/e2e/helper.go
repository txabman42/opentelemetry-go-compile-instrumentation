// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"gotest.tools/v3/golden"
)

var (
	rootDir string
	testDir string
)

func init() {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("Failed to get caller information")
	}
	testDir = filepath.Dir(filename)
	rootDir = filepath.Dir(filepath.Dir(testDir))
}

// getOtelBinary returns the path to the otel binary.
func getOtelBinary(t *testing.T) string {
	t.Helper()
	otelBinary := filepath.Join(rootDir, "otel")
	if _, err := os.Stat(otelBinary); os.IsNotExist(err) {
		t.Skip("otel binary not found. Run 'make build' first.")
	}
	return otelBinary
}

// buildApp builds the application with instrumentation.
func buildApp(t *testing.T, otelBinary, appDir string) {
	t.Helper()
	cmd := exec.Command(otelBinary, "go", "build", "-a", "-o", "testapp")
	cmd.Dir = appDir
	cmd.Env = append(os.Environ(), "GO111MODULE=on")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Build failed: %v\n%s", err, out)
	}
}

// runApp runs the built application and captures its output.
func runApp(t *testing.T, appDir string) (stdout, stderr string) {
	t.Helper()
	cmd := exec.Command("./testapp")
	cmd.Dir = appDir
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	if err := cmd.Run(); err != nil {
		t.Logf("stdout:\n%s", stdoutBuf.String())
		t.Logf("stderr:\n%s", stderrBuf.String())
		t.Fatalf("App failed: %v", err)
	}

	return stdoutBuf.String(), stderrBuf.String()
}

// Build builds the application with instrumentation and runs it.
func Build(t *testing.T, appName string) (stdout, stderr string) {
	t.Helper()
	appDir := filepath.Join(testDir, appName)
	otelBinary := getOtelBinary(t)
	buildApp(t, otelBinary, appDir)
	return runApp(t, appDir)
}

// FilterJSON removes JSON lines (lines starting with '{') from the output.
func FilterJSON(text string) string {
	var filtered []string
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "{") {
			filtered = append(filtered, line)
		}
	}
	return strings.Join(filtered, "\n")
}

// Golden compares the actual output against the golden file.
func Golden(t *testing.T, actual, goldenFile string) {
	t.Helper()
	golden.Assert(t, actual, goldenFile)
}
