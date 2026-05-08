package main

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadRules reads and parses a structural-rules YAML file from path.
// It returns a validated RuleSet or an error if the file is missing,
// malformed, or fails validation.
func LoadRules(path string) (*RuleSet, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading rules file %q: %w", path, err)
	}

	var rs RuleSet
	if err := yaml.Unmarshal(data, &rs); err != nil {
		return nil, fmt.Errorf("parsing rules YAML %q: %w", path, err)
	}

	if err := rs.Validate(); err != nil {
		return nil, fmt.Errorf("invalid rules in %q: %w", path, err)
	}

	return &rs, nil
}

// Validate checks the RuleSet for internal consistency.
// It returns the first error found, or nil if the rules are valid.
func (rs *RuleSet) Validate() error {
	if rs.Version == "" {
		return fmt.Errorf("rules version is required")
	}

	if len(rs.Layers) == 0 {
		return fmt.Errorf("at least one layer rule is required")
	}

	for i, lr := range rs.Layers {
		if lr.Name == "" {
			return fmt.Errorf("layer rule [%d]: name is required", i)
		}
		if lr.PathPattern == "" {
			return fmt.Errorf("layer rule %q: pathPattern is required", lr.Name)
		}
	}

	for i, fr := range rs.FileSizes {
		if fr.Name == "" {
			return fmt.Errorf("fileSize rule [%d]: name is required", i)
		}
		if fr.PathPattern == "" {
			return fmt.Errorf("fileSize rule %q: pathPattern is required", fr.Name)
		}
		if fr.MaxEffectiveLines < 1 {
			return fmt.Errorf("fileSize rule %q: maxEffectiveLines must be >= 1", fr.Name)
		}
	}

	for i, fn := range rs.FunctionSizes {
		if fn.Name == "" {
			return fmt.Errorf("functionSize rule [%d]: name is required", i)
		}
		if fn.PathPattern == "" {
			return fmt.Errorf("functionSize rule %q: pathPattern is required", fn.Name)
		}
		if fn.MaxEffectiveLines < 1 {
			return fmt.Errorf("functionSize rule %q: maxEffectiveLines must be >= 1", fn.Name)
		}
	}

	validKinds := map[string]bool{
		"file-size":     true,
		"function-size": true,
	}
	for i, ex := range rs.Exemptions {
		if !validKinds[ex.Kind] {
			return fmt.Errorf("exemption [%d]: kind %q must be one of: file-size, function-size", i, ex.Kind)
		}
		for _, bn := range ex.Match.Basename {
			if strings.ContainsAny(bn, "/\\") {
				return fmt.Errorf("exemption [%d]: basename %q must not contain path separators", i, bn)
			}
		}
	}

	return nil
}
