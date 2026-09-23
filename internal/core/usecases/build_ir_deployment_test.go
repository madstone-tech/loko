package usecases

import (
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// env builds an environment with one instance at the given node path.
func envWithInstance(name string, path []string, instName, of string) arch.EnvironmentDecl {
	inst := arch.InstanceDecl{Name: instName, Range: at(20), Of: ref(of, 21)}

	if len(path) == 0 {
		return arch.EnvironmentDecl{Name: name, Range: at(1), Instances: []arch.InstanceDecl{inst}}
	}
	// Build the node chain from the inside out.
	group := arch.GroupDecl{Name: path[len(path)-1], Range: at(10), Instances: []arch.InstanceDecl{inst}}
	for i := len(path) - 2; i >= 0; i-- {
		group = arch.GroupDecl{Name: path[i], Range: at(10), Groups: []arch.GroupDecl{group}}
	}
	return arch.EnvironmentDecl{Name: name, Range: at(1), Groups: []arch.GroupDecl{group}}
}

func modelWith(env arch.EnvironmentDecl) *arch.SourceModel {
	return &arch.SourceModel{
		Elements:     []arch.ElementDecl{elem(arch.KindContainer, "api", 1)},
		Environments: []arch.EnvironmentDecl{env},
	}
}

// TestInstanceAddressOmitsPlacementPath is the clarification of 2026-09-22
// (FR-012, FR-024) expressed end to end.
func TestInstanceAddressOmitsPlacementPath(t *testing.T) {
	t.Parallel()

	model := modelWith(envWithInstance("prod", []string{"vpc-main", "subnet-a"}, "api", "container.api"))
	res, diags := ResolveModel(model)
	if diags.HasErrors() {
		t.Fatalf("unexpected diagnostics: %v", codes(diags))
	}

	ir := BuildIR(model, res)
	inst, ok := ir.Instance("deployment.prod.instance.api")
	if !ok {
		t.Fatalf("instance not addressable as deployment.prod.instance.api; got %+v",
			ir.Environments[0].Instances)
	}
	if got, want := inst.PlacedIn, arch.Address("deployment.prod.node.vpc-main.subnet-a"); got != want {
		t.Errorf("placedIn = %q, want %q", got, want)
	}
	if got, want := inst.Of, arch.Address("container.api"); got != want {
		t.Errorf("of = %q, want %q", got, want)
	}

	// The group tree is preserved in full and cross-references the instance.
	env := ir.Environments[0]
	if len(env.Groups) != 1 || len(env.Groups[0].Groups) != 1 {
		t.Fatalf("node tree not preserved: %+v", env.Groups)
	}
	inner := env.Groups[0].Groups[0]
	if got, want := inner.Address, arch.Address("deployment.prod.node.vpc-main.subnet-a"); got != want {
		t.Errorf("group address = %q, want %q", got, want)
	}
	if len(inner.Contains) != 1 || inner.Contains[0] != inst.Address {
		t.Errorf("group.contains = %v, want [%s]", inner.Contains, inst.Address)
	}
}

// TestReParentingPreservesIdentity is the property the diff stage depends on:
// moving an instance between nodes must NOT look like a delete plus a create.
func TestReParentingPreservesIdentity(t *testing.T) {
	t.Parallel()

	deep := modelWith(envWithInstance("prod", []string{"vpc-main", "subnet-a"}, "api", "container.api"))
	shallow := modelWith(envWithInstance("prod", []string{"vpc-main"}, "api", "container.api"))
	flat := modelWith(envWithInstance("prod", nil, "api", "container.api"))

	var addrs []arch.Address
	var placements []arch.Address
	for _, m := range []*arch.SourceModel{deep, shallow, flat} {
		res, _ := ResolveModel(m)
		inst := BuildIR(m, res).Environments[0].Instances[0]
		addrs = append(addrs, inst.Address)
		placements = append(placements, inst.PlacedIn)
	}

	for i := 1; i < len(addrs); i++ {
		if addrs[i] != addrs[0] {
			t.Errorf("address changed with placement: %q vs %q — a re-parent would read "+
				"as a remove plus an add in the diff stage", addrs[i], addrs[0])
		}
	}
	if placements[0] == placements[1] || placements[1] == placements[2] {
		t.Errorf("placedIn did not change with placement: %v", placements)
	}
	if placements[2] != "" {
		t.Errorf("an instance directly under the environment has placedIn %q, want empty", placements[2])
	}
}

// TestDuplicateInstanceNameAcrossNodes covers FR-012a: names are unique per
// environment, not per node, precisely because the address omits the path.
func TestDuplicateInstanceNameAcrossNodes(t *testing.T) {
	t.Parallel()

	model := &arch.SourceModel{
		Elements: []arch.ElementDecl{elem(arch.KindContainer, "api", 1)},
		Environments: []arch.EnvironmentDecl{{
			Name: "prod", Range: at(1),
			Groups: []arch.GroupDecl{
				{Name: "vpc-a", Range: at(10), Instances: []arch.InstanceDecl{
					{Name: "api", Range: at(11), Of: ref("container.api", 12)}}},
				{Name: "vpc-b", Range: at(20), Instances: []arch.InstanceDecl{
					{Name: "api", Range: at(21), Of: ref("container.api", 22)}}},
			},
		}},
	}

	_, diags := ResolveModel(model)
	d := findCode(diags, arch.CodeDuplicateInstanceName)
	if d == nil {
		t.Fatalf("two instances named api in one environment were accepted: %v", codes(diags))
	}
	// The instance-specific code exists so a consumer can tell this apart from
	// any other duplicate; the message has to explain why two names in
	// different nodes collide at all.
	if !contains(d.Detail, "unique per deployment") {
		t.Errorf("detail does not explain per-environment uniqueness: %q", d.Detail)
	}
	if len(d.Related) == 0 {
		t.Error("the diagnostic does not name the first declaration")
	}
}

// TestDuplicateClaimAcrossInstances covers FR-028.
func TestDuplicateClaimAcrossInstances(t *testing.T) {
	t.Parallel()

	claim := func(addr string) []arch.ClaimDecl {
		return []arch.ClaimDecl{{Kind: arch.ClaimTerraform, Address: addr, Range: at(30)}}
	}
	model := &arch.SourceModel{
		Elements: []arch.ElementDecl{elem(arch.KindContainer, "api", 1)},
		Environments: []arch.EnvironmentDecl{{
			Name: "prod", Range: at(1),
			Instances: []arch.InstanceDecl{
				{Name: "a", Range: at(10), Of: ref("container.api", 11), Claims: claim("module.x")},
				{Name: "b", Range: at(20), Of: ref("container.api", 21), Claims: claim("module.x")},
			},
		}},
	}

	res, _ := ResolveModel(model)
	diags := ValidateDeployment(model, res)

	d := findCode(diags, arch.CodeDuplicateClaim)
	if d == nil {
		t.Fatalf("the same physical address claimed twice was accepted: %v", codes(diags))
	}
	if len(d.Related) == 0 {
		t.Error("the diagnostic does not name the other claimant")
	}
}

func TestBindingSelectorArity(t *testing.T) {
	t.Parallel()

	tests := map[string]arch.ClaimDecl{
		"no selector":   {Kind: arch.ClaimTerraform, Range: at(30)},
		"two selectors": {Kind: arch.ClaimTerraform, Address: "a", Addresses: []string{"b"}, Range: at(30)},
	}
	for name, c := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			model := &arch.SourceModel{Environments: []arch.EnvironmentDecl{{
				Name: "prod", Range: at(1),
				Instances: []arch.InstanceDecl{{Name: "a", Range: at(10), Claims: []arch.ClaimDecl{c}}},
			}}}
			if !hasCode(ValidateDeployment(model, &Resolved{}), arch.CodeUnknownAttribute) {
				t.Error("a malformed binding selector was accepted")
			}
		})
	}
}
