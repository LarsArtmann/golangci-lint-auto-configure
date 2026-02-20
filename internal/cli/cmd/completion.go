package cmd

import (
	"github.com/spf13/cobra"
)

// NewCompletionCommand creates the completion command for shell autocompletion.
func NewCompletionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion script",
		Long: `Generate shell completion script for golangci-linter-auto-configure.

To load completions:

Bash:
  $ source <(golangci-linter-auto-configure completion bash)
  # To load completions for each session, execute once:
  # Linux:
  $ golangci-linter-auto-configure completion bash > /etc/bash_completion.d/golangci-linter-auto-configure
  # macOS:
  $ golangci-linter-auto-configure completion bash > $(brew --prefix)/etc/bash_completion.d/golangci-linter-auto-configure

Zsh:
  $ source <(golangci-linter-auto-configure completion zsh)
  # To load completions for each session, execute once:
  $ golangci-linter-auto-configure completion zsh > "${fpath[1]}/_golangci-linter-auto-configure"

Fish:
  $ golangci-linter-auto-configure completion fish | source
  # To load completions for each session, execute once:
  $ golangci-linter-auto-configure completion fish > ~/.config/fish/completions/golangci-linter-auto-configure.fish

PowerShell:
  PS> golangci-linter-auto-configure completion powershell | Out-String | Invoke-Expression
  # To load completions for every new session, run:
  PS> golangci-linter-auto-configure completion powershell > golangci-linter-auto-configure.ps1
  # and source this file from your PowerShell profile.
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "bash":
				_ = cmd.Root().GenBashCompletion(cmd.OutOrStdout())
			case "zsh":
				_ = cmd.Root().GenZshCompletion(cmd.OutOrStdout())
			case "fish":
				_ = cmd.Root().GenFishCompletion(cmd.OutOrStdout(), true)
			case "powershell":
				_ = cmd.Root().GenPowerShellCompletionWithDesc(cmd.OutOrStdout())
			}
		},
	}
}
