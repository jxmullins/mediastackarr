package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/jxmullins/mediastack/internal/config"
	"github.com/spf13/cobra"
)

var (
	// Global flags
	cfgDir    string
	variant   string
	dryRun    bool
	verbose   bool

	// Config instance
	cfg *config.Config

	// Version info
	Version   = "dev"
	BuildTime = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "mediastack",
	Short: "MediaStack CLI - Manage your Docker media stack",
	Long: `MediaStack CLI is a tool for managing Docker-based media server stacks.

It provides commands to deploy, stop, restart, and monitor your media stack
including services like Jellyfin, Plex, *ARR apps, Traefik, and more.

Variants:
  full    - Full VPN: All traffic routed through Gluetun
  mini    - Mini VPN: Only downloads through Gluetun
  no-vpn  - No VPN: Direct internet access`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip config loading for version command
		if cmd.Name() == "version" {
			return nil
		}

		// Resolve config directory
		if cfgDir == "" {
			// Try to find base-working-files relative to current directory or parent
			cwd, _ := os.Getwd()
			candidates := []string{
				filepath.Join(cwd, "base-working-files"),
				filepath.Join(cwd, "..", "base-working-files"),
				"/docker",
			}
			for _, c := range candidates {
				if _, err := os.Stat(filepath.Join(c, ".env")); err == nil {
					cfgDir = c
					break
				}
			}
		}

		if cfgDir == "" {
			return fmt.Errorf("could not find config directory with .env file\nUse --config to specify the path")
		}

		// Load configuration
		var err error
		cfg, err = config.Load(cfgDir)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Override variant if specified
		if variant != "" {
			cfg.Variant = variant
		}

		if verbose {
			color.Cyan("Config directory: %s", cfgDir)
			color.Cyan("Variant: %s", cfg.Variant)
		}

		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgDir, "config", "c", "", "Path to config directory containing .env and yaml files")
	rootCmd.PersistentFlags().StringVarP(&variant, "variant", "v", "", "Stack variant: full, mini, or no-vpn")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Show what would be done without executing")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Enable verbose output")

	// Add subcommands
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(deployCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(restartCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(logsCmd)
	rootCmd.AddCommand(pullCmd)
	rootCmd.AddCommand(apikeysCmd)
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

// GetConfig returns the loaded configuration
func GetConfig() *config.Config {
	return cfg
}

// IsDryRun returns whether dry-run mode is enabled
func IsDryRun() bool {
	return dryRun
}

// IsVerbose returns whether verbose mode is enabled
func IsVerbose() bool {
	return verbose
}

// Version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("mediastack version %s (built %s)\n", Version, BuildTime)
	},
}
