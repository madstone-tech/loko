package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities"
)

func TestLoader_LoadConfig_Defaults(t *testing.T) {
	loader := NewLoader(nil)
	ctx := context.Background()

	// Use a temp directory with no config file
	tmpDir := t.TempDir()

	config, err := loader.LoadConfig(ctx, tmpDir)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify defaults
	defaults := entities.DefaultProjectConfig()
	if config.SourceDir != defaults.SourceDir {
		t.Errorf("SourceDir = %q, want %q", config.SourceDir, defaults.SourceDir)
	}
	if config.OutputDir != defaults.OutputDir {
		t.Errorf("OutputDir = %q, want %q", config.OutputDir, defaults.OutputDir)
	}
	if config.HTMLEnabled != defaults.HTMLEnabled {
		t.Errorf("HTMLEnabled = %v, want %v", config.HTMLEnabled, defaults.HTMLEnabled)
	}
	if config.MarkdownEnabled != defaults.MarkdownEnabled {
		t.Errorf("MarkdownEnabled = %v, want %v", config.MarkdownEnabled, defaults.MarkdownEnabled)
	}
	if config.PDFEnabled != defaults.PDFEnabled {
		t.Errorf("PDFEnabled = %v, want %v", config.PDFEnabled, defaults.PDFEnabled)
	}
}

func TestLoader_LoadConfig_FromFile(t *testing.T) {
	loader := NewLoader(nil)
	ctx := context.Background()

	// Create temp directory with config file
	tmpDir := t.TempDir()
	configContent := `
[paths]
source = "./architecture"
output = "./docs"

[d2]
theme = "dark"
layout = "dagre"
cache = false

[outputs]
html = true
markdown = true
pdf = false

[build]
parallel = false
max_workers = 2

[server]
serve_port = 3000
api_port = 3001
hot_reload = false
`
	configPath := filepath.Join(tmpDir, "loko.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	config, err := loader.LoadConfig(ctx, tmpDir)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify loaded values
	if config.SourceDir != "./architecture" {
		t.Errorf("SourceDir = %q, want %q", config.SourceDir, "./architecture")
	}
	if config.OutputDir != "./docs" {
		t.Errorf("OutputDir = %q, want %q", config.OutputDir, "./docs")
	}
	if config.D2Theme != "dark" {
		t.Errorf("D2Theme = %q, want %q", config.D2Theme, "dark")
	}
	if config.D2Layout != "dagre" {
		t.Errorf("D2Layout = %q, want %q", config.D2Layout, "dagre")
	}
	if config.D2Cache != false {
		t.Errorf("D2Cache = %v, want false", config.D2Cache)
	}
	if config.HTMLEnabled != true {
		t.Errorf("HTMLEnabled = %v, want true", config.HTMLEnabled)
	}
	if config.MarkdownEnabled != true {
		t.Errorf("MarkdownEnabled = %v, want true", config.MarkdownEnabled)
	}
	if config.PDFEnabled != false {
		t.Errorf("PDFEnabled = %v, want false", config.PDFEnabled)
	}
	if config.Parallel != false {
		t.Errorf("Parallel = %v, want false", config.Parallel)
	}
	if config.MaxWorkers != 2 {
		t.Errorf("MaxWorkers = %d, want 2", config.MaxWorkers)
	}
	if config.ServePort != 3000 {
		t.Errorf("ServePort = %d, want 3000", config.ServePort)
	}
	if config.APIPort != 3001 {
		t.Errorf("APIPort = %d, want 3001", config.APIPort)
	}
	if config.HotReload != false {
		t.Errorf("HotReload = %v, want false", config.HotReload)
	}
}

func TestLoader_SaveConfig(t *testing.T) {
	loader := NewLoader(nil)
	ctx := context.Background()
	tmpDir := t.TempDir()

	config := entities.DefaultProjectConfig()
	config.SourceDir = "./custom-src"
	config.OutputDir = "./custom-dist"
	config.MarkdownEnabled = true
	config.PDFEnabled = true

	err := loader.SaveConfig(ctx, tmpDir, config)
	if err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Verify file was created
	configPath := filepath.Join(tmpDir, "loko.toml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// Load it back and verify
	loadedConfig, err := loader.LoadConfig(ctx, tmpDir)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if loadedConfig.SourceDir != "./custom-src" {
		t.Errorf("SourceDir = %q, want %q", loadedConfig.SourceDir, "./custom-src")
	}
	if loadedConfig.OutputDir != "./custom-dist" {
		t.Errorf("OutputDir = %q, want %q", loadedConfig.OutputDir, "./custom-dist")
	}
	if loadedConfig.MarkdownEnabled != true {
		t.Errorf("MarkdownEnabled = %v, want true", loadedConfig.MarkdownEnabled)
	}
	if loadedConfig.PDFEnabled != true {
		t.Errorf("PDFEnabled = %v, want true", loadedConfig.PDFEnabled)
	}
}

func TestLoader_SaveConfig_NilConfig(t *testing.T) {
	loader := NewLoader(nil)
	ctx := context.Background()
	tmpDir := t.TempDir()

	err := loader.SaveConfig(ctx, tmpDir, nil)
	if err == nil {
		t.Error("Expected error for nil config")
	}
}

func TestGetOutputFormats(t *testing.T) {
	tests := []struct {
		name     string
		config   *entities.ProjectConfig
		expected []string
	}{
		{
			name: "defaults (HTML only)",
			config: &entities.ProjectConfig{
				HTMLEnabled:     true,
				MarkdownEnabled: false,
				PDFEnabled:      false,
			},
			expected: []string{"html"},
		},
		{
			name: "all formats enabled",
			config: &entities.ProjectConfig{
				HTMLEnabled:     true,
				MarkdownEnabled: true,
				PDFEnabled:      true,
			},
			expected: []string{"html", "markdown", "pdf"},
		},
		{
			name: "only markdown",
			config: &entities.ProjectConfig{
				HTMLEnabled:     false,
				MarkdownEnabled: true,
				PDFEnabled:      false,
			},
			expected: []string{"markdown"},
		},
		{
			name: "none enabled (defaults to HTML)",
			config: &entities.ProjectConfig{
				HTMLEnabled:     false,
				MarkdownEnabled: false,
				PDFEnabled:      false,
			},
			expected: []string{"html"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formats := GetOutputFormats(tt.config)

			if len(formats) != len(tt.expected) {
				t.Errorf("got %d formats, want %d", len(formats), len(tt.expected))
				return
			}

			for i, f := range formats {
				if f != tt.expected[i] {
					t.Errorf("format[%d] = %q, want %q", i, f, tt.expected[i])
				}
			}
		})
	}
}
