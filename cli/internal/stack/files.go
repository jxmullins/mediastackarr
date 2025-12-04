package stack

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/fatih/color"
)

// ConfigFile represents a config file to copy
type ConfigFile struct {
	Source      string // Filename in config directory
	Destination string // Relative path in data folder
	Permission  os.FileMode
}

// ConfigFiles are the configuration files to copy during deployment
var ConfigFiles = []ConfigFile{
	{
		Source:      "headplane-config.yaml",
		Destination: "headplane/config.yaml",
		Permission:  0664,
	},
	{
		Source:      "headscale-config.yaml",
		Destination: "headscale/config.yaml",
		Permission:  0664,
	},
	{
		Source:      "traefik-static.yaml",
		Destination: "traefik/traefik.yaml",
		Permission:  0664,
	},
	{
		Source:      "traefik-dynamic.yaml",
		Destination: "traefik/dynamic.yaml",
		Permission:  0664,
	},
	{
		Source:      "traefik-internal.yaml",
		Destination: "traefik/internal.yaml",
		Permission:  0664,
	},
	{
		Source:      "crowdsec-acquis.yaml",
		Destination: "crowdsec/acquis.yaml",
		Permission:  0664,
	},
}

// SpecialFiles are files that need special handling
var SpecialFiles = []struct {
	Path       string
	Permission os.FileMode
	Create     bool // Whether to create if not exists
}{
	{
		Path:       "traefik/letsencrypt/acme.json",
		Permission: 0600,
		Create:     true,
	},
}

// CopyConfigFiles copies all configuration files to their destinations
func CopyConfigFiles(configDir, dataFolder string, uid, gid int, verbose bool, dryRun bool) error {
	if verbose {
		color.Cyan("Copying configuration files...")
	}

	for _, cf := range ConfigFiles {
		src := filepath.Join(configDir, cf.Source)
		dst := filepath.Join(dataFolder, cf.Destination)

		if dryRun {
			if verbose {
				fmt.Printf("  [dry-run] Would copy: %s -> %s\n", cf.Source, dst)
			}
			continue
		}

		// Check if source exists
		if _, err := os.Stat(src); os.IsNotExist(err) {
			if verbose {
				color.Yellow("  Warning: Source file not found: %s", src)
			}
			continue
		}

		// Copy the file
		if err := copyFile(src, dst, cf.Permission); err != nil {
			return fmt.Errorf("failed to copy %s: %w", cf.Source, err)
		}

		// Set ownership
		if err := os.Chown(dst, uid, gid); err != nil {
			if verbose {
				color.Yellow("  Warning: Could not set ownership on %s: %v", dst, err)
			}
		}

		if verbose {
			fmt.Printf("  Copied: %s -> %s\n", cf.Source, dst)
		}
	}

	// Handle special files
	for _, sf := range SpecialFiles {
		fullPath := filepath.Join(dataFolder, sf.Path)

		if dryRun {
			if verbose {
				fmt.Printf("  [dry-run] Would create/set permissions: %s\n", fullPath)
			}
			continue
		}

		if sf.Create {
			// Create the file if it doesn't exist
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				// Ensure parent directory exists
				if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
					return fmt.Errorf("failed to create directory for %s: %w", sf.Path, err)
				}

				// Create empty file
				file, err := os.Create(fullPath)
				if err != nil {
					return fmt.Errorf("failed to create %s: %w", sf.Path, err)
				}
				file.Close()

				if verbose {
					fmt.Printf("  Created: %s\n", fullPath)
				}
			}
		}

		// Set permissions
		if err := os.Chmod(fullPath, sf.Permission); err != nil {
			return fmt.Errorf("failed to set permissions on %s: %w", sf.Path, err)
		}

		// Set ownership
		if err := os.Chown(fullPath, uid, gid); err != nil {
			if verbose {
				color.Yellow("  Warning: Could not set ownership on %s: %v", fullPath, err)
			}
		}

		if verbose {
			fmt.Printf("  Set permissions %o on: %s\n", sf.Permission, fullPath)
		}
	}

	color.Green("Configuration files copied successfully")
	return nil
}

// copyFile copies a file from src to dst with the specified permissions
func copyFile(src, dst string, perm os.FileMode) error {
	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source: %w", err)
	}
	defer srcFile.Close()

	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Create destination file
	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return fmt.Errorf("failed to create destination: %w", err)
	}
	defer dstFile.Close()

	// Copy contents
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy contents: %w", err)
	}

	return nil
}

// SetConfigPermissions sets proper permissions on config files in the config directory
func SetConfigPermissions(configDir string, uid, gid int, verbose bool, dryRun bool) error {
	if verbose {
		color.Cyan("Setting config file permissions...")
	}

	patterns := []string{"*.yaml", "*.yml", ".env", "*.sh"}

	for _, pattern := range patterns {
		matches, err := filepath.Glob(filepath.Join(configDir, pattern))
		if err != nil {
			continue
		}

		for _, match := range matches {
			if dryRun {
				if verbose {
					fmt.Printf("  [dry-run] Would set permissions on: %s\n", match)
				}
				continue
			}

			// Determine permission based on file type
			perm := os.FileMode(0664)
			if filepath.Ext(match) == ".sh" {
				perm = 0775
			}

			if err := os.Chmod(match, perm); err != nil {
				if verbose {
					color.Yellow("  Warning: Could not set permissions on %s: %v", match, err)
				}
			}

			if err := os.Chown(match, uid, gid); err != nil {
				if verbose {
					color.Yellow("  Warning: Could not set ownership on %s: %v", match, err)
				}
			}

			if verbose {
				fmt.Printf("  Set permissions %o on: %s\n", perm, filepath.Base(match))
			}
		}
	}

	return nil
}

// VerifyConfigFiles checks that all required config files exist
func VerifyConfigFiles(configDir string) []string {
	var missing []string

	for _, cf := range ConfigFiles {
		src := filepath.Join(configDir, cf.Source)
		if _, err := os.Stat(src); os.IsNotExist(err) {
			missing = append(missing, cf.Source)
		}
	}

	return missing
}
