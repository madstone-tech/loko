package arch

import (
	"sort"
	"testing"
)

// TestAddressConstructors covers the six canonical address forms fixed in
// specs/013-hcl-compiler-core/data-model.md §1. These forms are frozen for
// v1.x: the diff stage compares by address, so changing one silently would
// make every historical comparison wrong.
func TestAddressConstructors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		got  Address
		want Address
	}{
		{"element", NewElementAddress(KindContainer, "api"), "container.api"},
		{"person", NewElementAddress(KindPerson, "customer"), "person.customer"},
		{"external", NewElementAddress(KindExternal, "stripe"), "external.stripe"},
		{"relationship", NewRelationshipAddress(NewElementAddress(KindContainer, "api"), "orders"), "container.api.uses.orders"},
		{"environment", NewEnvironmentAddress("prod"), "deployment.prod"},
		{"group one level", NewGroupAddress(NewEnvironmentAddress("prod"), []string{"vpc-main"}), "deployment.prod.node.vpc-main"},
		{"group nested", NewGroupAddress(NewEnvironmentAddress("prod"), []string{"vpc-main", "subnet-a"}), "deployment.prod.node.vpc-main.subnet-a"},
		{"instance", NewInstanceAddress(NewEnvironmentAddress("prod"), "api"), "deployment.prod.instance.api"},
		{"view", NewViewAddress("payment-path"), "view.payment-path"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}

// TestInstanceAddressIgnoresPlacement is the clarification of 2026-09-22
// (FR-012, FR-024): an instance's address must not carry its placement-group
// path, so moving an instance between groups preserves its identity and the
// diff stage reads a re-parent as a change rather than a remove plus an add.
func TestInstanceAddressIgnoresPlacement(t *testing.T) {
	t.Parallel()

	env := NewEnvironmentAddress("prod")
	shallow := NewInstanceAddress(env, "api")
	deep := NewInstanceAddress(env, "api")

	if shallow != deep {
		t.Fatalf("instance address depends on placement: %q != %q", shallow, deep)
	}
	if got, want := shallow, Address("deployment.prod.instance.api"); got != want {
		t.Errorf("got %q, want %q — no node segment may appear", got, want)
	}
}

func TestValidName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		valid bool
	}{
		{"api", true},
		{"orders_db", true},
		{"vpc-main", true},
		{"_internal", true},
		{"API2", true},
		{"a", true},
		{"", false},
		{"2fast", false},
		{"-leading", false},
		{"has space", false},
		{"has.dot", false},
		{"has/slash", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := ValidName(tt.name); got != tt.valid {
				t.Errorf("ValidName(%q) = %v, want %v", tt.name, got, tt.valid)
			}
		})
	}
}

// TestAddressOrderingIsBytewise guards FR-040. Ordering must not depend on
// locale collation, or the same source would export differently on two
// machines and the golden-file suite would be unwritable.
func TestAddressOrderingIsBytewise(t *testing.T) {
	t.Parallel()

	got := []Address{
		"container.beta",
		"container.Alpha",
		"container.alpha",
		"container.api-2",
		"container.api_1",
	}
	sort.Slice(got, func(i, j int) bool { return got[i].Compare(got[j]) < 0 })

	// Byte order: uppercase letters (0x41-) sort before lowercase (0x61-),
	// and '-' (0x2D) before '_' (0x5F).
	want := []Address{
		"container.Alpha",
		"container.alpha",
		"container.api-2",
		"container.api_1",
		"container.beta",
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("byte-wise order wrong at %d: got %v, want %v", i, got, want)
		}
	}
}

func TestAddressKind(t *testing.T) {
	t.Parallel()

	tests := []struct {
		addr Address
		want string
	}{
		{"container.api", "container"},
		{"container.api.uses.orders", "container"},
		{"deployment.prod.instance.api", "deployment"},
		{"view.payment-path", "view"},
		{"", ""},
		{"nodot", "nodot"},
	}

	for _, tt := range tests {
		t.Run(string(tt.addr), func(t *testing.T) {
			t.Parallel()
			if got := tt.addr.Prefix(); got != tt.want {
				t.Errorf("Address(%q).Prefix() = %q, want %q", tt.addr, got, tt.want)
			}
		})
	}
}
