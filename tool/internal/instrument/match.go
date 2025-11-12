// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package instrument

import (
	"os"

	"github.com/open-telemetry/opentelemetry-go-compile-instrumentation/tool/ex"
	"github.com/open-telemetry/opentelemetry-go-compile-instrumentation/tool/internal/rule"
	"github.com/open-telemetry/opentelemetry-go-compile-instrumentation/tool/util"
)

// load loads the matched rules from the build temp directory.
// TODO: Shared memory across all sub-processes is possible
func (ip *InstrumentPhase) load() ([]*rule.InstRuleSet, error) {
	f := util.GetMatchedRuleFile()
	content, err := os.ReadFile(f)
	if err != nil {
		return nil, ex.Wrapf(err, "failed to read file %s", f)
	}

	// Use the factory function to load and validate rule sets from JSON
	rset, err := rule.LoadInstRuleSetsJSON(content)
	if err != nil {
		return nil, ex.Wrapf(err, "failed to load rule sets from JSON")
	}

	ip.Debug("Load matched rule sets", "path", f)
	return rset, nil
}

// match matches the rules with the compile command.
func (ip *InstrumentPhase) match(allSet []*rule.InstRuleSet, args []string) *rule.InstRuleSet {
	// One package can only be matched with one rule set, so it's safe to return
	// the first matched rule set.
	importPath := util.FindFlagValue(args, "-p")
	util.Assert(importPath != "", "sanity check")
	for _, rset := range allSet {
		if rset.ModulePath == importPath {
			ip.Debug("Match rule set", "set", rset)
			return rset
		}
	}
	return nil
}
