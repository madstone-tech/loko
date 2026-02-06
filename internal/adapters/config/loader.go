// Package config provides configuration loading from loko.toml files.
// It implements the ConfigLoader interface for reading and writing project configuration.
package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	toml "github.com/pelletier/go-toml/v2"
	"github.com/spf13/viper"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// Loader implements the ConfigLoader interface using Viper for config reading
// and pelletier/go-toml/v2 for TOML writing.
type Loader struct {
	paths *XDGPathResolver
}

// NewLoader creates a new config loader.
// If paths is nil, a default XDGPathResolver is used.
func NewLoader(paths *XDGPathResolver) *Loader {
	if paths == nil {
		paths = NewXDGPathResolver()
	}
	return &Loader{paths: paths}
}

// LoadConfig reads configuration from both global and project-local loko.toml files.
// Global config is read first, then project-local config is merged on top.
// Missing files are silently ignored; defaults from entities.DefaultProjectConfig() apply.
func (l *Loader) LoadConfig(ctx context.Context, projectRoot string) (*entities.ProjectConfig, error) {
	v := viper.New()
	v.SetConfigType("toml")

	// Try to read global config first.
	globalFile := l.paths.ConfigFile()
	if globalFile != "" {
		v.SetConfigFile(globalFile)
		if err := v.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				// Only fail on errors other than file-not-found.
				if !os.IsNotExist(err) {
					return nil, fmt.Errorf("failed to load global config: %w", err)
				}
			}
		}
	}

	// Merge project-local config on top of global.
	projectConfigPath := filepath.Join(projectRoot, "loko.toml")
	v.SetConfigFile(projectConfigPath)
	if err := v.MergeInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("failed to load project config: %w", err)
			}
		}
	}

	return viperToConfig(v), nil
}

// LoadGlobalConfig reads only the global config file (~/.config/loko/config.toml).
// Missing file is silently ignored; defaults from entities.DefaultProjectConfig() apply.
func (l *Loader) LoadGlobalConfig(ctx context.Context) (*entities.ProjectConfig, error) {
	v := viper.New()
	v.SetConfigType("toml")

	globalFile := l.paths.ConfigFile()
	if globalFile != "" {
		v.SetConfigFile(globalFile)
		if err := v.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				if !os.IsNotExist(err) {
					return nil, fmt.Errorf("failed to load global config: %w", err)
				}
			}
		}
	}

	return viperToConfig(v), nil
}

// SaveConfig persists configuration to loko.toml in the given project root.
func (l *Loader) SaveConfig(ctx context.Context, projectRoot string, config *entities.ProjectConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	if err := os.MkdirAll(projectRoot, 0o755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	configPath := filepath.Join(projectRoot, "loko.toml")
	return writeConfigToFile(configPath, config)
}

// SaveGlobalConfig persists configuration to the global config file.
func (l *Loader) SaveGlobalConfig(ctx context.Context, config *entities.ProjectConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	if err := os.MkdirAll(l.paths.ConfigDir(), 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	return writeConfigToFile(l.paths.ConfigFile(), config)
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
	// Default to HTML if none enabled.
	if len(formats) == 0 {
		formats = append(formats, "html")
	}
	return formats
}

// viperToConfig maps Viper keys to a ProjectConfig, starting from defaults.
func viperToConfig(v *viper.Viper) *entities.ProjectConfig {
	config := entities.DefaultProjectConfig()

	if v.IsSet("paths.source") {
		config.SourceDir = v.GetString("paths.source")
	}
	if v.IsSet("paths.output") {
		config.OutputDir = v.GetString("paths.output")
	}
	if v.IsSet("d2.theme") {
		config.D2Theme = v.GetString("d2.theme")
	}
	if v.IsSet("d2.layout") {
		config.D2Layout = v.GetString("d2.layout")
	}
	if v.IsSet("d2.cache") {
		config.D2Cache = v.GetBool("d2.cache")
	}
	if v.IsSet("outputs.html") {
		config.HTMLEnabled = v.GetBool("outputs.html")
	}
	if v.IsSet("outputs.markdown") {
		config.MarkdownEnabled = v.GetBool("outputs.markdown")
	}
	if v.IsSet("outputs.pdf") {
		config.PDFEnabled = v.GetBool("outputs.pdf")
	}
	if v.IsSet("build.parallel") {
		config.Parallel = v.GetBool("build.parallel")
	}
	if v.IsSet("build.max_workers") {
		config.MaxWorkers = v.GetInt("build.max_workers")
	}
	if v.IsSet("server.serve_port") {
		config.ServePort = v.GetInt("server.serve_port")
	}
	if v.IsSet("server.api_port") {
		config.APIPort = v.GetInt("server.api_port")
	}
	if v.IsSet("server.hot_reload") {
		config.HotReload = v.GetBool("server.hot_reload")
	}
	if v.IsSet("project.template") {
		config.Template = v.GetString("project.template")
	}

	return config
}

// tomlConfig is the TOML serialization structure for SaveConfig/SaveGlobalConfig.
type tomlConfig struct {
	Paths   tomlPaths   `toml:"paths"`
	D2      tomlD2      `toml:"d2"`
	Outputs tomlOutputs `toml:"outputs"`
	Build   tomlBuild   `toml:"build"`
	Server  tomlServer  `toml:"server"`
}

type tomlPaths struct {
	Source string `toml:"source"`
	Output string `toml:"output"`
}

type tomlD2 struct {
	Theme  string `toml:"theme"`
	Layout string `toml:"layout"`
	Cache  bool   `toml:"cache"`
}

type tomlOutputs struct {
	HTML     bool `toml:"html"`
	Markdown bool `toml:"markdown"`
	PDF      bool `toml:"pdf"`
}

type tomlBuild struct {
	Parallel   bool `toml:"parallel"`
	MaxWorkers int  `toml:"max_workers"`
}

type tomlServer struct {
	ServePort int  `toml:"serve_port"`
	APIPort   int  `toml:"api_port"`
	HotReload bool `toml:"hot_reload"`
}

// writeConfigToFile marshals a ProjectConfig to TOML and writes it to the given path.
func writeConfigToFile(path string, config *entities.ProjectConfig) error {
	tc := tomlConfig{
		Paths: tomlPaths{
			Source: config.SourceDir,
			Output: config.OutputDir,
		},
		D2: tomlD2{
			Theme:  config.D2Theme,
			Layout: config.D2Layout,
			Cache:  config.D2Cache,
		},
		Outputs: tomlOutputs{
			HTML:     config.HTMLEnabled,
			Markdown: config.MarkdownEnabled,
			PDF:      config.PDFEnabled,
		},
		Build: tomlBuild{
			Parallel:   config.Parallel,
			MaxWorkers: config.MaxWorkers,
		},
		Server: tomlServer{
			ServePort: config.ServePort,
			APIPort:   config.APIPort,
			HotReload: config.HotReload,
		},
	}

	data, err := toml.Marshal(tc)
	if err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	header := "# loko project configuration\n# See https://github.com/madstone-tech/loko for documentation\n\n"
	content := append([]byte(header), data...)

	if err := os.WriteFile(path, content, 0o644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
