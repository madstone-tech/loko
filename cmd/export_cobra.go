package cmd

import "github.com/spf13/cobra"

var exportCmd = &cobra.Command{
	Use:     "export",
	Short:   "Export documentation in various formats",
	Long:    "Export the architecture documentation as HTML, Markdown, or PDF.",
	GroupID: "building",
}

var exportHTMLCmd = &cobra.Command{
	Use:     "html",
	Short:   "Export as HTML documentation site",
	Example: "  loko export html\n  loko export html --output ./docs",
	RunE: func(cmd *cobra.Command, args []string) error {
		output, _ := cmd.Flags().GetString("output")
		buildCommand := NewBuildCommand(ProjectRoot)
		buildCommand.WithOutputDir(output)
		buildCommand.WithFormats([]string{"html"})
		return buildCommand.Execute(cmd.Context())
	},
}

var exportMarkdownCmd = &cobra.Command{
	Use:     "markdown",
	Short:   "Export as Markdown documentation",
	Example: "  loko export markdown\n  loko export markdown --output ./docs",
	RunE: func(cmd *cobra.Command, args []string) error {
		output, _ := cmd.Flags().GetString("output")
		buildCommand := NewBuildCommand(ProjectRoot)
		buildCommand.WithOutputDir(output)
		buildCommand.WithFormats([]string{"markdown"})
		return buildCommand.Execute(cmd.Context())
	},
}

var exportPDFCmd = &cobra.Command{
	Use:     "pdf",
	Short:   "Export as PDF document",
	Example: "  loko export pdf\n  loko export pdf --output ./docs",
	RunE: func(cmd *cobra.Command, args []string) error {
		output, _ := cmd.Flags().GetString("output")
		buildCommand := NewBuildCommand(ProjectRoot)
		buildCommand.WithOutputDir(output)
		buildCommand.WithFormats([]string{"pdf"})
		return buildCommand.Execute(cmd.Context())
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)

	exportCmd.AddCommand(exportHTMLCmd)
	exportHTMLCmd.Flags().StringP("output", "o", "dist", "output directory")

	exportCmd.AddCommand(exportMarkdownCmd)
	exportMarkdownCmd.Flags().StringP("output", "o", "dist", "output directory")

	exportCmd.AddCommand(exportPDFCmd)
	exportPDFCmd.Flags().StringP("output", "o", "dist", "output directory")
}
