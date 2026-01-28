// Package config provides configuration loading from loko.toml files.
// It implements the ConfigLoader interface for reading and writing project configuration.
package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/madstone-tech/loko/internal/core/entities"
)

// Loader implements the ConfigLoader interface for TOML configuration files.
type Loader struct {
	globalConfigPath string // Path to global config (~/.loko/config.toml)
}

// NewLoader creates a new config loader.
func NewLoader() *Loader {
	homeDir, _ := os.UserHomeDir()
	globalPath := ""
	if homeDir != "" {
		globalPath = filepath.Join(homeDir, ".loko", "config.toml")
	}
	return &Loader{
		globalConfigPath: globalPath,
	}
}

// tomlConfig represents the structure of loko.toml file.
type tomlConfig struct {
	Project  projectSection  `toml:"project"`
	Paths    pathsSection    `toml:"paths"`
	D2       d2Section       `toml:"d2"`
	Outputs  outputsSection  `toml:"outputs"`
	Build    buildSection    `toml:"build"`
	Server   serverSection   `toml:"server"`
}

type projectSection struct {
	Name        string `toml:"name"`
	Description string `toml:"description"`
	Version     string `toml:"version"`
}

type pathsSection struct {
	Source string `toml:"source"`
	Output string `toml:"output"`
}

type d2Section struct {
	Theme  string `toml:"theme"`
	Layout string `toml:"layout"`
	Cache  *bool  `toml:"cache"`
}

type outputsSection struct {
	HTML     *bool `toml:"html"`
	Markdown *bool `toml:"markdown"`
	PDF      *bool `toml:"pdf"`
}

type buildSection struct {
	Parallel   *bool `toml:"parallel"`
	MaxWorkers *int  `toml:"max_workers"`
}

type serverSection struct {
	ServePort *int  `toml:"serve_port"`
	APIPort   *int  `toml:"api_port"`
	HotReload *bool `toml:"hot_reload"`
}

// LoadConfig reads loko.toml and applies defaults.
// It reads both global (~/.loko/config.toml) and project-local (./loko.toml) configs,
// with project-local overriding global settings.
func (l *Loader) LoadConfig(ctx context.Context, projectRoot string) (*entities.ProjectConfig, error) {
	// Start with defaults
	config := entities.DefaultProjectConfig()

	// Try to load global config first
	if l.globalConfigPath != "" {
		if _, err := os.Stat(l.globalConfigPath); err == nil {
			if err := l.loadFromFile(l.globalConfigPath, config); err != nil {
				return nil, fmt.Errorf("failed to load global config: %w", err)
			}
		}
	}

	// Load project-local config (overrides global)
	projectConfigPath := filepath.Join(projectRoot, "loko.toml")
	if _, err := os.Stat(projectConfigPath); err == nil {
		if err := l.loadFromFile(projectConfigPath, config); err != nil {
			return nil, fmt.Errorf("failed to load project config: %w", err)
		}
	}

	return config, nil
}

// loadFromFile loads configuration from a TOML file into the config.
func (l *Loader) loadFromFile(path string, config *entities.ProjectConfig) error {
	var tc tomlConfig
	if _, err := toml.DecodeFile(path, &tc); err != nil {
		return fmt.Errorf("failed to parse TOML: %w", err)
	}

	// Apply paths section
	if tc.Paths.Source != "" {
		config.SourceDir = tc.Paths.Source
	}
	if tc.Paths.Output != "" {
		config.OutputDir = tc.Paths.Output
	}

	// Apply D2 section
	if tc.D2.Theme != "" {
		config.D2Theme = tc.D2.Theme
	}
	if tc.D2.Layout != "" {
		config.D2Layout = tc.D2.Layout
	}
	if tc.D2.Cache != nil {
		config.D2Cache = *tc.D2.Cache
	}

	// Apply outputs section
	if tc.Outputs.HTML != nil {
		config.HTMLEnabled = *tc.Outputs.HTML
	}
	if tc.Outputs.Markdown != nil {
		config.MarkdownEnabled = *tc.Outputs.Markdown
	}
	if tc.Outputs.PDF != nil {
		config.PDFEnabled = *tc.Outputs.PDF
	}

	// Apply build section
	if tc.Build.Parallel != nil {
		config.Parallel = *tc.Build.Parallel
	}
	if tc.Build.MaxWorkers != nil {
		config.MaxWorkers = *tc.Build.MaxWorkers
	}

	// Apply server section
	if tc.Server.ServePort != nil {
		config.ServePort = *tc.Server.ServePort
	}
	if tc.Server.APIPort != nil {
		config.APIPort = *tc.Server.APIPort
	}
	if tc.Server.HotReload != nil {
		config.HotReload = *tc.Server.HotReload
	}

	return nil
}

// SaveConfig persists configuration to loko.toml.
func (l *Loader) SaveConfig(ctx context.Context, projectRoot string, config *entities.ProjectConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// Build TOML structure
	tc := tomlConfig{
		Paths: pathsSection{
			Source: config.SourceDir,
			Output: config.OutputDir,
		},
		D2: d2Section{
			Theme:  config.D2Theme,
			Layout: config.D2Layout,
			Cache:  &config.D2Cache,
		},
		Outputs: outputsSection{
			HTML:     &config.HTMLEnabled,
			Markdown: &config.MarkdownEnabled,
			PDF:      &config.PDFEnabled,
		},
		Build: buildSection{
			Parallel:   &config.Parallel,
			MaxWorkers: &config.MaxWorkers,
		},
		Server: serverSection{
			ServePort: &config.ServePort,
			APIPort:   &config.APIPort,
			HotReload: &config.HotReload,
		},
	}

	// Ensure directory exists
	if err := os.MkdirAll(projectRoot, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// Write to file
	configPath := filepath.Join(projectRoot, "loko.toml")
	f, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer f.Close()

	// Write header comment
	f.WriteString("# loko project configuration\n")
	f.WriteString("# See https://github.com/madstone-tech/loko for documentation\n\n")

	encoder := toml.NewEncoder(f)
	if err := encoder.Encode(tc); err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	return nil
}

// GetOutputFormats returns the enabled output formats from config.
func GetOutputFormats(config *entities.ProjectConfig) []string {
	var formats []string
	if config.HTMLEnabled {
		formats = append(formats, "html")
	}
	if config.MarkdownEnabled {
		formats = append(formats, "markdown")
	}
	if config.PDFEnabled {
		formats = append(formats, "pdf")
	}
	// Default to HTML if none enabled
	if len(formats) == 0 {
		formats = append(formats, "html")
	}
	return formats
}
