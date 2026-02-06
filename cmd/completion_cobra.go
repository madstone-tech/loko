package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion script",
	Long: `Generate the autocompletion script for loko for the specified shell.

To load completions:

Bash:
  $ source <(loko completion bash)
  # To load completions for each session, execute once:
  # Linux:
  $ loko completion bash > /etc/bash_completion.d/loko
  # macOS:
  $ loko completion bash > $(brew --prefix)/etc/bash_completion.d/loko

Zsh:
  $ source <(loko completion zsh)
  # To load completions for each session, execute once:
  $ loko completion zsh > "${fpath[1]}/_loko"

Fish:
  $ loko completion fish | source
  # To load completions for each session, execute once:
  $ loko completion fish > ~/.config/fish/completions/loko.fish

PowerShell:
  PS> loko completion powershell | Out-String | Invoke-Expression
  # To load completions for each session, execute once:
  PS> loko completion powershell > loko.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			return rootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			return rootCmd.GenFishCompletion(os.Stdout, true)
		case "powershell":
			return rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
		}
		return nil
	},
}

func init() {
	// Disable the default completion command since we provide our own.
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.AddCommand(completionCmd)
}
