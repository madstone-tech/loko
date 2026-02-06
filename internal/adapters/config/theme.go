package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/madstone-tech/loko/internal/core/entities"
	toml "github.com/pelletier/go-toml/v2"
)

// ThemeStore implements the ThemeLoader interface.
type ThemeStore struct {
	themesDir string
}

// NewThemeStore creates a theme store using the given themes directory.
func NewThemeStore(themesDir string) *ThemeStore {
	return &ThemeStore{themesDir: themesDir}
}

// tomlTheme represents the TOML structure of a theme file.
type tomlTheme struct {
	Theme  tomlThemeInfo     `toml:"theme"`
	Colors map[string]string `toml:"colors"`
	Styles map[string]string `toml:"styles"`
}

type tomlThemeInfo struct {
	Name    string `toml:"name"`
	D2Theme string `toml:"d2_theme"`
}

// LoadTheme loads a theme by name from the themes directory.
func (s *ThemeStore) LoadTheme(ctx context.Context, name string) (*entities.Theme, error) {
	path := filepath.Join(s.themesDir, name+".toml")

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("theme %q not found: %w", name, err)
	}

	var tt tomlTheme
	if err := toml.Unmarshal(data, &tt); err != nil {
		return nil, fmt.Errorf("failed to parse theme %q: %w", name, err)
	}

	theme, err := entities.NewTheme(name)
	if err != nil {
		return nil, err
	}

	theme.Path = path
	theme.D2Theme = tt.Theme.D2Theme

	if tt.Colors != nil {
		theme.Colors = tt.Colors
	}
	if tt.Styles != nil {
		theme.Styles = tt.Styles
	}

	return theme, nil
}

// ListThemes returns the names of all available themes.
func (s *ThemeStore) ListThemes(ctx context.Context) ([]string, error) {
	entries, err := os.ReadDir(s.themesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No themes dir = no themes
		}
		return nil, fmt.Errorf("failed to read themes directory: %w", err)
	}

	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".toml") {
			names = append(names, strings.TrimSuffix(name, ".toml"))
		}
	}
	return names, nil
}
