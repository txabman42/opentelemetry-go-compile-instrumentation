// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package setup

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"

	"github.com/open-telemetry/opentelemetry-go-compile-instrumentation/tool/ex"
	"github.com/open-telemetry/opentelemetry-go-compile-instrumentation/tool/internal/rule"
	"github.com/open-telemetry/opentelemetry-go-compile-instrumentation/tool/util"
)

func parseGoMod(gomod string) (*modfile.File, error) {
	data, err := os.ReadFile(gomod)
	if err != nil {
		return nil, ex.Wrapf(err, "failed to read go.mod file")
	}
	modFile, err := modfile.Parse(gomod, data, nil)
	if err != nil {
		return nil, ex.Wrapf(err, "failed to parse go.mod file")
	}
	return modFile, nil
}

func writeGoMod(gomod string, modfile *modfile.File) error {
	data, err := modfile.Format()
	if err != nil {
		return ex.Wrapf(err, "failed to format go.mod file")
	}
	err = os.WriteFile(gomod, data, 0o644) //nolint:gosec // 0644 is ok
	if err != nil {
		return ex.Wrapf(err, "failed to write go.mod file")
	}
	return nil
}

func runModTidy(ctx context.Context) error {
	return util.RunCmd(ctx, "go", "mod", "tidy")
}

// copyInstrumentationToVendor copies the instrumentation packages to vendor directory
// since go mod vendor doesn't copy local replace directives
func (sp *SetupPhase) copyInstrumentationToVendor(matched []*rule.InstRuleSet) error {
	// Collect all instrumentation package paths
	pkgPaths := make(map[string]string) // module path -> local path

	// Add the base pkg module
	oldPath := util.OtelRoot + "/pkg"
	newPath := filepath.Join(util.GetBuildTempDir(), "pkg")
	pkgPaths[oldPath] = newPath

	// Add instrumentation modules
	for _, m := range matched {
		funcRules := m.GetFuncRules()
		for _, rule := range funcRules {
			if !strings.HasPrefix(rule.Path, util.OtelRoot) {
				continue
			}
			oldPath := rule.Path
			newPath := strings.TrimPrefix(oldPath, util.OtelRoot)
			newPath = filepath.Join(util.GetBuildTempDir(), newPath)
			pkgPaths[oldPath] = newPath
		}
	}

	// Copy each package to vendor/
	for modulePath, localPath := range pkgPaths {
		// Convert module path to vendor path
		vendorPath := filepath.Join("vendor", modulePath)

		// Check if source exists
		if !util.PathExists(localPath) {
			sp.Warn("Instrumentation package not found", "path", localPath)
			continue
		}

		// Copy directory recursively
		sp.Info("Copying instrumentation package to vendor", "from", localPath, "to", vendorPath)
		err := copyDir(localPath, vendorPath)
		if err != nil {
			return ex.Wrapf(err, "failed to copy instrumentation package to vendor")
		}
	}

	return nil
}

// copyDir recursively copies a directory
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		// Calculate destination path
		dstPath := filepath.Join(dst, relPath)

		// If it's a directory, create it
		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		// If it's a file, copy it
		return util.CopyFile(path, dstPath)
	})
}

// fixVendorModulesTxt removes replace directives for instrumentation packages in vendor/modules.txt
// so Go uses the vendored files instead of looking in .otel-build
func (sp *SetupPhase) fixVendorModulesTxt() error {
	modulesFile := "vendor/modules.txt"
	content, err := os.ReadFile(modulesFile)
	if err != nil {
		return ex.Wrapf(err, "failed to read vendor/modules.txt")
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string

	for _, line := range lines {
		// Remove replace directives that point to .otel-build
		if strings.Contains(line, "opentelemetry-go-compile-instrumentation") &&
			strings.Contains(line, "=> /") &&
			strings.Contains(line, ".otel-build") {
			// Skip this line (it's a replace directive)
			sp.Debug("Removing replace directive from vendor/modules.txt", "line", line)
			continue
		}
		newLines = append(newLines, line)
	}

	err = os.WriteFile(modulesFile, []byte(strings.Join(newLines, "\n")), 0644)
	if err != nil {
		return ex.Wrapf(err, "failed to write vendor/modules.txt")
	}

	sp.Info("Updated vendor/modules.txt to remove replace directives")
	return nil
}

func addReplace(modfile *modfile.File, path, version, rpath, rversion string) (bool, error) {
	hasReplace := false
	for _, r := range modfile.Replace {
		if r.Old.Path == path {
			hasReplace = true
			break
		}
	}
	if !hasReplace {
		err := modfile.AddReplace(path, version, rpath, rversion)
		if err != nil {
			return false, ex.Wrapf(err, "failed to add replace directive")
		}
		return true, nil
	}
	return false, nil
}

func (sp *SetupPhase) syncDeps(ctx context.Context, matched []*rule.InstRuleSet) error {
	rules := make([]*rule.InstFuncRule, 0)
	for _, m := range matched {
		funcRules := m.GetFuncRules()
		rules = append(rules, funcRules...)
	}
	if len(rules) == 0 {
		return nil
	}

	// In a matching rule, such as InstFuncRule, the hook code is defined in a
	// separate module. Since this module is local, we need to add a replace
	// directive in go.mod to point the module name to its local path.
	const goModFile = "go.mod"
	modfile, err := parseGoMod(goModFile)
	if err != nil {
		return err
	}
	changed := false
	// Add matched dependencies to go.mod
	for _, m := range rules {
		util.Assert(strings.HasPrefix(m.Path, util.OtelRoot), "sanity check")
		// TODO: Since we haven't published the instrumentation packages yet,
		// we need to add the replace directive to the local path.
		// Once the instrumentation packages are published, we can remove this.
		oldPath := m.Path
		newPath := strings.TrimPrefix(oldPath, util.OtelRoot)
		newPath = filepath.Join(util.GetBuildTempDir(), newPath)
		added, addErr := addReplace(modfile, oldPath, "", newPath, "")
		if addErr != nil {
			return addErr
		}
		changed = changed || added
		if changed {
			sp.Info("Replace dependency", "old", oldPath, "new", newPath)
		}
	}
	// TODO: Since we haven't published the pkg packages yet, we need to add the
	// replace directive to the local path. Once the pkg packages are published,
	// we can remove this.
	// Add special pkg module to go.mod
	oldPath := util.OtelRoot + "/pkg"
	newPath := filepath.Join(util.GetBuildTempDir(), "pkg")
	added, addErr := addReplace(modfile, oldPath, "", newPath, "")
	if addErr != nil {
		return addErr
	}
	changed = changed || added
	if changed {
		sp.Info("Replace dependency", "old", oldPath, "new", newPath)
	}
	if changed {
		err = writeGoMod(goModFile, modfile)
		if err != nil {
			return err
		}
		err = runModTidy(ctx)
		if err != nil {
			return err
		}
		// Check if vendor directory exists and sync it
		if util.PathExists("vendor") {
			sp.Info("Vendor directory detected, syncing vendor/modules.txt")
			err = util.RunCmd(ctx, "go", "mod", "vendor")
			if err != nil {
				return ex.Wrapf(err, "failed to sync vendor directory")
			}
			// go mod vendor doesn't copy local replace directives to vendor/
			// so we need to manually copy the instrumentation packages
			err = sp.copyInstrumentationToVendor(matched)
			if err != nil {
				return err
			}
			// Update vendor/modules.txt to remove replace paths
			err = sp.fixVendorModulesTxt()
			if err != nil {
				return err
			}
		}
		sp.keepForDebug(goModFile)
	}
	return nil
}
