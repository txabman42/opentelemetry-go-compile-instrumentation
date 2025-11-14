// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:build !windows

package instrument

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestInstrument verifies that the complete instrumentation workflow works end-to-end.
// This test uses the embedded helloworld instrumentation rules to verify that:
//  1. The otel tool successfully wraps the go build command
//  2. The setup phase finds and matches instrumentation rules
//  3. The toolexec phase applies instrumentation during compilation
//  4. The resulting binary executes the injected hooks
func TestInstrument(t *testing.T) {
	tempDir := t.TempDir()
	setupTestFiles(t, tempDir)

	// Build with instrumentation using the full otel workflow.
	buildWithInstrumentation(t, tempDir)

	// Verify the binary was created.
	binaryPath := filepath.Join(tempDir, "test")
	require.FileExists(t, binaryPath)

	// Run and verify instrumentation was applied.
	output := runBinary(t, binaryPath)
	require.Contains(t, output, "MyHook", "embedded hook should be executed")
}

func buildWithInstrumentation(t *testing.T, appDir string) {
	t.Helper()

	pwd, err := os.Getwd()
	require.NoError(t, err)

	// Navigate to workspace root (3 levels up from tool/internal/instrument).
	otelPath := filepath.Join(pwd, "..", "..", "..", "otel")

	cmd := exec.Command(otelPath, "go", "build", "-a")
	cmd.Dir = appDir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "build failed: %s", string(out))
}

func runBinary(t *testing.T, binaryPath string) string {
	t.Helper()

	cmd := exec.Command(binaryPath)
	cmd.Dir = filepath.Dir(binaryPath)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "binary execution failed: %s", string(output))
	return string(output)
}

func setupTestFiles(t *testing.T, tempDir string) {
	t.Helper()

	// Calculate workspace root path for consistent replace directive.
	pwd, err := os.Getwd()
	require.NoError(t, err)
	workspaceRoot := filepath.Join(pwd, "..", "..", "..")
	pkgPath := filepath.Join(workspaceRoot, "pkg")

	// Create a simple main.go that uses the Example function (for helloworld instrumentation).
	mainContent := `// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package main

type MyStruct struct{}

func (m *MyStruct) Example() { println("MyStruct.Example") }

// Example demonstrates how to use the instrumenter.
func Example() {
	println("Original Example function")
}

func main() {
	Example()
	m := &MyStruct{}
	m.Example()
}
`
	err = os.WriteFile(filepath.Join(tempDir, "main.go"), []byte(mainContent), 0o644)
	require.NoError(t, err)

	// Create go.mod.
	goModContent := `module test

go 1.23

require github.com/open-telemetry/opentelemetry-go-compile-instrumentation/pkg v0.0.0

replace github.com/open-telemetry/opentelemetry-go-compile-instrumentation/pkg => ` + pkgPath + `
`
	err = os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goModContent), 0o644)
	require.NoError(t, err)
}
