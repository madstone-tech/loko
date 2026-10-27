package usecases

import (
	"reflect"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
	"github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

// deepEnv nests groups five levels deep with instances at levels 1, 3 and 5,
// mirroring testdata/projects/deployment-nested.
func deepEnvIR() *arch.IR {
	env := arch.NewEnvironmentAddress("prod")
	g := func(path ...string) arch.Address { return arch.NewGroupAddress(env, path) }
	inst := func(name string) arch.Address { return arch.NewInstanceAddress(env, name) }
	groups := []arch.Group{{
		Address: g("l1"), Name: "l1", Contains: []arch.Address{inst("edge")},
		Groups: []arch.Group{{
			Address: g("l1", "l2"), Name: "l2",
			Groups: []arch.Group{{
				Address: g("l1", "l2", "l3"), Name: "l3", Contains: []arch.Address{inst("svc")},
				Groups: []arch.Group{{
					Address: g("l1", "l2", "l3", "l4"), Name: "l4",
					Groups: []arch.Group{{
						Address: g("l1", "l2", "l3", "l4", "l5"), Name: "l5",
						Contains: []arch.Address{inst("store")},
					}},
				}},
			}},
		}},
	}}
	return irOf(
		[]arch.Element{
			el(arch.KindSystem, "core", ""),
			el(arch.KindContainer, "edge", "system.core"),
			el(arch.KindContainer, "svc", "system.core"),
			el(arch.KindContainer, "store", "system.core"),
			el(arch.KindContainer, "reporting", "system.core"),
			el(arch.KindComponent, "inner", "container.svc"),
		},
		[]arch.Relationship{
			rl("container.edge", "svc", "container.svc", "calls", ""),
			rl("component.inner", "store", "container.store", "stores", ""), // lifts to the svc instance
			rl("container.svc", "report", "container.reporting", "reports", ""),
		},
		arch.Environment{Address: env, Name: "prod", Groups: groups, Instances: []arch.Instance{
			{Address: inst("edge"), Name: "edge", Of: "container.edge", PlacedIn: g("l1")},
			{Address: inst("store"), Name: "store", Of: "container.store", PlacedIn: g("l1", "l2", "l3", "l4", "l5")},
			{Address: inst("svc"), Name: "svc", Of: "container.svc", PlacedIn: g("l1", "l2", "l3")},
		}},
	)
}

func TestProjectDeployment(t *testing.T) {
	t.Parallel()
	vm := projectOne(t, deepEnvIR(), "deployment-prod")
	idx := nodeIndex(vm)

	env := "deployment__prod"
	chain := []string{env, env + "__node__l1", env + "__node__l1__l2", env + "__node__l1__l2__l3",
		env + "__node__l1__l2__l3__l4", env + "__node__l1__l2__l3__l4__l5"}
	if idx[env].Role != viewmodel.RoleSubject {
		t.Fatalf("environment node = %+v", idx[env])
	}
	for i := 1; i < len(chain); i++ {
		n := idx[chain[i]]
		if n.Role != viewmodel.RoleGroup || n.Parent != chain[i-1] {
			t.Errorf("group level %d = %+v, want parent %s", i, n, chain[i-1])
		}
	}
	for inst, parent := range map[string]string{
		env + "__instance__edge":  chain[1],
		env + "__instance__svc":   chain[3],
		env + "__instance__store": chain[5],
	} {
		n := idx[inst]
		if n.Role != viewmodel.RoleInstance || n.Parent != parent || n.Kind != "container" {
			t.Errorf("instance %s = %+v, want parent %s", inst, n, parent)
		}
		if !reflect.DeepEqual(n.Style, viewmodel.StyleFor(viewmodel.RoleInstance, "container", nil)) {
			t.Errorf("instance %s style = %+v", inst, n.Style)
		}
	}

	want := []string{
		env + "__instance__edge--" + env + "__instance__svc",
		env + "__instance__svc--" + env + "__instance__store", // component lifted to its instantiated container
		env + "__instance__svc--outside",                      // reporting has no instance in prod
	}
	if got := edgeIDs(vm); !reflect.DeepEqual(got, want) {
		t.Errorf("edges = %v, want %v", got, want)
	}
	if _, ok := idx[viewmodel.OutsideNodeID]; !ok {
		t.Error("missing outside node for the boundary edge")
	}
}

// TestProjectDeploymentDetails: groups carry the node attributes, and an
// instance's own technology replaces its element's on this diagram only
// (feature 018).
func TestProjectDeploymentDetails(t *testing.T) {
	t.Parallel()
	ir := deepEnvIR()
	env := &ir.Environments[0]
	l1 := &env.Groups[0]
	l1.Title, l1.Technology, l1.Description, l1.Tags = "AWS account", "AWS", "Production account", []string{"cloud"}
	env.Instances[0].Technology = "local process" // edge
	for i := range ir.Elements {
		if ir.Elements[i].Name == "edge" || ir.Elements[i].Name == "svc" {
			ir.Elements[i].Technology = "Lambda"
		}
	}
	idx := nodeIndex(projectOne(t, ir, "deployment-prod"))
	g := idx["deployment__prod__node__l1"]
	if g.Title != "AWS account" || g.Technology != "AWS" || g.Description != "Production account" || !reflect.DeepEqual(g.Tags, []string{"cloud"}) {
		t.Errorf("group = %+v", g)
	}
	if l2 := idx["deployment__prod__node__l1__l2"]; l2.Title != "" || l2.Technology != "" || l2.Tags != nil {
		t.Errorf("nested group inherited attributes: %+v", l2)
	}
	if got := idx["deployment__prod__instance__edge"].Technology; got != "local process" {
		t.Errorf("edge instance technology = %q, want the instance's", got)
	}
	if got := idx["deployment__prod__instance__svc"].Technology; got != "Lambda" {
		t.Errorf("svc instance technology = %q, want the element's", got)
	}
}

func TestProjectDeploymentManyInstances(t *testing.T) {
	t.Parallel()
	env := arch.NewEnvironmentAddress("e")
	ir := irOf(
		[]arch.Element{el(arch.KindContainer, "a", ""), el(arch.KindContainer, "b", "")},
		[]arch.Relationship{rl("container.a", "b", "container.b", "", "")},
		arch.Environment{Address: env, Name: "e", Instances: []arch.Instance{
			{Address: arch.NewInstanceAddress(env, "a1"), Name: "a1", Of: "container.a"},
			{Address: arch.NewInstanceAddress(env, "a2"), Name: "a2", Of: "container.a"},
			{Address: arch.NewInstanceAddress(env, "b1"), Name: "b1", Of: "container.b"},
		}},
	)
	vm := projectOne(t, ir, "deployment-e")
	if got := len(vm.Edges); got != 2 {
		t.Errorf("an element instantiated twice draws an edge from each instance: got %d edges %v", got, edgeIDs(vm))
	}
}
