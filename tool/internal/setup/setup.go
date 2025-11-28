// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package setup

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"golang.org/x/tools/go/packages"

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

	// Get build packages to determine module directories
	pkgs, err := util.GetBuildPackages(args)
	if err != nil {
		return err
	}

	// Global operations (executed once)
	// Find all dependencies of the project being built
	deps, err := sp.findDeps(ctx, args)
	if err != nil {
		return err
	}
	// Match the hook code with these dependencies
	matched, err := sp.matchDeps(ctx, deps)
	if err != nil {
		return err
	}
	// Extract the embedded instrumentation modules into local directory
	err = sp.extract()
	if err != nil {
		return err
	}

	// Per-package operations
	// Track processed directories to avoid generating otel.runtime.go multiple times
	processedPkgDirs := make(map[string]bool)
	processedModDirs := make(map[string]bool)
	for _, pkg := range pkgs {
		if pkg.Module == nil {
			continue
		}
		moduleDir := pkg.Module.Dir
		pkgDir := util.GetPackageDir(pkg)
		if pkgDir == "" {
			// Fallback to module directory if no Go files found
			pkgDir = moduleDir
		}

		// Generate otel.runtime.go in the package directory (not module directory)
		// This ensures the file is compiled with the correct main package
		if !processedPkgDirs[pkgDir] {
			processedPkgDirs[pkgDir] = true
			err = sp.addDeps(matched, pkgDir)
			if err != nil {
				return err
			}
		}
		// Sync new dependencies to go.mod (only once per module)
		if !processedModDirs[moduleDir] {
			processedModDirs[moduleDir] = true
			err = sp.syncDeps(ctx, matched, moduleDir)
			if err != nil {
				return err
			}
		}
	}

	// Write the matched hook to matched.txt for further instrument phase
	return sp.store(matched)
}

// BuildWithToolexec builds the project with the toolexec mode
func BuildWithToolexec(ctx context.Context, args []string) error {
	logger := util.LoggerFromContext(ctx)

	execPath, err := os.Executable()
	if err != nil {
		return ex.Wrapf(err, "failed to get executable path")
	}

	// Extract flags and package patterns from original args
	flags, pkgPatterns := util.SplitArgsAndPackages(args)

	// If multiple packages, build each one separately so binaries are created
	// go build with multiple packages only compiles but doesn't produce binaries
	if len(pkgPatterns) > 1 {
		for _, pattern := range pkgPatterns {
			err = buildSinglePackage(ctx, logger, execPath, flags, pattern)
			if err != nil {
				return err
			}
		}
		return nil
	}

	// Single package: build normally
	return buildSinglePackage(ctx, logger, execPath, args, "")
}

// buildSinglePackage builds a single package with toolexec.
// When pkgPattern is empty, args contains the full build args including package patterns.
// When pkgPattern is non-empty, args contains only the flags (e.g., ["build", "-a"])
// and pkgPattern is the specific package to build.
func buildSinglePackage(ctx context.Context, logger *slog.Logger, execPath string, args []string, pkgPattern string) error {
	insert := "-toolexec=" + execPath + " toolexec"
	const additionalCount = 4
	newArgs := make([]string, 0, len(args)+additionalCount)
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

	if pkgPattern != "" {
		// Building a specific package from multi-package build
		// args already contains only flags like ["build", "-a"], skip "build"
		newArgs = append(newArgs, args[1:]...)
		newArgs = append(newArgs, pkgPattern)
	} else {
		// Single package build: add the rest of the args (flags + package)
		newArgs = append(newArgs, args[1:]...)
	}

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
	defer func() {
		var pkgs []*packages.Package
		pkgs, err = util.GetBuildPackages(os.Args[1:])
		if err != nil {
			logger.DebugContext(ctx, "failed to get build packages", "error", err)
		}
		// Track removed directories to avoid duplicate cleanup
		removedDirs := make(map[string]bool)
		for _, pkg := range pkgs {
			// Remove otel.runtime.go from the package directory (where it was created)
			pkgDir := util.GetPackageDir(pkg)
			if pkgDir == "" && pkg.Module != nil {
				pkgDir = pkg.Module.Dir
			}
			if pkgDir != "" && !removedDirs[pkgDir] {
				removedDirs[pkgDir] = true
				if err = os.RemoveAll(filepath.Join(pkgDir, OtelRuntimeFile)); err != nil {
					logger.DebugContext(ctx, "failed to remove otel.runtime.go", "path", pkgDir, "error", err)
				}
			}
		}
		if err = os.RemoveAll(unzippedPkgDir); err != nil {
			logger.DebugContext(ctx, "failed to remove unzipped pkg", "error", err)
		}
		if err = util.RestoreFile(backupFiles); err != nil {
			logger.DebugContext(ctx, "failed to restore files", "error", err)
		}
	}()

	err = Setup(ctx, os.Args[1:])
	if err != nil {
		return err
	}
	logger.InfoContext(ctx, "Setup completed successfully")

	err = BuildWithToolexec(ctx, args)
	if err != nil {
		return err
	}
	logger.InfoContext(ctx, "Instrumentation completed successfully")
	return nil
}
