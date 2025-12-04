package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/jxmullins/mediastack/internal/config"
	"github.com/jxmullins/mediastack/internal/docker"
	"github.com/jxmullins/mediastack/internal/stack"
	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:     "deploy",
	Aliases: []string{"up"},
	Short:   "Deploy the media stack",
	Long: `Deploy the MediaStack by performing the following steps:

1. Create required directory structure
2. Set proper file permissions
3. Copy configuration files to data folder
4. Validate docker-compose configuration
5. Pull Docker images (optional)
6. Stop any existing containers
7. Start all services

This command replaces the functionality of restart.sh with improved
error handling and proper container management.`,
	RunE: runDeploy,
}

func init() {
	deployCmd.Flags().Bool("pull", false, "Pull images before deploying")
	deployCmd.Flags().Bool("no-directories", false, "Skip directory creation")
	deployCmd.Flags().Bool("no-files", false, "Skip config file copying")
	deployCmd.Flags().Bool("force", false, "Force recreate all containers")
	deployCmd.Flags().Bool("prune", true, "Prune unused resources after successful deploy")
}

func runDeploy(cmd *cobra.Command, args []string) error {
	pullFirst, _ := cmd.Flags().GetBool("pull")
	noDirs, _ := cmd.Flags().GetBool("no-directories")
	noFiles, _ := cmd.Flags().GetBool("no-files")
	force, _ := cmd.Flags().GetBool("force")
	prune, _ := cmd.Flags().GetBool("prune")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// Normalize variant if needed
	if cfg.Variant != "" {
		cfg.Variant = config.NormalizeVariant(cfg.Variant)
	}

	color.Cyan("Deploying MediaStack...")
	color.Cyan("  Variant: %s", cfg.Variant)
	color.Cyan("  Config:  %s", cfg.ConfigDir)
	color.Cyan("  Data:    %s", cfg.DataFolder)
	color.Cyan("  Media:   %s", cfg.MediaFolder)
	fmt.Println()

	if dryRun {
		color.Yellow("[dry-run mode - no changes will be made]")
		fmt.Println()
	}

	// Step 1: Create directories
	if !noDirs {
		color.Cyan("Step 1: Creating directories...")
		if err := stack.CreateDirectories(
			cfg.DataFolder,
			cfg.MediaFolder,
			cfg.PUID,
			cfg.PGID,
			verbose,
			dryRun,
		); err != nil {
			return fmt.Errorf("failed to create directories: %w", err)
		}
	} else {
		color.Yellow("Step 1: Skipping directory creation (--no-directories)")
	}

	// Step 2: Set config file permissions
	color.Cyan("\nStep 2: Setting file permissions...")
	if err := stack.SetConfigPermissions(
		cfg.ConfigDir,
		cfg.PUID,
		cfg.PGID,
		verbose,
		dryRun,
	); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	// Step 3: Copy config files
	if !noFiles {
		color.Cyan("\nStep 3: Copying configuration files...")
		if err := stack.CopyConfigFiles(
			cfg.ConfigDir,
			cfg.DataFolder,
			cfg.PUID,
			cfg.PGID,
			verbose,
			dryRun,
		); err != nil {
			return fmt.Errorf("failed to copy config files: %w", err)
		}
	} else {
		color.Yellow("Step 3: Skipping config file copy (--no-files)")
	}

	if dryRun {
		color.Yellow("\n[dry-run] Would validate, pull, and start containers")
		return nil
	}

	// Step 4: Validate compose configuration
	color.Cyan("\nStep 4: Validating Docker Compose configuration...")
	compose := docker.NewCompose(cfg.ProjectName, cfg.ConfigDir, cfg.ComposeFile())
	compose.SetVerbose(verbose)

	if err := compose.Config(ctx); err != nil {
		return fmt.Errorf("compose configuration is invalid: %w", err)
	}
	color.Green("  Configuration is valid")

	// Step 5: Pull images
	if pullFirst {
		color.Cyan("\nStep 5: Pulling Docker images...")
		if err := compose.Pull(ctx); err != nil {
			return fmt.Errorf("failed to pull images: %w", err)
		}
	} else {
		color.Yellow("\nStep 5: Skipping image pull (use --pull to update)")
	}

	// Step 6: Stop existing containers
	color.Cyan("\nStep 6: Stopping existing containers...")
	client, err := docker.NewClient(cfg.ProjectName)
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer client.Close()

	// Get list of running containers for this project
	containers, err := client.ListContainers(ctx, false)
	if err != nil {
		color.Yellow("  Warning: Could not list containers: %v", err)
	} else if len(containers) > 0 {
		color.Cyan("  Found %d running containers", len(containers))
		for _, c := range containers {
			if verbose {
				fmt.Printf("    Stopping: %s\n", c.Name)
			}
			if err := client.StopContainer(ctx, c.ID); err != nil {
				color.Yellow("    Warning: Failed to stop %s: %v", c.Name, err)
			}
		}
		color.Green("  Stopped existing containers")
	} else {
		color.Green("  No existing containers to stop")
	}

	// Prune old containers
	if err := client.PruneContainers(ctx); err != nil {
		color.Yellow("  Warning: Failed to prune containers: %v", err)
	}

	// Prune volumes and networks
	if err := client.PruneVolumes(ctx); err != nil {
		color.Yellow("  Warning: Failed to prune volumes: %v", err)
	}
	if err := client.PruneNetworks(ctx); err != nil {
		color.Yellow("  Warning: Failed to prune networks: %v", err)
	}

	// Step 7: Start services
	color.Cyan("\nStep 7: Starting services...")
	if err := compose.Up(ctx, true, force); err != nil {
		return fmt.Errorf("failed to start services: %w", err)
	}

	// Step 8: Verify services are running
	color.Cyan("\nStep 8: Verifying services...")
	time.Sleep(5 * time.Second) // Give containers time to start

	services, err := compose.ConfigServices(ctx)
	if err != nil {
		color.Yellow("  Warning: Could not get service list: %v", err)
	} else {
		running := 0
		failed := 0

		for _, service := range services {
			isRunning, err := compose.IsRunning(ctx, service)
			if err != nil || !isRunning {
				color.Red("  Not running: %s", service)
				failed++
			} else {
				running++
			}
		}

		if failed > 0 {
			color.Yellow("  %d/%d services running", running, len(services))
		} else {
			color.Green("  All %d services running", running)
		}
	}

	// Step 9: Prune unused images
	if prune {
		color.Cyan("\nStep 9: Pruning unused images...")
		if err := client.PruneImages(ctx); err != nil {
			color.Yellow("  Warning: Failed to prune images: %v", err)
		} else {
			color.Green("  Unused images removed")
		}
	}

	fmt.Println()
	color.Green("MediaStack deployed successfully!")

	return nil
}
