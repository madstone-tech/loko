package entities

import "path/filepath"

// XDGPaths holds resolved XDG-compliant paths for loko application data.
// Path resolution is performed by the PathResolver adapter; this entity
// stores the results as a value object.
type XDGPaths struct {
	// ConfigHome is the resolved configuration directory.
	// Typically ~/.config/loko/ or overridden by LOKO_CONFIG_HOME/XDG_CONFIG_HOME.
	ConfigHome string

	// DataHome is the resolved data directory.
	// Typically ~/.local/share/loko/ or overridden by XDG_DATA_HOME.
	DataHome string

	// CacheHome is the resolved cache directory.
	// Typically ~/.cache/loko/ or overridden by XDG_CACHE_HOME.
	CacheHome string
}

// ConfigFile returns the path to the global config file (config.toml).
func (p XDGPaths) ConfigFile() string {
	return filepath.Join(p.ConfigHome, "config.toml")
}

// ThemesDir returns the path to the themes directory.
func (p XDGPaths) ThemesDir() string {
	return filepath.Join(p.DataHome, "themes")
}

// CacheDir returns the cache directory path (same as CacheHome).
func (p XDGPaths) CacheDir() string {
	return p.CacheHome
}

// Validate checks that all required paths are set and absolute.
func (p XDGPaths) Validate() error {
	if p.ConfigHome == "" {
		return NewValidationError("XDGPaths", "ConfigHome", "", "config home path is required", nil)
	}
	if !filepath.IsAbs(p.ConfigHome) {
		return NewValidationError("XDGPaths", "ConfigHome", p.ConfigHome, "config home path must be absolute", nil)
	}
	if p.DataHome == "" {
		return NewValidationError("XDGPaths", "DataHome", "", "data home path is required", nil)
	}
	if !filepath.IsAbs(p.DataHome) {
		return NewValidationError("XDGPaths", "DataHome", p.DataHome, "data home path must be absolute", nil)
	}
	if p.CacheHome == "" {
		return NewValidationError("XDGPaths", "CacheHome", "", "cache home path is required", nil)
	}
	if !filepath.IsAbs(p.CacheHome) {
		return NewValidationError("XDGPaths", "CacheHome", p.CacheHome, "cache home path must be absolute", nil)
	}
	return nil
}
