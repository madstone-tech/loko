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
	envAddr := arch.NewEnvironmentAddress(env.Name)

	var diags arch.Diagnostics
	claimed := map[string]claimSite{}
	// Keys are collected so reporting order does not depend on map iteration
	// (FR-033).
	var order []string
	dupes := map[string][]claimSite{}

	for _, placed := range env.AllInstances() {
		instAddr := arch.NewInstanceAddress(envAddr, placed.Instance.Name)

		for _, claim := range placed.Instance.Claims {
			if n := claim.Selectors(); n != 1 {
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

			for _, id := range claim.Identifiers() {
				key := string(claim.Kind) + ":" + id
				site := claimSite{instance: instAddr, rng: claim.Range}
				if prev, exists := claimed[key]; exists {
					if _, seen := dupes[key]; !seen {
						order = append(order, key)
						dupes[key] = []claimSite{prev}
					}
					dupes[key] = append(dupes[key], site)
					continue
				}
				claimed[key] = site
			}
		}
	}

	sort.Strings(order)
	for _, key := range order {
		sites := dupes[key]
		var related []arch.RelatedRange
		for _, s := range sites[1:] {
			related = append(related, arch.RelatedRange{
				Message: fmt.Sprintf("also claimed by %s", s.instance),
				Range:   s.rng,
			})
		}
		diags = append(diags, arch.Diagnostic{
			Severity: arch.SeverityError,
			Code:     arch.CodeDuplicateClaim,
			Summary:  "Physical resource claimed twice",
			Detail: fmt.Sprintf("%s is claimed by more than one instance. Reconciliation cannot "+
				"decide which element owns it.", key),
			Address: sites[0].instance,
			Range:   sites[0].rng,
			Related: related,
		})
	}

	return diags
}
