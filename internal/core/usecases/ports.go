package usecases

import (
	"context"
	"time"

	"github.com/madstone-tech/loko/internal/core/entities"
)

// ProjectRepository defines the interface for persisting and loading projects.
//
// Implementations MUST handle the file system as the canonical storage,
// supporting both TOML configuration and YAML frontmatter in markdown files.
type ProjectRepository interface {
	// LoadProject retrieves a project by its root directory path.
	// Returns ErrProjectNotFound if the project doesn't exist.
	LoadProject(ctx context.Context, projectRoot string) (*entities.Project, error)

	// SaveProject persists a project to disk.
	// Creates directories and files as needed; returns error if write fails.
	SaveProject(ctx context.Context, project *entities.Project) error

	// ListSystems returns all systems in a project.
	ListSystems(ctx context.Context, projectRoot string) ([]*entities.System, error)

	// LoadSystem retrieves a system by name within a project.
	LoadSystem(ctx context.Context, projectRoot, systemName string) (*entities.System, error)

	// SaveSystem persists a system to disk.
	SaveSystem(ctx context.Context, projectRoot string, system *entities.System) error

	// LoadContainer retrieves a container by name within a system.
	LoadContainer(ctx context.Context, projectRoot, systemName, containerName string) (*entities.Container, error)

	// SaveContainer persists a container to disk.
	SaveContainer(ctx context.Context, projectRoot, systemName string, container *entities.Container) error

	// LoadComponent retrieves a component by name within a container.
	LoadComponent(ctx context.Context, projectRoot, systemName, containerName, componentName string) (*entities.Component, error)

	// SaveComponent persists a component to disk.
	SaveComponent(ctx context.Context, projectRoot, systemName, containerName string, component *entities.Component) error
}

// TemplateEngine defines the interface for rendering templates using variable substitution.
//
// Implementations MUST support template discovery from both global (~/.loko/templates/)
// and project-local (.loko/templates/) directories.
type TemplateEngine interface {
	// RenderTemplate loads a template by name and applies variable substitution.
	// Returns the rendered content or error if template not found.
	RenderTemplate(ctx context.Context, templateName string, variables map[string]string) (string, error)

	// ListTemplates returns available template names from discovery paths.
	ListTemplates(ctx context.Context) ([]string, error)

	// AddSearchPath adds a directory to the template search path.
	AddSearchPath(path string)
}

// DiagramRenderer defines the interface for rendering D2 source code to SVG and other formats.
//
// Implementations MUST shell out to the d2 CLI binary and handle missing dependencies gracefully.
type DiagramRenderer interface {
	// RenderDiagram compiles D2 source code to SVG.
	// Returns SVG content or error if d2 binary missing or compilation fails.
	RenderDiagram(ctx context.Context, d2Source string) (svgContent string, err error)

	// RenderDiagramWithTimeout compiles with a specified timeout in seconds.
	RenderDiagramWithTimeout(ctx context.Context, d2Source string, timeoutSec int) (svgContent string, err error)

	// IsAvailable checks if the d2 binary is installed and accessible.
	IsAvailable() bool
}

// SiteBuilder defines the interface for generating static HTML documentation.
//
// Implementations MUST produce a static website with sidebar navigation, breadcrumbs,
// and search functionality suitable for serving via HTTP.
type SiteBuilder interface {
	// BuildSite generates HTML documentation from a project.
	// Creates an output directory with index.html, system pages, container pages, diagrams, and static assets.
	BuildSite(ctx context.Context, project *entities.Project, systems []*entities.System, outputDir string) error

	// BuildSystemPage generates a single system HTML page with embedded diagrams.
	BuildSystemPage(ctx context.Context, system *entities.System, containers []*entities.Container, outputDir string) error
}

// FileWatcher defines the interface for monitoring file system changes.
//
// Implementations MUST use efficient file system APIs (e.g., fsnotify on Linux/macOS)
// and batch changes to prevent excessive rebuilds.
type FileWatcher interface {
	// Watch starts monitoring a directory for changes.
	// Sends change events to the provided channel; returns error if setup fails.
	Watch(ctx context.Context, rootPath string) (<-chan FileChangeEvent, error)

	// Stop halts file watching and closes all channels.
	Stop() error
}

// FileChangeEvent describes a change detected by the file watcher.
type FileChangeEvent struct {
	// Path relative to the watched root
	Path string
	// Op is one of: create, write, remove, rename, chmod
	Op string
}

// Logger defines the interface for structured logging.
//
// Implementations MUST emit JSON logs to stdout in production mode.
// The logger is used throughout the application for tracing and debugging.
type Logger interface {
	// Debug logs a debug-level message.
	Debug(msg string, keysAndValues ...any)

	// Info logs an info-level message.
	Info(msg string, keysAndValues ...any)

	// Warn logs a warning-level message.
	Warn(msg string, keysAndValues ...any)

	// Error logs an error-level message.
	Error(msg string, err error, keysAndValues ...any)

	// WithContext returns a logger that includes the given context (for request/operation tracking).
	WithContext(ctx context.Context) Logger

	// WithFields returns a logger with additional structured fields.
	WithFields(keysAndValues ...any) Logger
}

