package usecases

import (
	"fmt"
	"sort"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// ValidateDeployment checks the deployment-plane rules: no two claims over the
// same physical identifier, and a well-formed binding selector (FR-028).
//
// Duplicate instance names within an environment are caught earlier, by the
// symbol table, because an instance address does not include its placement
// path — two instances of the same name in different nodes collide on one
// address (FR-012a).
func ValidateDeployment(model *arch.SourceModel, _ *Resolved) arch.Diagnostics {
	var diags arch.Diagnostics
	for _, env := range model.Environments {
		diags = append(diags, checkClaims(env)...)
	}
	return diags
}

// claimSite records where a physical identifier was claimed.
type claimSite struct {
	instance arch.Address
	rng      arch.SourceRange
}

func checkClaims(env arch.EnvironmentDecl) arch.Diagnostics {
	var diags arch.Diagnostics
	diags = append(diags, checkSelectorArity(env)...)
	diags = append(diags, checkDuplicateClaims(env)...)
	return diags
}

// checkSelectorArity reports a binding that selects by none, or more than one,
// of address / addresses / tags.
func checkSelectorArity(env arch.EnvironmentDecl) arch.Diagnostics {
	envAddr := arch.NewEnvironmentAddress(env.Name)

	var diags arch.Diagnostics
	for _, placed := range env.AllInstances() {
		instAddr := arch.NewInstanceAddress(envAddr, placed.Instance.Name)
		for _, claim := range placed.Instance.Claims {
			n := claim.Selectors()
			if n == 1 {
				continue
			}
			diags = append(diags, arch.Diagnostic{
				Severity: arch.SeverityError,
				Code:     arch.CodeUnknownAttribute,
				Summary:  "Binding needs exactly one selector",
				Detail: fmt.Sprintf(
					"A binding selects resources by address, addresses, or tags — exactly one of "+
						"the three. This one has %d.", n),
				Address: instAddr,
				Range:   claim.Range,
			})
		}
	}
	return diags
}

// checkDuplicateClaims reports a physical identifier claimed by more than one
// instance, which leaves reconciliation unable to decide who owns it.
func checkDuplicateClaims(env arch.EnvironmentDecl) arch.Diagnostics {
	envAddr := arch.NewEnvironmentAddress(env.Name)

	claimed := map[string]claimSite{}
	dupes := map[string][]claimSite{}
	// Keys are collected so reporting order does not depend on map iteration
	// (FR-033).
	var order []string

	for _, placed := range env.AllInstances() {
		instAddr := arch.NewInstanceAddress(envAddr, placed.Instance.Name)
		for _, claim := range placed.Instance.Claims {
			for _, id := range claim.Identifiers() {
				key := string(claim.Kind) + ":" + id
				site := claimSite{instance: instAddr, rng: claim.Range}
				prev, exists := claimed[key]
				if !exists {
					claimed[key] = site
					continue
				}
				if _, seen := dupes[key]; !seen {
					order = append(order, key)
					dupes[key] = []claimSite{prev}
				}
				dupes[key] = append(dupes[key], site)
			}
		}
	}

	sort.Strings(order)
	var diags arch.Diagnostics
	for _, key := range order {
		diags = append(diags, duplicateClaimDiagnostic(key, dupes[key]))
	}
	return diags
}

func duplicateClaimDiagnostic(key string, sites []claimSite) arch.Diagnostic {
	var related []arch.RelatedRange
	for _, s := range sites[1:] {
		related = append(related, arch.RelatedRange{
			Message: fmt.Sprintf("also claimed by %s", s.instance),
			Range:   s.rng,
		})
	}
	return arch.Diagnostic{
		Severity: arch.SeverityError,
		Code:     arch.CodeDuplicateClaim,
		Summary:  "Physical resource claimed twice",
		Detail: fmt.Sprintf("%s is claimed by more than one instance. Reconciliation cannot "+
			"decide which element owns it.", key),
		Address: sites[0].instance,
		Range:   sites[0].rng,
		Related: related,
	}
}
