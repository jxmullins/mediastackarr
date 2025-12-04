package cli

import (
	"github.com/jxmullins/mediastack/internal/shell"
	"github.com/spf13/cobra"
)

var shellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Start interactive shell",
	Long: `Start an interactive shell with slash commands.

The shell provides a REPL interface with:
  • Command history (up/down arrows)
  • Tab completion
  • Slash commands (/help, /status, /deploy, etc.)

Example commands:
  /status     - Show container status
  /deploy     - Deploy the stack
  /logs nginx - View nginx logs
  /help       - Show all commands`,
	Aliases: []string{"sh", "interactive", "repl"},
	RunE:    runShell,
}

func init() {
	rootCmd.AddCommand(shellCmd)
}

func runShell(cmd *cobra.Command, args []string) error {
	sh := shell.New(cfg)
	return sh.Run(Version)
}
