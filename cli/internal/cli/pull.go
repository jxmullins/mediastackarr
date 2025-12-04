package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/jxmullins/mediastack/internal/docker"
	"github.com/spf13/cobra"
)

var pullCmd = &cobra.Command{
	Use:   "pull [service...]",
	Short: "Pull/update Docker images",
	Long: `Pull the latest Docker images for all or specific services.

If no service names are provided, all images will be pulled.`,
	RunE: runPull,
}

func init() {
	pullCmd.Flags().Int("parallel", 3, "Number of parallel image pulls")
}

func runPull(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	if dryRun {
		color.Cyan("[dry-run] Would pull Docker images")
		return nil
	}

	compose := docker.NewCompose(cfg.ProjectName, cfg.ConfigDir, cfg.ComposeFile())
	compose.SetVerbose(verbose)

	if len(args) > 0 {
		// Pull specific services
		for _, service := range args {
			color.Cyan("Pulling image for: %s", service)
			if err := compose.PullService(ctx, service); err != nil {
				return fmt.Errorf("failed to pull %s: %w", service, err)
			}
		}
	} else {
		// Pull all services
		if err := compose.Pull(ctx); err != nil {
			return fmt.Errorf("failed to pull images: %w", err)
		}
	}

	color.Green("\nAll images pulled successfully")
	return nil
}
