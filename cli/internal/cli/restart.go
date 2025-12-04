package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/jxmullins/mediastack/internal/docker"
	"github.com/spf13/cobra"
)

var restartCmd = &cobra.Command{
	Use:   "restart [service...]",
	Short: "Restart the media stack",
	Long: `Restart all or specific MediaStack containers.

If no service names are provided, all services will be restarted.
Use --pull to update images before restarting.`,
	RunE: runRestart,
}

func init() {
	restartCmd.Flags().Bool("pull", false, "Pull images before restarting")
	restartCmd.Flags().Bool("force", false, "Force recreate containers")
}

func runRestart(cmd *cobra.Command, args []string) error {
	pullFirst, _ := cmd.Flags().GetBool("pull")
	force, _ := cmd.Flags().GetBool("force")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	if dryRun {
		color.Cyan("[dry-run] Would restart MediaStack containers")
		if pullFirst {
			color.Cyan("[dry-run] Would pull images first")
		}
		return nil
	}

	compose := docker.NewCompose(cfg.ProjectName, cfg.ConfigDir, cfg.ComposeFile())
	compose.SetVerbose(verbose)

	// Pull images if requested
	if pullFirst {
		color.Cyan("Pulling images...")
		if err := compose.Pull(ctx); err != nil {
			return fmt.Errorf("failed to pull images: %w", err)
		}
	}

	if len(args) > 0 {
		// Restart specific services
		for _, service := range args {
			color.Cyan("Restarting service: %s", service)
			if err := compose.RestartService(ctx, service); err != nil {
				return fmt.Errorf("failed to restart %s: %w", service, err)
			}
			color.Green("Restarted: %s", service)
		}
	} else if force {
		// Force recreate all containers
		color.Cyan("Force recreating all containers...")
		if err := compose.Down(ctx, false, true); err != nil {
			return fmt.Errorf("failed to stop stack: %w", err)
		}
		if err := compose.Up(ctx, true, false); err != nil {
			return fmt.Errorf("failed to start stack: %w", err)
		}
	} else {
		// Simple restart
		if err := compose.Restart(ctx); err != nil {
			return fmt.Errorf("failed to restart stack: %w", err)
		}
	}

	color.Green("\nMediaStack restarted successfully")
	return nil
}
