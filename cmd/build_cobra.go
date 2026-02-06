package cmd

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var buildCmd = &cobra.Command{
	Use:     "build",
	Aliases: []string{"b"},
	Short:   "Build documentation site",
	Long:    "Build the architecture documentation site, rendering D2 diagrams and generating output.",
	GroupID: "building",
	Example: `  loko build
  loko build --clean
  loko build --format html,markdown --d2-theme dark-mauve
  loko build --output ./docs --d2-layout dagre`,
	RunE: runBuild,
}

func init() {
	rootCmd.AddCommand(buildCmd)
	buildCmd.Flags().Bool("clean", false, "rebuild everything (ignore cache)")
	buildCmd.Flags().StringP("output", "o", "dist", "output directory")
	buildCmd.Flags().StringSliceP("format", "f", []string{"html"}, "output formats (html,markdown,pdf)")
	buildCmd.Flags().String("d2-theme", "neutral-default", "D2 diagram theme")
	buildCmd.Flags().String("d2-layout", "elk", "D2 layout engine (dagre, elk, tala)")

	// Bind flags to Viper keys so config/env values apply when flags aren't set.
	_ = viper.BindPFlag("d2.theme", buildCmd.Flags().Lookup("d2-theme"))
	_ = viper.BindPFlag("d2.layout", buildCmd.Flags().Lookup("d2-layout"))
	_ = viper.BindPFlag("output.dir", buildCmd.Flags().Lookup("output"))

	// Flag completion functions.
	_ = buildCmd.RegisterFlagCompletionFunc("d2-theme", completeD2Themes)
	_ = buildCmd.RegisterFlagCompletionFunc("d2-layout", completeD2Layouts)
	_ = buildCmd.RegisterFlagCompletionFunc("format", completeFormats)
}

// completeD2Themes returns available D2 diagram themes.
func completeD2Themes(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return []string{
		"neutral-default\tDefault neutral theme",
		"neutral-grey\tGrey neutral theme",
		"flagship-terrastruct\tTerrastruct flagship theme",
		"cool-classics\tCool classic colors",
		"mixed-berry-blue\tMixed berry blue tones",
		"grape-soda\tGrape soda purple tones",
		"aubergine\tDark aubergine tones",
		"colorblind-clear\tColorblind-friendly palette",
		"vanilla-nitro-cola\tVanilla nitro cola tones",
		"shirley-temple\tShirley temple pastel tones",
		"earth-tones\tEarthy natural tones",
		"everglade-green\tEverglade green tones",
		"buttered-toast\tButtered toast warm tones",
		"dark-mauve\tDark mauve tones",
		"terminal\tTerminal green-on-black",
		"terminal-grayscale\tTerminal grayscale",
		"origami\tOrigami paper tones",
	}, cobra.ShellCompDirectiveNoFileComp
}

// completeD2Layouts returns available D2 layout engines.
func completeD2Layouts(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return []string{
		"dagre\tFast hierarchical layout",
		"elk\tEclipse Layout Kernel",
		"tala\tTerrastruct auto-layout",
	}, cobra.ShellCompDirectiveNoFileComp
}

// completeFormats returns available output formats.
func completeFormats(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return []string{
		"html\tHTML documentation site",
		"markdown\tMarkdown documentation",
		"pdf\tPDF document",
	}, cobra.ShellCompDirectiveNoFileComp
}

func runBuild(cmd *cobra.Command, args []string) error {
	buildCommand := NewBuildCommand(ProjectRoot)

	if clean, _ := cmd.Flags().GetBool("clean"); clean {
		buildCommand.WithClean(true)
	}

	// Use Viper values (respects flag > env > config > default hierarchy).
	output := viper.GetString("output.dir")
	if output != "" && output != "dist" {
		buildCommand.WithOutputDir(output)
	}

	if formats, _ := cmd.Flags().GetStringSlice("format"); len(formats) > 0 {
		for i := range formats {
			formats[i] = strings.TrimSpace(formats[i])
		}
		buildCommand.WithFormats(formats)
	}

	// d2-theme and d2-layout are available via viper.GetString("d2.theme") / viper.GetString("d2.layout")
	// The build command will use these when the config system is fully wired to the D2 renderer.

	return buildCommand.Execute(cmd.Context())
}
