package config

import (
	"os"
	"path/filepath"

	"github.com/madstone-tech/loko/internal/core/entities"
)

const appName = "loko"

// XDGPathResolver implements usecases.PathResolver using XDG Base Directory Specification.
type XDGPathResolver struct {
	paths entities.XDGPaths
}

// NewXDGPathResolver creates a path resolver with XDG-compliant directory resolution.
func NewXDGPathResolver() *XDGPathResolver {
	home, _ := os.UserHomeDir()

	return &XDGPathResolver{
		paths: entities.XDGPaths{
			ConfigHome: resolveDir(
				os.Getenv("LOKO_CONFIG_HOME"),
				envWithSuffix("XDG_CONFIG_HOME", appName),
				filepath.Join(home, ".config", appName),
			),
			DataHome: resolveDir(
				envWithSuffix("XDG_DATA_HOME", appName),
				filepath.Join(home, ".local", "share", appName),
			),
			CacheHome: resolveDir(
				envWithSuffix("XDG_CACHE_HOME", appName),
				filepath.Join(home, ".cache", appName),
			),
		},
	}
}

func (r *XDGPathResolver) ConfigDir() string  { return r.paths.ConfigHome }
func (r *XDGPathResolver) DataDir() string    { return r.paths.DataHome }
func (r *XDGPathResolver) CacheDir() string   { return r.paths.CacheHome }
func (r *XDGPathResolver) ConfigFile() string { return r.paths.ConfigFile() }
func (r *XDGPathResolver) ThemesDir() string  { return r.paths.ThemesDir() }

// EnsureDir creates the directory if it doesn't exist (lazy creation on first write).
func (r *XDGPathResolver) EnsureDir(path string) error {
	return os.MkdirAll(path, 0o755)
}

// Paths returns the resolved XDG paths as a value object.
func (r *XDGPathResolver) Paths() entities.XDGPaths {
	return r.paths
}

// resolveDir returns the first non-empty path from the candidates.
func resolveDir(candidates ...string) string {
	for _, c := range candidates {
		if c != "" {
			return c
		}
	}
	return ""
}

// envWithSuffix returns the env var value with appName appended, or empty string if not set.
func envWithSuffix(envVar, suffix string) string {
	val := os.Getenv(envVar)
	if val == "" {
		return ""
	}
	return filepath.Join(val, suffix)
}