// ProgressReporter defines the interface for communicating progress to the user.
//
// Implementations MAY use terminal formatting (via lipgloss) for CLI output.
// Progress events include task completion percentage, current step, and status messages.
type ProgressReporter interface {
	// ReportProgress sends a progress update.
	ReportProgress(step string, current int, total int, message string)

	// ReportError sends an error status (typically with red/bold formatting).
	ReportError(err error)

	// ReportSuccess sends a success status (typically with green formatting).
	ReportSuccess(message string)

	// ReportInfo sends an informational message.
	ReportInfo(message string)
}

// OutputEncoder defines the interface for serializing data to various formats.
//
// Implementations MUST support JSON and TOON (token-optimized) formats for
// efficient representation of architecture data.
type OutputEncoder interface {
	// EncodeJSON serializes a value to JSON bytes.
	EncodeJSON(value any) ([]byte, error)

	// EncodeTOON serializes a value to TOON format (token-efficient).
	EncodeTOON(value any) ([]byte, error)

	// DecodeJSON deserializes JSON bytes to a value.
	DecodeJSON(data []byte, value any) error

	// DecodeTOON deserializes TOON format to a value.
	DecodeTOON(data []byte, value any) error
}

// PDFRenderer defines the interface for rendering PDF documents.
//
// Implementations shell out to veve-cli for PDF generation.
// PDF rendering is optional; implementations MUST return ErrPDFNotAvailable if
// the veve-cli binary is not installed.
type PDFRenderer interface {
	// RenderPDF converts HTML to PDF.
	// Returns error if veve-cli not available or rendering fails.
	RenderPDF(ctx context.Context, htmlPath string, outputPath string) error

	// IsAvailable checks if the veve-cli binary is installed.
	IsAvailable() bool
}

// ConfigLoader defines the interface for loading and parsing configuration files.
//
// Implementations MUST support loko.toml (TOML format) with hierarchical config
// (project-level overrides global defaults).
type ConfigLoader interface {
	// LoadConfig reads loko.toml and applies defaults.
	// Supports both ~/.loko/config.toml (global) and ./loko.toml (project-local).
	LoadConfig(ctx context.Context, projectRoot string) (*entities.ProjectConfig, error)

	// SaveConfig persists configuration to loko.toml.
	SaveConfig(ctx context.Context, projectRoot string, config *entities.ProjectConfig) error

	// LoadGlobalConfig reads the global config file (~/.config/loko/config.toml).
	LoadGlobalConfig(ctx context.Context) (*entities.ProjectConfig, error)

	// SaveGlobalConfig persists the global config file.
	SaveGlobalConfig(ctx context.Context, config *entities.ProjectConfig) error
}

// Validator defines the interface for validation logic.
//
// Implementations check structural integrity, reference validity, and convention compliance.
type Validator interface {
	// ValidateProject checks a project for errors (orphaned references, missing files, hierarchy violations).
	ValidateProject(ctx context.Context, project *entities.Project) ([]ValidationError, error)

	// ValidateSystem checks a system for internal consistency.
	ValidateSystem(ctx context.Context, system *entities.System) ([]ValidationError, error)

	// ValidateContainer checks a container for internal consistency.
	ValidateContainer(ctx context.Context, container *entities.Container) ([]ValidationError, error)
}

// ValidationError represents a single validation issue.
type ValidationError struct {
	// Code is the error code (e.g., "orphaned_ref", "missing_file", "invalid_hierarchy")
	Code string
	// Message is the human-readable error description
	Message string
	// Path is the file or entity path where the error occurred
	Path string
	// Line is the line number (if applicable)
	Line int
}

// CompressedFormatter defines the interface for generating token-efficient representations.
//
// Implementations support multiple detail levels (summary, structure, full)
// for querying architecture with minimal token overhead.
type CompressedFormatter interface {
	// FormatSummary returns a brief overview (~200 tokens).
	FormatSummary(project *entities.Project) (string, error)

	// FormatStructure returns system/container hierarchy (~500 tokens).
	FormatStructure(project *entities.Project, systems []*entities.System) (string, error)

	// FormatFull returns complete architecture details (no compression).
	FormatFull(project *entities.Project, systems []*entities.System) (string, error)

	// EstimateTokenCount returns an estimate of token consumption for a format.
	EstimateTokenCount(format string, numSystems int) int
}

// HTTPServer defines the interface for the HTTP API server.
//
// Implementations handle routing, middleware, and request/response formatting
// for the RESTful API.
type HTTPServer interface {
	// Start begins listening on the specified address (e.g., "localhost:8080").
	Start(ctx context.Context, address string) error

	// Stop gracefully shuts down the server.
	Stop(ctx context.Context) error

	// IsRunning returns true if the server is currently listening.
	IsRunning() bool
}

