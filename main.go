// Package main is the entry point for the loko CLI.
// loko is a C4 model architecture documentation tool.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/madstone-tech/loko/cmd"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	subcommand := os.Args[1]

	switch subcommand {
	case "--version":
		fmt.Printf("loko %s (commit: %s, built: %s by %s)\n", version, commit, date, builtBy)
		os.Exit(0)

	case "init":
		handleInit()

	case "new":
		handleNew()

	case "build":
		handleBuild()

	case "serve":
		handleServe()

	case "watch":
		handleWatch()

	case "validate":
		handleValidate()

	case "help", "-h", "--help":
		printUsage()
		os.Exit(0)

	default:
		fmt.Printf("Unknown command: %s\n\n", subcommand)
		printUsage()
		os.Exit(1)
	}
}

// handleInit handles the 'loko init' command.
func handleInit() {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	description := fs.String("description", "", "Project description")
	path := fs.String("path", "", "Project path (defaults to project name)")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: loko init <project-name> [options]\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(os.Args[2:]); err != nil {
		os.Exit(1)
	}

	args := fs.Args()
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Error: project name is required\n")
		fs.Usage()
		os.Exit(1)
	}

	projectName := args[0]

	// Create and execute init command
	initCmd := cmd.NewInitCommand(projectName)
	if *description != "" {
		initCmd.WithDescription(*description)
	}
	if *path != "" {
		initCmd.WithPath(*path)
	}

	ctx := context.Background()
	if err := initCmd.Execute(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Project '%s' initialized at %s\n", projectName, projectName)
}

// handleNew handles the 'loko new' command.
func handleNew() {
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "Error: entity type and name are required\n")
		fmt.Fprintf(os.Stderr, "Usage: loko new <type> <name> [options]\n")
		os.Exit(1)
	}

	entityType := strings.ToLower(os.Args[2])
	entityName := os.Args[3]

	// Parse remaining arguments as flags
	fs := flag.NewFlagSet("new", flag.ContinueOnError)
	description := fs.String("description", "", "Entity description")
	technology := fs.String("technology", "", "Technology stack")
	parent := fs.String("parent", "", "Parent entity name (for containers/components)")
	projectRoot := fs.String("project", ".", "Project root directory")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: loko new <type> <name> [options]\n\n")
		fmt.Fprintf(os.Stderr, "Types: system, container, component\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(os.Args[4:]); err != nil {
		os.Exit(1)
	}

	// Validate entity type
	if entityType != "system" && entityType != "container" && entityType != "component" {
		fmt.Fprintf(os.Stderr, "Error: unknown entity type '%s'\n", entityType)
		fs.Usage()
		os.Exit(1)
	}

	// Create and execute new command
	newCmd := cmd.NewNewCommand(entityType, entityName)
	if *description != "" {
		newCmd.WithDescription(*description)
	}
	if *technology != "" {
		newCmd.WithTechnology(*technology)
	}
	if *parent != "" {
		newCmd.WithParent(*parent)
	}
	newCmd.WithProjectRoot(*projectRoot)

	ctx := context.Background()
	if err := newCmd.Execute(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ %s '%s' created\n", capitalize(entityType), entityName)
}

// handleBuild handles the 'loko build' command.
func handleBuild() {
	fs := flag.NewFlagSet("build", flag.ExitOnError)
	projectRoot := fs.String("project", ".", "Project root directory")
	clean := fs.Bool("clean", false, "Rebuild everything (ignore cache)")
	outputDir := fs.String("output", "dist", "Output directory")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: loko build [options]\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(os.Args[2:]); err != nil {
		os.Exit(1)
	}

	buildCmd := cmd.NewBuildCommand(*projectRoot)
	if *clean {
		buildCmd.WithClean(true)
	}
	if *outputDir != "dist" {
		buildCmd.WithOutputDir(*outputDir)
	}

	ctx := context.Background()
	if err := buildCmd.Execute(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// handleServe handles the 'loko serve' command.
func handleServe() {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	outputDir := fs.String("output", "dist", "Output directory to serve")
	address := fs.String("address", "localhost", "Server address")
	port := fs.String("port", "8080", "Server port")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: loko serve [options]\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(os.Args[2:]); err != nil {
		os.Exit(1)
	}

	serveCmd := cmd.NewServeCommand(*outputDir)
	if *address != "localhost" {
		serveCmd.WithAddress(*address)
	}
	if *port != "8080" {
		serveCmd.WithPort(*port)
	}

	ctx := context.Background()
	if err := serveCmd.Execute(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// handleWatch handles the 'loko watch' command.
func handleWatch() {
	fs := flag.NewFlagSet("watch", flag.ExitOnError)
	projectRoot := fs.String("project", ".", "Project root directory")
	outputDir := fs.String("output", "dist", "Output directory")
	debounce := fs.Int("debounce", 500, "Debounce delay in milliseconds")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: loko watch [options]\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(os.Args[2:]); err != nil {
		os.Exit(1)
	}

	watchCmd := cmd.NewWatchCommand(*projectRoot)
	if *outputDir != "dist" {
		watchCmd.WithOutputDir(*outputDir)
	}
	if *debounce != 500 {
		watchCmd.WithDebounce(*debounce)
	}

	ctx := context.Background()
	if err := watchCmd.Execute(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// handleValidate handles the 'loko validate' command.
func handleValidate() {
	fs := flag.NewFlagSet("validate", flag.ExitOnError)
	projectRoot := fs.String("project", ".", "Project root directory")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: loko validate [options]\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(os.Args[2:]); err != nil {
		os.Exit(1)
	}

	validateCmd := cmd.NewValidateCommand(*projectRoot)

	ctx := context.Background()
	if err := validateCmd.Execute(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// capitalize capitalizes the first letter of a string.
func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// printUsage prints the CLI usage information.
func printUsage() {
	fmt.Println("🪇 loko - Guardian of Architectural Wisdom")
	fmt.Println()
	fmt.Println("C4 model architecture documentation with LLM integration")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println()
	fmt.Println("  Scaffolding:")
	fmt.Println("    loko init <project-name>              Initialize a new project")
	fmt.Println("    loko new system <name>                Create a new system")
	fmt.Println("    loko new container <name>             Create a new container")
	fmt.Println("    loko new component <name>             Create a new component")
	fmt.Println()
	fmt.Println("  Building & Serving:")
	fmt.Println("    loko build                            Build documentation (render diagrams, generate HTML)")
	fmt.Println("    loko serve                            Serve documentation locally (http://localhost:8080)")
	fmt.Println("    loko watch                            Watch for changes and rebuild automatically")
	fmt.Println("    loko validate                         Validate project structure")
	fmt.Println()
	fmt.Println("  Other:")
	fmt.Println("    loko --version                        Show version")
	fmt.Println("    loko help                             Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  loko init myproject")
	fmt.Println("  cd myproject")
	fmt.Println("  loko new system PaymentService")
	fmt.Println("  loko new container -parent PaymentService API")
	fmt.Println("  loko build                              # Build documentation once")
	fmt.Println("  loko watch                              # Watch and rebuild on changes")
	fmt.Println("  loko serve                              # Serve in another terminal")
	fmt.Println()
	fmt.Println("See https://github.com/madstone-tech/loko for more information")
}
