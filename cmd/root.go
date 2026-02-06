// Package cmd implements the loko CLI commands using Cobra.
package cmd

import (
	"fmt"
	"strings"

	"github.com/madstone-tech/loko/internal/adapters/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Build-time version information, set via SetVersionInfo from main.go.
var (
	appVersion = "dev"
	appCommit  = "none"
	appDate    = "unknown"
	appBuiltBy = "unknown"
)

// Persistent flag values accessible to all subcommands.
var (
	cfgFile     string
	ProjectRoot string
	Verbose     bool
)

// rootCmd is the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "loko",
	Short: "Guardian of Architectural Wisdom",
	Long: `loko is a C4 model architecture documentation tool with LLM integration.

It helps teams create, maintain, and visualize software architecture documentation
using the C4 model (Context, Containers, Components, Code) with automatic
diagram generation via D2 and LLM-powered querying via MCP.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return initConfig(cmd.Root())
	},
	SilenceUsage: true,
}

func init() {
	// Persistent flags available to all subcommands.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "path to config file or directory (env: LOKO_CONFIG_HOME)")
	rootCmd.PersistentFlags().StringVarP(&ProjectRoot, "project", "p", ".", "project root directory")
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "enable verbose output (env: LOKO_VERBOSE)")

	// Command groups for organized help output.
	rootCmd.AddGroup(
		&cobra.Group{ID: "scaffolding", Title: "Scaffolding"},
		&cobra.Group{ID: "building", Title: "Building"},
		&cobra.Group{ID: "serving", Title: "Serving"},
	)
}

// Execute runs the root command. This is the main entry point called from main.go.
func Execute() error {
	return rootCmd.Execute()
}

// SetVersionInfo sets build-time version information from ldflags.
// Call this from main.go before Execute().
func SetVersionInfo(version, commit, date, builtBy string) {
	appVersion = version
	appCommit = commit
	appDate = date
	appBuiltBy = builtBy

	rootCmd.Version = version
	rootCmd.SetVersionTemplate(
		fmt.Sprintf("loko %s (commit: %s, built: %s by %s)\n", version, commit, date, builtBy),
	)
}

// initConfig sets up Viper configuration with the full hierarchy:
// CLI flags > LOKO_* env vars > project loko.toml > global XDG config.toml > defaults
func initConfig(root *cobra.Command) error {
	viper.SetConfigType("toml")

	// 1. Set built-in defaults.
	viper.SetDefault("d2.theme", "neutral-default")
	viper.SetDefault("d2.layout", "elk")
	viper.SetDefault("d2.cache", true)
	viper.SetDefault("paths.source", "./src")
	viper.SetDefault("paths.output", "./dist")
	viper.SetDefault("outputs.html", true)
	viper.SetDefault("outputs.markdown", false)
	viper.SetDefault("outputs.pdf", false)
	viper.SetDefault("build.parallel", true)
	viper.SetDefault("build.max_workers", 4)
	viper.SetDefault("server.serve_port", 8080)
	viper.SetDefault("server.api_port", 8081)
	viper.SetDefault("server.hot_reload", true)

	// 2. Read global config (lowest priority file).
	if cfgFile != "" {
		// --config flag overrides all path resolution.
		viper.SetConfigFile(cfgFile)
		if err := viper.ReadInConfig(); err != nil {
			return fmt.Errorf("failed to read config file %s: %w", cfgFile, err)
		}
	} else {
		// Try XDG global config path.
		paths := config.NewXDGPathResolver()
		viper.SetConfigFile(paths.ConfigFile())
		_ = viper.ReadInConfig() // Silent fail if not found.
	}

	// 3. Merge project config (overrides global).
	viper.SetConfigFile("loko.toml")
	_ = viper.MergeInConfig() // Silent fail if not found.

	// 4. Environment variables override config files.
	viper.SetEnvPrefix("LOKO")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// 5. Apply custom command aliases from [aliases] config section.
	applyCustomAliases(root)

	return nil
}

// applyCustomAliases reads the [aliases] section from config and appends
// custom aliases to matching top-level commands. Config values can be a
// single string or an array of strings. Invalid entries are silently skipped.
func applyCustomAliases(root *cobra.Command) {
	aliasMap := viper.GetStringMap("aliases")
	if len(aliasMap) == 0 {
		return
	}

	commands := root.Commands()
	cmdByName := make(map[string]*cobra.Command, len(commands))
	for _, cmd := range commands {
		cmdByName[cmd.Name()] = cmd
	}

	for name, value := range aliasMap {
		cmd, ok := cmdByName[name]
		if !ok {
			continue
		}

		var aliases []string
		switch v := value.(type) {
		case string:
			aliases = []string{v}
		case []any:
			for _, item := range v {
				if s, ok := item.(string); ok {
					aliases = append(aliases, s)
				}
			}
		default:
			continue
		}

		cmd.Aliases = append(cmd.Aliases, aliases...)
	}
}
