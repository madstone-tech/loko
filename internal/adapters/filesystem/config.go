package filesystem

import (
	"fmt"
	"os"
	"strings"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// loadConfig loads the loko.toml configuration file.
// If the file doesn't exist, returns default configuration.
func loadConfig(configPath string) (*entities.ProjectConfig, error) {
	config, _, err := loadConfigWithName(configPath)
	return config, err
}

// loadConfigWithName loads the loko.toml configuration file and extracts the project name.
// If the file doesn't exist, returns default configuration.
func loadConfigWithName(configPath string) (*entities.ProjectConfig, string, error) {
	// Check if config file exists
	if _, err := os.Stat(configPath); err != nil {
		// Return default config if file doesn't exist
		return entities.DefaultProjectConfig(), "", nil
	}

	// Read config file
	content, err := os.ReadFile(configPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse TOML (simple parser for now)
	config := entities.DefaultProjectConfig()
	projectName := ""
	if err := parseTomlWithName(string(content), config, &projectName); err != nil {
		return nil, "", fmt.Errorf("failed to parse config: %w", err)
	}

	return config, projectName, nil
}

// saveConfigWithProject saves the configuration with project metadata to loko.toml.
func saveConfigWithProject(configPath string, project *entities.Project) error {
	if project == nil {
		return fmt.Errorf("project cannot be nil")
	}

	if project.Config == nil {
		return fmt.Errorf("project config cannot be nil")
	}

	// Generate TOML content with project metadata
	content := generateTomlWithProject(project)

	// Write to file
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// parseTomlWithName parses a simple TOML configuration and extracts the project name.
// This is a minimal parser that handles the loko.toml format.
func parseTomlWithName(content string, config *entities.ProjectConfig, projectName *string) error {
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Skip section headers for now
		if strings.HasPrefix(line, "[") {
			continue
		}

		// Parse key = value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, "\"'")

		// Extract project name if present
		if key == "name" && projectName != nil {
			*projectName = value
		}

		// Map to config fields
		switch key {
		case "source":
			config.SourceDir = value
		case "output":
			config.OutputDir = value
		case "theme":
			config.D2Theme = value
		case "layout":
			config.D2Layout = value
		case "cache":
			config.D2Cache = value == "true"
		case "parallel":
			config.Parallel = value == "true"
		case "max_workers":
			// Parse integer
			if n, err := parseInt(value); err == nil {
				config.MaxWorkers = n
			}
		case "serve_port":
			if n, err := parseInt(value); err == nil {
				config.ServePort = n
			}
		case "api_port":
			if n, err := parseInt(value); err == nil {
				config.APIPort = n
			}
		case "hot_reload":
			config.HotReload = value == "true"
		}
	}

	return nil
}

// generateTomlWithProject generates TOML content with project metadata.
func generateTomlWithProject(project *entities.Project) string {
	var sb strings.Builder

	sb.WriteString("[project]\n")
	sb.WriteString(fmt.Sprintf("name = %q\n", project.Name))
	if project.Description != "" {
		sb.WriteString(fmt.Sprintf("description = %q\n", project.Description))
	}
	if project.Version != "" {
		sb.WriteString(fmt.Sprintf("version = %q\n", project.Version))
	}
	sb.WriteString("\n")

	sb.WriteString("[paths]\n")
	sb.WriteString(fmt.Sprintf("source = %q\n", project.Config.SourceDir))
	sb.WriteString(fmt.Sprintf("output = %q\n", project.Config.OutputDir))
	sb.WriteString("\n")

	sb.WriteString("[d2]\n")
	sb.WriteString(fmt.Sprintf("theme = %q\n", project.Config.D2Theme))
	sb.WriteString(fmt.Sprintf("layout = %q\n", project.Config.D2Layout))
	sb.WriteString(fmt.Sprintf("cache = %v\n", project.Config.D2Cache))
	sb.WriteString("\n")

	sb.WriteString("[outputs]\n")
	sb.WriteString(fmt.Sprintf("html = %v\n", project.Config.HTMLEnabled))
	sb.WriteString(fmt.Sprintf("markdown = %v\n", project.Config.MarkdownEnabled))
	sb.WriteString(fmt.Sprintf("pdf = %v\n", project.Config.PDFEnabled))
	sb.WriteString("\n")

	sb.WriteString("[build]\n")
	sb.WriteString(fmt.Sprintf("parallel = %v\n", project.Config.Parallel))
	sb.WriteString(fmt.Sprintf("max_workers = %d\n", project.Config.MaxWorkers))
	sb.WriteString("\n")

	sb.WriteString("[server]\n")
	sb.WriteString(fmt.Sprintf("serve_port = %d\n", project.Config.ServePort))
	sb.WriteString(fmt.Sprintf("api_port = %d\n", project.Config.APIPort))
	sb.WriteString(fmt.Sprintf("hot_reload = %v\n", project.Config.HotReload))

	return sb.String()
}

// parseInt parses a string to an integer.
func parseInt(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}