// MCPServer defines the interface for the Model Context Protocol server.
//
// Implementations handle stdio transport, JSON-RPC protocol, and tool invocation.
type MCPServer interface {
	// Start begins listening on stdin/stdout for MCP messages.
	Start(ctx context.Context) error

	// Stop gracefully shuts down the server.
	Stop(ctx context.Context) error

	// RegisterTool registers an MCP tool handler.
	RegisterTool(name string, schema any, handler func(ctx context.Context, args map[string]any) (any, error)) error
}

// MarkdownBuilder defines the interface for generating Markdown documentation.
//
// Implementations MUST produce a single README.md file with complete architecture
// documentation suitable for version control and viewing in text editors.
type MarkdownBuilder interface {
	// BuildMarkdown generates a single README.md with complete architecture.
	BuildMarkdown(ctx context.Context, project *entities.Project, systems []*entities.System) (content string, err error)

	// BuildSystemMarkdown generates Markdown for a single system.
	BuildSystemMarkdown(ctx context.Context, system *entities.System, containers []*entities.Container) (content string, err error)
}

// MarkdownRenderer defines the interface for converting Markdown to HTML.
//
// Implementations MUST support parsing YAML frontmatter, converting markdown syntax
// to HTML, and embedding styled output suitable for standalone viewing.
type MarkdownRenderer interface {
	// RenderMarkdownToHTML converts markdown string to styled HTML.
	// Handles YAML frontmatter stripping, inline formatting (bold, italic, code, links),
	// lists, tables, headings, and code blocks.
	RenderMarkdownToHTML(markdown string) string
}

// PathResolver resolves XDG-compliant paths for application data.
//
// Implementations MUST support XDG Base Directory Specification with env var
// overrides (LOKO_CONFIG_HOME, XDG_CONFIG_HOME, XDG_DATA_HOME, XDG_CACHE_HOME).
type PathResolver interface {
	// ConfigDir returns the configuration directory path.
	// Resolution: LOKO_CONFIG_HOME → XDG_CONFIG_HOME/loko/ → ~/.config/loko/
	ConfigDir() string

	// DataDir returns the data directory path.
	// Resolution: XDG_DATA_HOME/loko/ → ~/.local/share/loko/
	DataDir() string

	// CacheDir returns the cache directory path.
	// Resolution: XDG_CACHE_HOME/loko/ → ~/.cache/loko/
	CacheDir() string

	// ConfigFile returns the path to the global config file.
	// Returns ConfigDir()/config.toml
	ConfigFile() string

	// ThemesDir returns the path to the themes directory.
	// Returns DataDir()/themes/
	ThemesDir() string
}

// ThemeLoader loads and lists available themes.
//
// Implementations read TOML theme files from the themes directory.
type ThemeLoader interface {
	// LoadTheme loads a theme by name from the themes directory.
	// Returns error if theme file not found or invalid.
	LoadTheme(ctx context.Context, name string) (*entities.Theme, error)

	// ListThemes returns the names of all available themes.
	ListThemes(ctx context.Context) ([]string, error)
}

// DiagramGenerator defines the interface for generating D2 diagram source code
// from domain entities. This is a domain service that knows C4 conventions for
// how to represent systems, containers, and components as D2 diagrams.
type DiagramGenerator interface {
	// GenerateSystemContextDiagram generates D2 source showing a system in its context,
	// including external systems, key users, and their relationships.
	GenerateSystemContextDiagram(system *entities.System) (string, error)

	// GenerateContainerDiagram generates D2 source showing containers within a system,
	// with technology labels and inter-container relationships.
	GenerateContainerDiagram(system *entities.System) (string, error)

	// GenerateComponentDiagram generates D2 source showing components within a container,
	// with relationship arrows and technology labels.
	GenerateComponentDiagram(container *entities.Container) (string, error)
}

// UserPrompter defines the interface for interactive user input in CLI mode.
//
// This is a CLI-only concern; MCP and API handlers do not use this interface.
// Implementations MAY return an error if stdin is not a terminal (non-interactive mode).
type UserPrompter interface {
	// PromptString displays a prompt and returns the user's input.
	// If the user provides empty input, returns defaultValue.
	PromptString(prompt string, defaultValue string) (string, error)

	// PromptStringMulti displays a prompt and collects multiple lines until an empty line.
	// Returns a slice of non-empty strings.
	PromptStringMulti(prompt string) ([]string, error)
}

// ReportFormatter defines the interface for formatting reports for human display.
//
// Implementations MAY use terminal formatting (via lipgloss) for CLI output
// and plain text for non-TTY environments.
type ReportFormatter interface {
	// PrintValidationReport formats and displays validation errors grouped by severity.
	PrintValidationReport(errors []ValidationError)

	// PrintBuildReport formats and displays build statistics.
	PrintBuildReport(stats BuildStats)
}

// BuildStats holds statistics from a documentation build for reporting.
type BuildStats struct {
	// FilesGenerated is the count of output files created.
	FilesGenerated int
	// DiagramCount is the number of diagrams rendered.
	DiagramCount int
	// Duration is the total build time.
	Duration time.Duration
	// Format is the output format used (html, markdown, pdf).
	Format string
}
