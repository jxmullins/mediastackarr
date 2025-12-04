package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/jxmullins/mediastack/internal/docker"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop [service...]",
	Short: "Stop the media stack",
	Long: `Stop all or specific MediaStack containers.

If no service names are provided, all services will be stopped.
Use --prune to also remove unused containers, volumes, and networks.`,
	RunE: runStop,
}

func init() {
	stopCmd.Flags().Bool("remove-orphans", true, "Remove orphaned containers")
	stopCmd.Flags().BoolP("volumes", "v", false, "Also remove volumes")
	stopCmd.Flags().Bool("prune", false, "Prune unused resources after stop")
}

func runStop(cmd *cobra.Command, args []string) error {
	removeOrphans, _ := cmd.Flags().GetBool("remove-orphans")
	removeVolumes, _ := cmd.Flags().GetBool("volumes")
	prune, _ := cmd.Flags().GetBool("prune")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if dryRun {
		color.Cyan("[dry-run] Would stop MediaStack containers")
		if prune {
			color.Cyan("[dry-run] Would prune unused resources")
		}
		return nil
	}

	compose := docker.NewCompose(cfg.ProjectName, cfg.ConfigDir, cfg.ComposeFile())
	compose.SetVerbose(verbose)

	// Stop specific services or all
	if len(args) > 0 {
		for _, service := range args {
			color.Cyan("Stopping service: %s", service)
			if err := compose.StopService(ctx, service); err != nil {
				return fmt.Errorf("failed to stop %s: %w", service, err)
			}
			color.Green("Stopped: %s", service)
		}
	} else {
		// Stop all services using docker compose down
		if err := compose.Down(ctx, removeVolumes, removeOrphans); err != nil {
			return fmt.Errorf("failed to stop stack: %w", err)
		}
	}

	// Prune if requested
	if prune {
		color.Cyan("\nPruning unused resources...")

		client, err := docker.NewClient(cfg.ProjectName)
		if err != nil {
			return fmt.Errorf("failed to create Docker client: %w", err)
		}
		defer client.Close()

		if err := client.PruneContainers(ctx); err != nil {
			color.Yellow("Warning: Failed to prune containers: %v", err)
		}

		if err := client.PruneVolumes(ctx); err != nil {
			color.Yellow("Warning: Failed to prune volumes: %v", err)
		}

		if err := client.PruneNetworks(ctx); err != nil {
			color.Yellow("Warning: Failed to prune networks: %v", err)
		}

		color.Green("Pruning complete")
	}

	color.Green("\nMediaStack stopped successfully")
	return nil
}
