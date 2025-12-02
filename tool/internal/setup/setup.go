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
	"github.com/open-telemetry/opentelemetry-go-compile-instrumentation/tool/internal/rule"
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
	_, err := SetupWithMatched(ctx, args)
	return err
}

// SetupWithMatched prepares the environment for further instrumentation and
// returns the matched rule sets for use in the build phase.
func SetupWithMatched(ctx context.Context, args []string) ([]*rule.InstRuleSet, error) {
	logger := util.LoggerFromContext(ctx)

	if isSetup() {
		logger.InfoContext(ctx, "Setup has already been completed, skipping setup.")
		return nil, nil
	}

	sp := &SetupPhase{
		logger: logger,
	}
	// Find all dependencies of the project being build
	deps, err := sp.findDeps(ctx, args)
	if err != nil {
		return nil, err
	}
	// Match the hook code with these dependencies
	matched, err := sp.matchDeps(ctx, deps)
	if err != nil {
		return nil, err
	}
	// Introduce additional hook code by generating otel.instrumentation.go
	err = sp.addDeps(matched)
	if err != nil {
		return nil, err
	}
	// Extract the embedded instrumentation modules into local directory
	err = sp.extract()
	if err != nil {
		return nil, err
	}
	// Sync new dependencies to go.mod or vendor/modules.txt
	err = sp.syncDeps(ctx, matched)
	if err != nil {
		return nil, err
	}
	// Write the matched hook to matched.txt for further instrument phase
	err = sp.store(matched)
	if err != nil {
		return nil, err
	}
	return matched, nil
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
	const additionalCount = 2
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

// BuildWithToolexecAndModules builds the project with toolexec mode and passes
// matched module paths via environment for fast filtering in subprocesses.
func BuildWithToolexecAndModules(ctx context.Context, args []string, matched []*rule.InstRuleSet) error {
	logger := util.LoggerFromContext(ctx)

	// Add -toolexec=otel to the original build command and run it
	execPath, err := os.Executable()
	if err != nil {
		return ex.Wrapf(err, "failed to get executable path")
	}
	insert := "-toolexec=" + execPath + " toolexec"
	const additionalCount = 2
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
	// Add the rest
	newArgs = append(newArgs, args[1:]...)
	logger.InfoContext(ctx, "Running go build with toolexec", "args", newArgs)

	// Tell the sub-process the working directory
	env := os.Environ()
	pwd := util.GetOtelWorkDir()
	util.Assert(pwd != "", "invalid working directory")
	env = append(env, fmt.Sprintf("%s=%s", util.EnvOtelWorkDir, pwd))

	// Pass matched module paths via environment for fast filtering
	// This avoids JSON file I/O in every subprocess
	modulePaths := make([]string, 0, len(matched))
	for _, m := range matched {
		modulePaths = append(modulePaths, m.ModulePath)
	}
	env = append(env, fmt.Sprintf("%s=%s", util.EnvOtelMatchedModules, strings.Join(modulePaths, ",")))
	logger.InfoContext(ctx, "Matched modules for fast filtering", "modules", modulePaths)

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
		err = os.RemoveAll(OtelRuntimeFile)
		if err != nil {
			logger.DebugContext(ctx, "failed to remove otel runtime file", "error", err)
		}
		err = os.RemoveAll(unzippedPkgDir)
		if err != nil {
			logger.DebugContext(ctx, "failed to remove unzipped pkg", "error", err)
		}
		err = util.RestoreFile(backupFiles)
		if err != nil {
			logger.DebugContext(ctx, "failed to restore files", "error", err)
		}
	}()

	matched, err := SetupWithMatched(ctx, os.Args[1:])
	if err != nil {
		return err
	}
	logger.InfoContext(ctx, "Setup completed successfully")

	// Use the new function that passes matched modules via environment
	err = BuildWithToolexecAndModules(ctx, args, matched)
	if err != nil {
		return err
	}
	logger.InfoContext(ctx, "Instrumentation completed successfully")
	return nil
}
