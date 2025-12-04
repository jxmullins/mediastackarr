package cli

import (
	"context"
	"fmt"

	"github.com/jxmullins/mediastack/internal/docker"
	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:   "logs [service]",
	Short: "View container logs",
	Long: `View logs from MediaStack containers.

If no service is specified, logs from all services are shown.
Use -f to follow logs in real-time.`,
	RunE: runLogs,
}

func init() {
	logsCmd.Flags().BoolP("follow", "f", false, "Follow log output")
	logsCmd.Flags().StringP("tail", "n", "100", "Number of lines to show from the end")
	logsCmd.Flags().BoolP("timestamps", "t", false, "Show timestamps")
	logsCmd.Flags().String("since", "", "Show logs since timestamp (e.g., 2023-01-01T00:00:00 or 10m)")
}

func runLogs(cmd *cobra.Command, args []string) error {
	follow, _ := cmd.Flags().GetBool("follow")
	tail, _ := cmd.Flags().GetString("tail")
	timestamps, _ := cmd.Flags().GetBool("timestamps")

	ctx := context.Background()

	compose := docker.NewCompose(cfg.ProjectName, cfg.ConfigDir, cfg.ComposeFile())
	compose.SetVerbose(verbose)

	service := ""
	if len(args) > 0 {
		service = args[0]
	}

	if err := compose.Logs(ctx, service, follow, tail, timestamps); err != nil {
		return fmt.Errorf("failed to get logs: %w", err)
	}

	return nil
}
