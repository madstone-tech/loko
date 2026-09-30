package usecases

import (
	"reflect"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

func file(name string, line int) arch.SourceRange {
	return arch.SourceRange{File: name, StartLine: line, StartColumn: 1}
}

func TestBuildProvenance(t *testing.T) {
	t.Parallel()
	model := &arch.SourceModel{
		Files: []string{"b.loko.hcl", "a.loko.hcl", "deploy.loko.hcl"},
		Elements: []arch.ElementDecl{
			{Kind: arch.KindSystem, Name: "s", Range: file("a.loko.hcl", 1),
				Relations: []arch.RelationDecl{{LocalName: "r", Range: file("a.loko.hcl", 2)}}},
			{Kind: arch.KindContainer, Name: "c", Range: file("b.loko.hcl", 3)},
		},
		Environments: []arch.EnvironmentDecl{{
			Name: "prod", Range: file("deploy.loko.hcl", 1),
			Instances: []arch.InstanceDecl{{Name: "top", Range: file("deploy.loko.hcl", 2)}},
			Groups: []arch.GroupDecl{{
				Name: "vpc", Range: file("deploy.loko.hcl", 3),
				Groups: []arch.GroupDecl{{
					Name: "sub", Range: file("deploy.loko.hcl", 4),
					Instances: []arch.InstanceDecl{{Name: "deep", Range: file("deploy.loko.hcl", 5)}},
				}},
			}},
		}},
		Views: []arch.ViewDecl{{Name: "v", Range: file("a.loko.hcl", 9)}},
	}
	p := BuildProvenance(model)

	env := arch.NewEnvironmentAddress("prod")
	for addr, want := range map[arch.Address]arch.SourceRange{
		"system.s":                          file("a.loko.hcl", 1),
		"system.s.uses.r":                   file("a.loko.hcl", 2),
		"container.c":                       file("b.loko.hcl", 3),
		env:                                 file("deploy.loko.hcl", 1),
		arch.NewInstanceAddress(env, "top"): file("deploy.loko.hcl", 2),
		arch.NewGroupAddress(env, []string{"vpc"}):        file("deploy.loko.hcl", 3),
		arch.NewGroupAddress(env, []string{"vpc", "sub"}): file("deploy.loko.hcl", 4),
		arch.NewInstanceAddress(env, "deep"):              file("deploy.loko.hcl", 5),
		arch.NewViewAddress("v"):                          file("a.loko.hcl", 9),
	} {
		got, ok := p.RangeOf(addr)
		if !ok || got != want {
			t.Errorf("RangeOf(%s) = %+v, %v; want %+v", addr, got, ok, want)
		}
	}
	if _, ok := p.RangeOf("system.nope"); ok {
		t.Error("RangeOf(unknown) reported found")
	}

	got := p.SourcesFor([]arch.Address{"container.c", "system.s", "system.s.uses.r", "nope"})
	if want := []string{"a.loko.hcl", "b.loko.hcl"}; !reflect.DeepEqual(got, want) {
		t.Errorf("SourcesFor = %v, want %v", got, want)
	}
	if want := []string{"a.loko.hcl", "b.loko.hcl", "deploy.loko.hcl"}; !reflect.DeepEqual(p.AllFiles(), want) {
		t.Errorf("AllFiles = %v, want %v", p.AllFiles(), want)
	}
}
