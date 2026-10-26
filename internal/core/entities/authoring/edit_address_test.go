package authoring

import (
	"slices"
	"testing"
)

// TestAddressForms: one form per target an edit may name, and every example
// parses as an address of its target.
func TestAddressForms(t *testing.T) {
	t.Parallel()
	public := []TargetKind{TargetElement, TargetRelationship, TargetEnvironment, TargetGroup,
		TargetInstance, TargetBinding, TargetView, TargetProject}
	var got []TargetKind
	for _, f := range AddressForms {
		got = append(got, f.Target)
		if f.Pattern == "" || f.Example == "" {
			t.Errorf("%s: empty pattern or example", f.Target)
		}
		if _, ok := SplitAddress(f.Target, f.Example); !ok {
			t.Errorf("%s: example %q does not parse", f.Target, f.Example)
		}
	}
	if !slices.Equal(got, public) {
		t.Errorf("forms cover %v, want %v", got, public)
	}
	i := slices.IndexFunc(AddressForms, func(f AddressForm) bool { return f.Target == TargetInstance })
	if want := "deployment.<env>[.node.<node>…].instance.<name>"; AddressForms[i].Pattern != want {
		t.Errorf("instance pattern = %q, want %q", AddressForms[i].Pattern, want)
	}
	if AddressForms[i].Note == "" {
		t.Error("instance form does not say how an existing instance is addressed")
	}
}
