package d2

import (
	"math"
	"strconv"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
	vm "github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

// diagramBackground is d2 theme 0's background (N7), which every edge label
// and boundary label is drawn on.
const diagramBackground = "#FFFFFF"

// TestTextContrast is FR-010 and SC-002: every text line on a card, at its
// role's opacity, meets WCAG AA (4.5) against the card's fill; edge and
// boundary text does against the diagram background.
func TestTextContrast(t *testing.T) {
	t.Parallel()
	for _, k := range arch.AllElementKinds {
		s := vm.StyleFor(vm.RoleElement, string(k), nil)
		for _, role := range []LineRole{roleTitle, roleKind, roleDescription} {
			if r := contrast(s.FontColor, s.Fill, lineStyles[role].opacity); r < 4.5 {
				t.Errorf("%s, line role %d: contrast %.2f, want ≥ 4.5", k, role, r)
			}
		}
	}
	for _, role := range []LineRole{roleEdgeLabel, roleEdgeTech, roleEdgeTags} {
		if r := contrast(vm.EdgePalette.Font, diagramBackground, lineStyles[role].opacity); r < 4.5 {
			t.Errorf("edge line role %d: contrast %.2f, want ≥ 4.5", role, r)
		}
	}
	boundary := vm.StyleFor(vm.RoleSubject, vm.KindSystem, nil)
	if r := contrast(boundary.FontColor, diagramBackground, 1); r < 4.5 {
		t.Errorf("boundary label: contrast %.2f, want ≥ 4.5", r)
	}
}

// TestLineSizes is User Story 4: no line is drawn below 11px at the
// diagram's natural size (d2 draws labels at 16px).
func TestLineSizes(t *testing.T) {
	t.Parallel()
	for role, s := range lineStyles {
		if s.size*16 < 11 {
			t.Errorf("line role %d renders at %.1fpx, want ≥ 11", role, s.size*16)
		}
	}
}

// contrast is the WCAG 2 contrast ratio of a text colour, drawn at the given
// opacity over a background, against that background.
func contrast(text, background string, opacity float64) float64 {
	fg, bg := rgb(text), rgb(background)
	var blended [3]float64
	for i := range fg {
		blended[i] = opacity*fg[i] + (1-opacity)*bg[i]
	}
	a, b := luminance(blended), luminance(bg)
	return (math.Max(a, b) + 0.05) / (math.Min(a, b) + 0.05)
}

func rgb(hex string) [3]float64 {
	var c [3]float64
	for i := range c {
		v, err := strconv.ParseUint(hex[1+2*i:3+2*i], 16, 8)
		if err != nil {
			panic(hex)
		}
		c[i] = float64(v)
	}
	return c
}

func luminance(c [3]float64) float64 {
	var l [3]float64
	for i, v := range c {
		v /= 255
		if v <= 0.04045 {
			l[i] = v / 12.92
		} else {
			l[i] = math.Pow((v+0.055)/1.055, 2.4)
		}
	}
	return 0.2126*l[0] + 0.7152*l[1] + 0.0722*l[2]
}
