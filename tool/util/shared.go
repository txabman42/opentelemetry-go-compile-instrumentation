// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

const (
	EnvOtelWorkDir        = "OTEL_WORK_DIR"
	EnvOtelMatchedModules = "OTEL_MATCHED_MODULES"
	BuildTempDir          = ".otel-build"
	OtelRoot              = "github.com/open-telemetry/opentelemetry-go-compile-instrumentation"
)

func GetMatchedRuleFile() string {
	const matchedRuleFile = "matched.json"
	return GetBuildTemp(matchedRuleFile)
}

func GetOtelWorkDir() string {
	wd := os.Getenv(EnvOtelWorkDir)
	if wd == "" {
		wd, _ = os.Getwd()
		return wd
	}
	return wd
}

// GetBuildTemp returns the path to the build temp directory $BUILD_TEMP/name
func GetBuildTempDir() string {
	return filepath.Join(GetOtelWorkDir(), BuildTempDir)
}

// GetBuildTemp returns the path to the build temp directory $BUILD_TEMP/name
func GetBuildTemp(name string) string {
	return filepath.Join(GetOtelWorkDir(), BuildTempDir, name)
}

func copyBackupFiles(names []string, src, dst string) error {
	var err error
	for _, name := range names {
		srcFile := filepath.Join(src, name)
		dstFile := filepath.Join(dst, name)
		err = errors.Join(err, CopyFile(srcFile, dstFile))
	}
	return err
}

// BackupFile backups the source file to $BUILD_TEMP/backup/name.
func BackupFile(names []string) error {
	return copyBackupFiles(names, ".", GetBuildTemp("backup"))
}

// RestoreFile restores the source file from $BUILD_TEMP/backup/name.
func RestoreFile(names []string) error {
	return copyBackupFiles(names, GetBuildTemp("backup"), ".")
}

// GetMatchedModules returns the list of matched module paths from environment.
// Returns nil if the environment variable is not set.
func GetMatchedModules() []string {
	env := os.Getenv(EnvOtelMatchedModules)
	if env == "" {
		return nil
	}
	return strings.Split(env, ",")
}

// IsModuleMatched checks if the given module path is in the matched modules list.
// This is a fast check that avoids loading the full rules JSON.
func IsModuleMatched(modulePath string) bool {
	modules := GetMatchedModules()
	if modules == nil {
		// Fallback: environment not set, need to check rules file
		return true
	}
	for _, m := range modules {
		if m == modulePath {
			return true
		}
	}
	return false
}
