package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"

	"gopkg.in/yaml.v3"
)

// Suppression is a single time-bound, owner-tagged exemption for a pre-existing
// violation that cannot be fixed within the current feature's scope.
type Suppression struct {
	Rule      string `yaml:"rule"       json:"rule"`
	File      string `yaml:"file"       json:"file"`
	Function  string `yaml:"function"   json:"function,omitempty"`
	Owner     string `yaml:"owner"      json:"owner"`
	ExpiresOn string `yaml:"expires_on" json:"expires_on"`
	Reason    string `yaml:"reason"     json:"reason"`
	Notes     string `yaml:"notes"      json:"notes,omitempty"`
}

const (
	suppressionMaxExpiryDays    = 90
	suppressionStaleWarningDays = 30
)

var ownerHandleRe = regexp.MustCompile(`^@[A-Za-z0-9][A-Za-z0-9-]*$`)

// LoadSuppressions reads and validates the suppression file at path. It returns
// the parsed list, a (possibly empty) list of validation errors, and an I/O
// error if the file could not be read. A missing file is not an error — it
// yields an empty list. `now` is injected so tests can pin time.
func LoadSuppressions(path string, now time.Time, knownRules map[string]bool) ([]Suppression, []error, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("read %s: %w", path, err)
	}

	var entries []Suppression
	if err := yaml.Unmarshal(data, &entries); err != nil {
		return nil, nil, fmt.Errorf("parse %s: %w", path, err)
	}

	var validationErrs []error
	for i, s := range entries {
		idx := i + 1
		if s.Rule == "" {
			validationErrs = append(validationErrs, fmt.Errorf("entry #%d: rule is required", idx))
		} else if knownRules != nil && !knownRules[s.Rule] {
			validationErrs = append(validationErrs, fmt.Errorf("entry #%d: rule %q does not match any rule or budget in the rules file", idx, s.Rule))
		}
		if s.File == "" {
			validationErrs = append(validationErrs, fmt.Errorf("entry #%d: file is required", idx))
		}
		if !ownerHandleRe.MatchString(s.Owner) {
			validationErrs = append(validationErrs, fmt.Errorf("entry #%d: owner %q must match @github-handle", idx, s.Owner))
		}
		if len(s.Reason) < 20 {
			validationErrs = append(validationErrs, fmt.Errorf("entry #%d: reason must be at least 20 characters", idx))
		}
		expiry, err := time.Parse("2006-01-02", s.ExpiresOn)
		if err != nil {
			validationErrs = append(validationErrs, fmt.Errorf("entry #%d: expires_on %q is not YYYY-MM-DD", idx, s.ExpiresOn))
			continue
		}
		maxExpiry := now.AddDate(0, 0, suppressionMaxExpiryDays)
		if expiry.After(maxExpiry) {
			validationErrs = append(validationErrs, fmt.Errorf("entry #%d: expires_on %s is more than %d days from today; longer-lived suppressions require an ADR", idx, s.ExpiresOn, suppressionMaxExpiryDays))
		}
	}

	return entries, validationErrs, nil
}

// ApplySuppressions partitions violations into (kept, suppressed) using the
// provided suppression set evaluated at `now`. Expired suppressions are NOT
// applied — their would-be matches remain in the kept list as normal failures.
func ApplySuppressions(violations []Violation, suppressions []Suppression, now time.Time) (kept, suppressed []Violation, stale []Suppression) {
	if len(suppressions) == 0 {
		return violations, nil, nil
	}

	for i := range violations {
		v := violations[i]
		matched := false
		for _, s := range suppressions {
			if !suppressionMatches(s, v) {
				continue
			}
			expiry, err := time.Parse("2006-01-02", s.ExpiresOn)
			if err != nil {
				continue
			}
			if now.After(expiry) {
				// Expired suppression: do not suppress; flag staleness.
				if now.Sub(expiry) > time.Duration(suppressionStaleWarningDays)*24*time.Hour {
					stale = appendUniqueSuppression(stale, s)
				}
				continue
			}
			matched = true
			break
		}
		if matched {
			suppressed = append(suppressed, v)
		} else {
			kept = append(kept, v)
		}
	}

	return kept, suppressed, stale
}

// suppressionMatches reports whether the given suppression covers the given
// violation. The file field supports glob patterns relative to the repo root.
// The function field, when set, must equal the violation's subject.
func suppressionMatches(s Suppression, v Violation) bool {
	if s.Rule != v.Rule {
		return false
	}
	if s.Function != "" && s.Function != v.Subject {
		return false
	}
	matched, err := filepath.Match(s.File, v.File)
	if err == nil && matched {
		return true
	}
	// Fall back to literal-equality for non-glob entries.
	return s.File == v.File
}

func appendUniqueSuppression(in []Suppression, s Suppression) []Suppression {
	for _, existing := range in {
		if existing.Rule == s.Rule && existing.File == s.File && existing.Function == s.Function {
			return in
		}
	}
	out := append(in, s)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].File != out[j].File {
			return out[i].File < out[j].File
		}
		return out[i].Rule < out[j].Rule
	})
	return out
}

// KnownRuleNames returns the set of all rule + budget names defined in the
// loaded RuleSet. Used to validate that suppression entries reference real rules.
func KnownRuleNames(rules *RuleSet) map[string]bool {
	out := make(map[string]bool)
	for _, r := range rules.Layers {
		out[r.Name] = true
	}
	for _, r := range rules.FileSizes {
		out[r.Name] = true
	}
	for _, r := range rules.FunctionSizes {
		out[r.Name] = true
	}
	return out
}
