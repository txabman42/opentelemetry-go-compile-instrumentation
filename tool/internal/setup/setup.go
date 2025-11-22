// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package setup

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/open-telemetry/opentelemetry-go-compile-instrumentation/tool/ex"
	"github.com/open-telemetry/opentelemetry-go-compile-instrumentation/tool/util"
)

type SetupPhase struct {
	logger *slog.Logger
}

func (sp *SetupPhase) Info(msg string, args ...any)  { sp.logger.Info(msg, args...) }
func (sp *SetupPhase) Error(msg string, args ...any) { sp.logger.Error(msg, args...) }
func (sp *SetupPhase) Warn(msg string, args ...any)  { sp.logger.Warn(msg, args...) }
func (sp *SetupPhase) Debug(msg string, args ...any) { sp.logger.Debug(msg, args...) }

// keepForDebug copies the file to the build temp directory for debugging
// Error is tolerated as it's not critical.
func (sp *SetupPhase) keepForDebug(name string) {
	dstFile := filepath.Join(util.GetBuildTemp("debug"), "main", name)
	err := util.CopyFile(name, dstFile)
	if err != nil {
		sp.Warn("failed to record added file", "file", name, "error", err)
	}
}

// This function can be used to check if the setup has been completed.
func isSetup() bool {
	// TODO: Implement Task
	return false
}

// extractPackagePath extracts the package path from build arguments
// E.g., for "go build -a ./app/vmctl", it returns "./app/vmctl"
// For "go build -a", it returns "." (current directory)
func extractPackagePath(args []string) string {
	// Args typically look like: ["build", "-a", "./app/vmctl"] or ["build", "-a"]
	// Find the package path after the command and flags
	foundCommand := false
	for i := 0; i < len(args); i++ {
		arg := args[i]

		// Identify the go command (build, install, run, test)
		if arg == "build" || arg == "install" || arg == "run" || arg == "test" {
			foundCommand = true
			continue
		}

		// Skip flags (arguments starting with -)
		if strings.HasPrefix(arg, "-") {
			// Some flags take values, but they also start with - so we can skip them
			continue
		}

		// After we've found the command, the next non-flag argument is the package path
		if foundCommand && arg != "" {
			// This is the package path
			return arg
		}
	}

	// No explicit package path found, use current directory
	return "."
}

// Setup prepares the environment for further instrumentation.
func Setup(ctx context.Context, args []string) error {
	logger := util.LoggerFromContext(ctx)

	if isSetup() {
		logger.InfoContext(ctx, "Setup has already been completed, skipping setup.")
		return nil
	}

	sp := &SetupPhase{
		logger: logger,
	}

	// If vendor directory exists, sync it with go.mod before starting
	// This handles the case where vendor/ is out of sync from a previous run
	if util.PathExists("vendor") {
		sp.Info("Syncing vendor directory with go.mod before setup")
		err := util.RunCmd(ctx, "go", "mod", "vendor")
		if err != nil {
			sp.Warn("Failed to sync vendor directory before setup", "error", err)
			// Continue anyway - the sync during setup might fix it
		}
	}
	// Find all dependencies of the project being build
	deps, err := sp.findDeps(ctx, args)
	if err != nil {
		return err
	}
	// Match the hook code with these dependencies
	matched, err := sp.matchDeps(ctx, deps)
	if err != nil {
		return err
	}
	// Determine the package path from build args (e.g., "./app/vmctl" from "go build -a ./app/vmctl")
	packagePath := extractPackagePath(args)
	// Introduce additional hook code by generating otel.instrumentation.go
	err = sp.addDeps(matched, packagePath)
	if err != nil {
		return err
	}
	// Extract the embedded instrumentation modules into local directory
	err = sp.extract()
	if err != nil {
		return err
	}
	// Sync new dependencies to go.mod or vendor/modules.txt
	err = sp.syncDeps(ctx, matched)
	if err != nil {
		return err
	}
	// Write the matched hook to matched.txt for further instrument phase
	err = sp.store(matched)
	if err != nil {
		return err
	}
	sp.Info("Setup completed successfully")
	return nil
}

// BuildWithToolexec builds the project with the toolexec mode
func BuildWithToolexec(ctx context.Context, args []string) error {
	logger := util.LoggerFromContext(ctx)

	// Add -toolexec=otel to the original build command and run it
	execPath, err := os.Executable()
	if err != nil {
		return ex.Wrapf(err, "failed to get executable path")
	}
	insert := "-toolexec=" + execPath + " toolexec"
	const additionalCount = 3                               // -work, -toolexec, -a, -mod
	newArgs := make([]string, 0, len(args)+additionalCount) // Avoid in-place modification
	// Add "go build"
	newArgs = append(newArgs, "go")
	newArgs = append(newArgs, args[:1]...)
	// Add "-work" to give us a chance to debug instrumented code if needed
	newArgs = append(newArgs, "-work")
	// Add "-toolexec=..."
	newArgs = append(newArgs, insert)
	// TODO: We should support incremental build in the future, so we don't need
	// to force rebuild here.
	// Add "-a" to force rebuild
	newArgs = append(newArgs, "-a")
	// If vendor directory exists, use -mod=mod to bypass vendor and use go.mod
	if util.PathExists("vendor") {
		newArgs = append(newArgs, "-mod=mod")
		logger.InfoContext(ctx, "Using -mod=mod to bypass vendor directory")
	}
	// Add the rest
	newArgs = append(newArgs, args[1:]...)
	logger.InfoContext(ctx, "Running go build with toolexec", "args", newArgs)

	// Tell the sub-process the working directory
	env := os.Environ()
	pwd := util.GetOtelWorkDir()
	util.Assert(pwd != "", "invalid working directory")
	env = append(env, fmt.Sprintf("%s=%s", util.EnvOtelWorkDir, pwd))

	return util.RunCmdWithEnv(ctx, env, newArgs...)
}

func GoBuild(ctx context.Context, args []string) error {
	logger := util.LoggerFromContext(ctx)
	backupFiles := []string{"go.mod", "go.sum", "go.work", "go.work.sum"}
	err := util.BackupFile(backupFiles)
	if err != nil {
		logger.DebugContext(ctx, "failed to back up files", "error", err)
	}

	// Determine package path to know where otel.runtime.go will be created
	packagePath := extractPackagePath(args)
	runtimeFilePath := OtelRuntimeFile
	if packagePath != "" && packagePath != "." {
		runtimeFilePath = filepath.Join(packagePath, OtelRuntimeFile)
	}

	defer func() {
		err = os.RemoveAll(runtimeFilePath)
		if err != nil {
			logger.DebugContext(ctx, "failed to remove otel runtime file", "error", err, "path", runtimeFilePath)
		}
		err = util.RestoreFile(backupFiles)
		if err != nil {
			logger.DebugContext(ctx, "failed to restore files", "error", err)
		}
		// Note: We don't re-run go mod vendor here because:
		// 1. The instrumentation packages were manually copied to vendor/
		// 2. Running go mod vendor would remove them since they're not in the restored go.mod
		// 3. The next build will sync vendor/ at the beginning of Setup if needed
	}()

	err = Setup(ctx, os.Args[1:])
	if err != nil {
		return err
	}
	err = BuildWithToolexec(ctx, args)
	if err != nil {
		return err
	}
	return nil
}
