package d2

import (
	"strings"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
	vm "github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

// TestIconTableComplete is FR-017: every shape and kind the language accepts
// has its own icon, so a new one added without an icon fails here rather
// than silently falling back to the system icon.
func TestIconTableComplete(t *testing.T) {
	t.Parallel()
	for _, s := range arch.Shapes {
		if _, ok := vm.ShapeIcons[vm.Shape(s)]; !ok {
			t.Errorf("shape %q has no icon", s)
		}
	}
	for _, k := range arch.AllElementKinds {
		if _, ok := vm.KindIcons[string(k)]; !ok {
			t.Errorf("kind %q has no icon", k)
		}
	}
}

func TestIconFor(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		shape vm.Shape
		kind  string
		want  string
	}{
		{vm.ShapeDatabase, vm.KindContainer, "database"},
		{vm.ShapeQueue, vm.KindContainer, "list-ordered"},
		{vm.ShapeTopic, vm.KindExternal, "megaphone"},
		{vm.ShapeFunction, vm.KindContainer, "square-function"},
		{vm.ShapeBucket, vm.KindExternal, "archive"},
		{vm.ShapePerson, vm.KindPerson, "user"},
		{vm.ShapeRectangle, vm.KindSystem, "boxes"},
		{vm.ShapeRectangle, vm.KindContainer, "container"},
		{vm.ShapeRectangle, vm.KindComponent, "blocks"},
		{vm.ShapeRectangle, vm.KindExternal, "globe"},
		{vm.ShapeRectangle, "widget", "boxes"},
	} {
		if got := iconFor(tt.shape, tt.kind); got != tt.want {
			t.Errorf("iconFor(%s, %s) = %s, want %s", tt.shape, tt.kind, got, tt.want)
		}
	}
}

// TestIconSymbols: each icon is the Lucide file's drawing wrapped in a
// symbol, with nothing that references a file or the network (FR-004).
func TestIconSymbols(t *testing.T) {
	t.Parallel()
	names := map[string]bool{vm.FallbackIcon: true}
	for _, n := range vm.ShapeIcons {
		names[n] = true
	}
	for _, n := range vm.KindIcons {
		names[n] = true
	}
	for name := range names {
		s := iconSymbol(name)
		// Lucide's stroke attributes, with the stroke in the <use>'s color.
		prefix := `<symbol id="loko-icon-` + name + `" viewBox="0 0 24 24"><g fill="none" stroke="currentColor"` +
			` stroke-width="2" stroke-linecap="round" stroke-linejoin="round">`
		if !strings.HasPrefix(s, prefix) || !strings.HasSuffix(s, "</g></symbol>") {
			t.Errorf("%s: symbol = %.200q", name, s)
		}
		for _, bad := range []string{"<svg", "xmlns", "http", "\n"} {
			if strings.Contains(s, bad) {
				t.Errorf("%s: symbol contains %q", name, bad)
			}
		}
		if !strings.Contains(s, "<path") && !strings.Contains(s, "<rect") && !strings.Contains(s, "<circle") {
			t.Errorf("%s: symbol draws nothing: %q", name, s)
		}
	}
}
