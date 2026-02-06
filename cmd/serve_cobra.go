package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var serveCmd = &cobra.Command{
	Use:     "serve",
	Aliases: []string{"s"},
	Short:   "Serve documentation locally",
	Long:    "Start a local HTTP server to preview the built documentation.",
	GroupID: "serving",
	Example: `  loko serve
  loko serve --port 3000
  loko serve --output ./docs --address 0.0.0.0`,
	RunE: runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().StringP("output", "o", "dist", "directory to serve")
	serveCmd.Flags().String("address", "localhost", "server address")
	serveCmd.Flags().String("port", "8080", "server port")

	_ = viper.BindPFlag("server.serve_port", serveCmd.Flags().Lookup("port"))
}

func runServe(cmd *cobra.Command, args []string) error {
	output, _ := cmd.Flags().GetString("output")
	serveCommand := NewServeCommand(output)

	if addr, _ := cmd.Flags().GetString("address"); addr != "localhost" {
		serveCommand.WithAddress(addr)
	}
	if port, _ := cmd.Flags().GetString("port"); port != "8080" {
		serveCommand.WithPort(port)
	}

	return serveCommand.Execute(cmd.Context())
}
