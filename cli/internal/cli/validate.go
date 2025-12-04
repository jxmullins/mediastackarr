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

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration and compose files",
	Long: `Validates the .env configuration file and docker-compose.yaml.

Checks performed:
- Required environment variables are set
- Docker daemon is accessible
- Docker Compose configuration is valid
- Required config files exist
- Directory structure can be verified`,
	RunE: runValidate,
}

func init() {
	validateCmd.Flags().Bool("strict", false, "Fail on warnings")
}

func runValidate(cmd *cobra.Command, args []string) error {
	strict, _ := cmd.Flags().GetBool("strict")
	hasWarnings := false
	hasErrors := false

	color.Cyan("Validating MediaStack configuration...\n")

	// 1. Validate config
	fmt.Println("Checking configuration...")
	if cfg == nil {
		color.Red("  Error: Configuration not loaded")
		hasErrors = true
	} else {
		color.Green("  Config loaded from: %s", cfg.ConfigDir)

		errors := cfg.Validate()
		if len(errors) > 0 {
			for _, err := range errors {
				color.Red("  Error: %v", err)
				hasErrors = true
			}
		} else {
			color.Green("  Configuration is valid")
		}
	}

	// 2. Check Docker daemon
	fmt.Println("\nChecking Docker daemon...")
	if err := docker.CheckDockerRunning(); err != nil {
		color.Red("  Error: %v", err)
		hasErrors = true
	} else {
		color.Green("  Docker daemon is running")
	}

	// 3. Check Docker Compose
	fmt.Println("\nChecking Docker Compose...")
	if err := docker.CheckComposeInstalled(); err != nil {
		color.Red("  Error: %v", err)
		hasErrors = true
	} else {
		color.Green("  Docker Compose is installed")
	}

	// 4. Validate compose file
	if cfg != nil {
		fmt.Println("\nValidating compose configuration...")
		compose := docker.NewCompose(cfg.ProjectName, cfg.ConfigDir, cfg.ComposeFile())
		compose.SetVerbose(verbose)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := compose.Config(ctx); err != nil {
			color.Red("  Error: Compose configuration is invalid")
			color.Red("  %v", err)
			hasErrors = true
		} else {
			color.Green("  Compose configuration is valid")

			// List services
			services, err := compose.ConfigServices(ctx)
			if err == nil {
				color.Green("  Found %d services", len(services))
			}
		}
	}

	// 5. Check config files
	if cfg != nil {
		fmt.Println("\nChecking configuration files...")
		missing := stack.VerifyConfigFiles(cfg.ConfigDir)
		if len(missing) > 0 {
			for _, f := range missing {
				color.Yellow("  Warning: Missing config file: %s", f)
				hasWarnings = true
			}
		} else {
			color.Green("  All configuration files present")
		}
	}

	// 6. Check required directories
	if cfg != nil {
		fmt.Println("\nChecking directory structure...")
		missing := stack.VerifyDirectories(cfg.DataFolder, cfg.MediaFolder)
		if len(missing) > 0 {
			color.Yellow("  Warning: %d directories need to be created", len(missing))
			if verbose {
				for _, d := range missing {
					fmt.Printf("    - %s\n", d)
				}
			}
			hasWarnings = true
		} else {
			color.Green("  All directories exist")
		}
	}

	// 7. Check environment variables
	if cfg != nil {
		fmt.Println("\nChecking environment variables...")
		requiredVars := []string{
			"FOLDER_FOR_MEDIA",
			"FOLDER_FOR_DATA",
			"PUID",
			"PGID",
			"TIMEZONE",
		}

		missingVars := config.ValidateRequiredVars(cfg.Env, requiredVars)
		if len(missingVars) > 0 {
			for _, v := range missingVars {
				color.Red("  Error: Missing required variable: %s", v)
				hasErrors = true
			}
		} else {
			color.Green("  Required environment variables are set")
		}

		// Check optional but recommended variables
		recommendedVars := []string{
			"CLOUDFLARE_ZONE",
			"CLOUDFLARE_EMAIL",
		}

		missingRecommended := config.ValidateRequiredVars(cfg.Env, recommendedVars)
		if len(missingRecommended) > 0 {
			for _, v := range missingRecommended {
				color.Yellow("  Warning: Recommended variable not set: %s", v)
				hasWarnings = true
			}
		}
	}

	// Summary
	fmt.Println()
	if hasErrors {
		color.Red("Validation failed with errors")
		return fmt.Errorf("validation failed")
	}

	if hasWarnings && strict {
		color.Yellow("Validation failed (strict mode) - warnings found")
		return fmt.Errorf("validation failed with warnings")
	}

	if hasWarnings {
		color.Yellow("Validation passed with warnings")
	} else {
		color.Green("Validation passed successfully")
	}

	return nil
}
