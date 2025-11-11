// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package rule

import (
	"gopkg.in/yaml.v3"

	"github.com/open-telemetry/opentelemetry-go-compile-instrumentation/tool/data"
	"github.com/open-telemetry/opentelemetry-go-compile-instrumentation/tool/ex"
	"github.com/open-telemetry/opentelemetry-go-compile-instrumentation/tool/util"
)

// LoadAllRules loads all available rules from the embedded files.
// It discovers all YAML files in the embedded files and parses them into rule instances.
func LoadAllRules() ([]InstRule, error) {
	availables, err := data.ListEmbedFiles()
	if err != nil {
		return nil, err
	}

	parsedRules := []InstRule{}
	for _, available := range availables {
		rs, perr := parseRuleFile(available)
		if perr != nil {
			return nil, perr
		}
		parsedRules = append(parsedRules, rs...)
	}
	return parsedRules, nil
}

// ParseRuleFile parses a YAML file at the given path and returns all rules defined in it.
func parseRuleFile(path string) ([]InstRule, error) {
	yamlFile, err := data.ReadEmbedFile(path)
	if err != nil {
		return nil, err
	}
	return parseRuleFromBytes(yamlFile)
}

// ParseRuleFromBytes parses YAML bytes and returns all rules defined in them.
func parseRuleFromBytes(yamlFile []byte) ([]InstRule, error) {
	var h map[string]map[string]any
	err := yaml.Unmarshal(yamlFile, &h)
	if err != nil {
		return nil, ex.Wrap(err)
	}
	rules := make([]InstRule, 0)
	for name, fields := range h {
		raw, err1 := yaml.Marshal(fields)
		if err1 != nil {
			return nil, ex.Wrap(err1)
		}

		r, err2 := CreateRuleFromFields(raw, name, fields)
		if err2 != nil {
			return nil, err2
		}
		rules = append(rules, r)
	}
	return rules, nil
}

// CreateRuleFromFields creates a rule instance based on the field type present in the YAML.
// It inspects the fields map to determine which rule type to instantiate and unmarshals
// the raw YAML data into the appropriate rule struct.
//
//nolint:nilnil // factory function
func CreateRuleFromFields(raw []byte, name string, fields map[string]any) (InstRule, error) {
	base := InstBaseRule{
		Name: name,
	}
	if target, ok := fields["target"].(string); ok {
		base.Target = target
	}
	if fields["version"] != nil {
		v, ok := fields["version"].(string)
		util.Assert(ok, "version is not a string")
		base.Version = v
	}

	switch {
	case fields["struct"] != nil:
		var r InstStructRule
		if err := yaml.Unmarshal(raw, &r); err != nil {
			return nil, ex.Wrap(err)
		}
		r.InstBaseRule = base
		return &r, nil
	case fields["file"] != nil:
		var r InstFileRule
		if err := yaml.Unmarshal(raw, &r); err != nil {
			return nil, ex.Wrap(err)
		}
		r.InstBaseRule = base
		return &r, nil
	case fields["raw"] != nil:
		var r InstRawRule
		if err := yaml.Unmarshal(raw, &r); err != nil {
			return nil, ex.Wrap(err)
		}
		r.InstBaseRule = base
		return &r, nil
	case fields["func"] != nil:
		var r InstFuncRule
		if err := yaml.Unmarshal(raw, &r); err != nil {
			return nil, ex.Wrap(err)
		}
		r.InstBaseRule = base
		return &r, nil
	default:
		util.ShouldNotReachHere()
		return nil, nil
	}
}
