package cmd

import (
	"github.com/spf13/cobra"
)

const completionLong = `Generate shell completion script for golangci-lint-auto-configure.

To load completions:

Bash:
  $ source <(golangci-lint-auto-configure completion bash)
  # To load completions for each session, execute once:
  # Linux:
  $ golangci-lint-auto-configure completion bash > /etc/bash_completion.d/golangci-lint-auto-configure
  # macOS:
  $ golangci-lint-auto-configure completion bash > $(brew --prefix)/etc/bash_completion.d/golangci-lint-auto-configure

Zsh:
  $ source <(golangci-lint-auto-configure completion zsh)
  # To load completions for each session, execute once:
  $ golangci-lint-auto-configure completion zsh > "${fpath[1]}/_golangci-lint-auto-configure"

Fish:
  $ golangci-lint-auto-configure completion fish | source
  # To load completions for each session, execute once:
  $ golangci-lint-auto-configure completion fish > ~/.config/fish/completions/golangci-lint-auto-configure.fish

PowerShell:
  PS> golangci-lint-auto-configure completion powershell | Out-String | Invoke-Expression
  # To load completions for every new session, run:
  PS> golangci-lint-auto-configure completion powershell > golangci-lint-auto-configure.ps1
  # and source this file from your PowerShell profile.
`

// NewCompletionCommand creates the completion command for shell autocompletion.
func NewCompletionCommand() *cobra.Command {
	return &cobra.Command{
		Use:                   "completion [bash|zsh|fish|powershell]",
		Short:                 "Generate shell completion script",
		Long:                  completionLong,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		Run: func(cmd *cobra.Command, args []string) {
			generateCompletion(cmd, args[0])
		},
	}
}

func generateCompletion(cmd *cobra.Command, shell string) {
	out := cmd.OutOrStdout()

	switch shell {
	case "bash":
		_ = cmd.Root().GenBashCompletion(out)
	case "zsh":
		_ = cmd.Root().GenZshCompletion(out)
	case "fish":
		_ = cmd.Root().GenFishCompletion(out, true)
	case "powershell":
		_ = cmd.Root().GenPowerShellCompletionWithDesc(out)
	}
}
