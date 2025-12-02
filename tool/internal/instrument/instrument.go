// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package instrument

import (
	"github.com/open-telemetry/opentelemetry-go-compile-instrumentation/tool/internal/rule"
	"github.com/open-telemetry/opentelemetry-go-compile-instrumentation/tool/util"
)

func groupRules(rset *rule.InstRuleSet) map[string][]rule.InstRule {
	file2rules := make(map[string][]rule.InstRule)
	for file, rules := range rset.FuncRules {
		for _, rule := range rules {
			file2rules[file] = append(file2rules[file], rule)
		}
	}
	for file, rules := range rset.StructRules {
		for _, rule := range rules {
			file2rules[file] = append(file2rules[file], rule)
		}
	}
	for file, rules := range rset.RawRules {
		for _, rule := range rules {
			file2rules[file] = append(file2rules[file], rule)
		}
	}
	return file2rules
}

// filterRulesWithQuickCheck filters out rules for functions that definitely
// don't exist in their target files. This is a performance optimization that
// avoids expensive AST parsing for files that don't contain the target functions.
func (ip *InstrumentPhase) filterRulesWithQuickCheck(file string, rules []rule.InstRule) []rule.InstRule {
	filtered := make([]rule.InstRule, 0, len(rules))
	for _, r := range rules {
		switch rt := r.(type) {
		case *rule.InstFuncRule:
			// Quick check if the function might exist in this file
			if quickFuncExistsCheck(file, rt.Func) {
				filtered = append(filtered, r)
			} else {
				ip.Debug("Quick check: function not found, skipping", "file", file, "func", rt.Func)
			}
		default:
			// For non-func rules (struct, raw), always include them
			filtered = append(filtered, r)
		}
	}
	return filtered
}

func (ip *InstrumentPhase) instrument(rset *rule.InstRuleSet) error {
	hasFuncRule := false
	// Apply file rules first because they can introduce new files that used
	// by other rules such as raw rules
	for _, rule := range rset.FileRules {
		err := ip.applyFileRule(rule, rset.PackageName)
		if err != nil {
			return err
		}
	}
	for file, rules := range groupRules(rset) {
		// Strategy D: Quick check to filter out rules for non-existent functions
		// This avoids expensive AST parsing for files that don't contain targets
		filteredRules := ip.filterRulesWithQuickCheck(file, rules)
		if len(filteredRules) == 0 {
			ip.Debug("No matching rules after quick check, skipping file", "file", file)
			continue
		}

		// Group rules by file, then parse the target file once
		root, err := ip.parseFile(file)
		if err != nil {
			return err
		}

		// Apply the rules to the target file
		for _, r := range filteredRules {
			switch rt := r.(type) {
			case *rule.InstFuncRule:
				err1 := ip.applyFuncRule(rt, root)
				if err1 != nil {
					return err1
				}
				hasFuncRule = true
			case *rule.InstStructRule:
				err1 := ip.applyStructRule(rt, root)
				if err1 != nil {
					return err1
				}
			case *rule.InstRawRule:
				err1 := ip.applyRawRule(rt, root)
				if err1 != nil {
					return err1
				}
				hasFuncRule = true
			default:
				util.ShouldNotReachHere()
			}
		}
		// Since trampoline-jump-if is performance-critical, perform AST level
		// optimization for them before writing to file
		err = ip.optimizeTJumps()
		if err != nil {
			return err
		}
		// Once all func rules targeting this file are applied, write instrumented
		// AST to new file and replace the original file in the compile command
		err = ip.writeInstrumented(root, file)
		if err != nil {
			return err
		}
	}

	// Write globals file if any function is instrumented because injected code
	// always requires some global variables and auxiliary declarations
	if hasFuncRule {
		return ip.writeGlobals(rset.PackageName)
	}
	return nil
}
