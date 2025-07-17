package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate completion script",
	Long: `To load completions:

Bash:
  $ source <(codepack completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ codepack completion bash > /etc/bash_completion.d/codepack
  # macOS:
  $ codepack completion bash > /usr/local/etc/bash_completion.d/codepack

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ codepack completion zsh > "${fpath[1]}/_codepack"

  # You will need to start a new shell for this setup to take effect.

Fish:
  $ codepack completion fish | source

  # To load completions for each session, execute once:
  $ codepack completion fish > ~/.config/fish/completions/codepack.fish

PowerShell:
  PS> codepack completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> codepack completion powershell > codepack.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		}
	},
}