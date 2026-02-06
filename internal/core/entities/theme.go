package entities

import (
	"fmt"
	"regexp"
	"strings"
)

// hexColorPattern matches valid hex color codes: #RGB or #RRGGBB (case-insensitive).
var hexColorPattern = regexp.MustCompile(`^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)

// Theme represents a visual theme configuration for diagram rendering.
// The theme name is derived from its filename (e.g., "dark" from "dark.toml").
type Theme struct {
	// Name is derived from the filename (no path separators allowed)
	Name string

	// Path is the absolute path to the theme file
	Path string

	// D2Theme is the D2 theme name override (optional, must be non-empty if set)
	D2Theme string

	// Colors holds color overrides as key -> hex value mappings
	Colors map[string]string

	// Styles holds D2 style overrides
	Styles map[string]string
}

// NewTheme creates a new Theme with the given name.
// The name must not be empty, must not contain path separators, and must
// match the standard name pattern (alphanumeric, hyphens, underscores, spaces).
func NewTheme(name string) (*Theme, error) {
	if err := validateThemeName(name); err != nil {
		return nil, err
	}

	return &Theme{
		Name:   name,
		Colors: make(map[string]string),
		Styles: make(map[string]string),
	}, nil
}

// Validate checks that the theme configuration is valid.
func (t *Theme) Validate() error {
	var errs ValidationErrors

	if err := validateThemeName(t.Name); err != nil {
		errs = append(errs, err)
	}

	// D2Theme must be non-empty if set (a whitespace-only string is invalid)
	if t.D2Theme != "" && strings.TrimSpace(t.D2Theme) == "" {
		errs.Add("Theme", "D2Theme", t.D2Theme, "D2 theme name cannot be whitespace-only", nil)
	}

	// Validate all color values are valid hex codes
	for key, color := range t.Colors {
		if !hexColorPattern.MatchString(color) {
			errs.Add("Theme", "Colors", fmt.Sprintf("%s=%s", key, color),
				fmt.Sprintf("invalid hex color for key %q: must be #RGB or #RRGGBB", key), nil)
		}
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

// SetD2Theme sets the D2 theme name override.
func (t *Theme) SetD2Theme(theme string) {
	t.D2Theme = theme
}

// SetColor sets a color override. The color must be a valid hex code (#RGB or #RRGGBB).
func (t *Theme) SetColor(key, color string) error {
	if !hexColorPattern.MatchString(color) {
		return NewValidationError("Theme", "Colors",
			fmt.Sprintf("%s=%s", key, color),
			fmt.Sprintf("invalid hex color for key %q: must be #RGB or #RRGGBB", key), nil)
	}
	t.Colors[key] = color
	return nil
}

// GetColor retrieves a color override by key.
func (t *Theme) GetColor(key string) (string, bool) {
	color, exists := t.Colors[key]
	return color, exists
}

// SetStyle sets a D2 style override.
func (t *Theme) SetStyle(key, value string) {
	t.Styles[key] = value
}

// GetStyle retrieves a D2 style override by key.
func (t *Theme) GetStyle(key string) (string, bool) {
	style, exists := t.Styles[key]
	return style, exists
}

// ColorCount returns the number of color overrides.
func (t *Theme) ColorCount() int {
	return len(t.Colors)
}

// StyleCount returns the number of style overrides.
func (t *Theme) StyleCount() int {
	return len(t.Styles)
}

// validateThemeName checks that a theme name is valid.
// Theme names must pass the standard name validation and must not contain
// path separators (/ or \), since the name is derived from a filename.
func validateThemeName(name string) *ValidationError {
	if err := ValidateName(name); err != nil {
		return NewValidationError("Theme", "Name", name, "invalid name", err)
	}

	if strings.ContainsAny(name, `/\`) {
		return NewValidationError("Theme", "Name", name,
			"name must not contain path separators (derived from filename)", ErrInvalidName)
	}

	return nil
}
