package stack

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
)

// DataDirectories are the directories needed in FOLDER_FOR_DATA
var DataDirectories = []string{
	"authentik/certs",
	"authentik/media",
	"authentik/templates",
	"bazarr",
	"chromium",
	"crowdsec/data",
	"ddns-updater",
	"filebot",
	"gluetun",
	"grafana",
	"headplane/data",
	"headscale/data",
	"heimdall",
	"homarr/configs",
	"homarr/data",
	"homarr/icons",
	"homepage",
	"huntarr",
	"jellyfin",
	"jellyseerr",
	"lidarr",
	"logs/unpackerr",
	"logs/traefik",
	"mylar",
	"plex",
	"portainer",
	"postgresql",
	"prometheus",
	"prowlarr",
	"qbittorrent",
	"radarr",
	"readarr",
	"sabnzbd",
	"sonarr",
	"tailscale",
	"tdarr/server",
	"tdarr/configs",
	"tdarr/logs",
	"tdarr-node",
	"traefik/letsencrypt",
	"traefik-certs-dumper",
	"unpackerr",
	"valkey",
	"whisparr",
}

// MediaDirectories are the directories needed in FOLDER_FOR_MEDIA
var MediaDirectories = []string{
	// Media categories
	"media/anime",
	"media/audio",
	"media/books",
	"media/comics",
	"media/movies",
	"media/music",
	"media/photos",
	"media/tv",
	"media/xxx",
	// Usenet download directories
	"usenet/anime",
	"usenet/audio",
	"usenet/books",
	"usenet/comics",
	"usenet/complete",
	"usenet/console",
	"usenet/incomplete",
	"usenet/movies",
	"usenet/music",
	"usenet/prowlarr",
	"usenet/software",
	"usenet/tv",
	"usenet/xxx",
	// Torrent download directories
	"torrents/anime",
	"torrents/audio",
	"torrents/books",
	"torrents/comics",
	"torrents/complete",
	"torrents/console",
	"torrents/incomplete",
	"torrents/movies",
	"torrents/music",
	"torrents/prowlarr",
	"torrents/software",
	"torrents/tv",
	"torrents/xxx",
	// Other directories
	"watch",
	"filebot/input",
	"filebot/output",
}

// CreateDirectories creates all required directories with proper permissions
func CreateDirectories(dataFolder, mediaFolder string, uid, gid int, verbose bool, dryRun bool) error {
	if verbose {
		color.Cyan("Creating directories...")
		color.Cyan("  Data folder: %s", dataFolder)
		color.Cyan("  Media folder: %s", mediaFolder)
		color.Cyan("  UID:GID: %d:%d", uid, gid)
	}

	// Create data directories
	for _, dir := range DataDirectories {
		fullPath := filepath.Join(dataFolder, dir)
		if err := createDir(fullPath, uid, gid, verbose, dryRun); err != nil {
			return fmt.Errorf("failed to create data directory %s: %w", dir, err)
		}
	}

	// Create media directories
	for _, dir := range MediaDirectories {
		fullPath := filepath.Join(mediaFolder, dir)
		if err := createDir(fullPath, uid, gid, verbose, dryRun); err != nil {
			return fmt.Errorf("failed to create media directory %s: %w", dir, err)
		}
	}

	color.Green("All directories created successfully")
	return nil
}

// createDir creates a single directory with proper permissions
func createDir(path string, uid, gid int, verbose bool, dryRun bool) error {
	if dryRun {
		if verbose {
			fmt.Printf("  [dry-run] Would create: %s\n", path)
		}
		return nil
	}

	// Create directory with parent directories
	if err := os.MkdirAll(path, 0775); err != nil {
		return err
	}

	// Set ownership
	if err := os.Chown(path, uid, gid); err != nil {
		// Don't fail on chown errors (might not have permission)
		if verbose {
			color.Yellow("  Warning: Could not set ownership on %s: %v", path, err)
		}
	}

	// Set permissions (setgid bit for shared group access)
	if err := os.Chmod(path, 02775); err != nil {
		if verbose {
			color.Yellow("  Warning: Could not set permissions on %s: %v", path, err)
		}
	}

	if verbose {
		fmt.Printf("  Created: %s\n", path)
	}

	return nil
}

// SetPermissions recursively sets ownership and permissions on directories
func SetPermissions(paths []string, uid, gid int, verbose bool, dryRun bool) error {
	if verbose {
		color.Cyan("Setting permissions...")
	}

	for _, path := range paths {
		if dryRun {
			if verbose {
				fmt.Printf("  [dry-run] Would set permissions on: %s\n", path)
			}
			continue
		}

		// Walk the directory tree
		err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Set ownership
			if err := os.Chown(p, uid, gid); err != nil {
				if verbose {
					color.Yellow("  Warning: Could not set ownership on %s: %v", p, err)
				}
			}

			// Set permissions based on type
			var perm os.FileMode
			if info.IsDir() {
				perm = 02775 // rwxrwsr-x with setgid
			} else {
				perm = 0664 // rw-rw-r--
			}

			if err := os.Chmod(p, perm); err != nil {
				if verbose {
					color.Yellow("  Warning: Could not set permissions on %s: %v", p, err)
				}
			}

			return nil
		})

		if err != nil {
			return fmt.Errorf("failed to set permissions on %s: %w", path, err)
		}
	}

	if verbose {
		color.Green("Permissions set successfully")
	}

	return nil
}

// VerifyDirectories checks that all required directories exist
func VerifyDirectories(dataFolder, mediaFolder string) []string {
	var missing []string

	for _, dir := range DataDirectories {
		fullPath := filepath.Join(dataFolder, dir)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			missing = append(missing, fullPath)
		}
	}

	for _, dir := range MediaDirectories {
		fullPath := filepath.Join(mediaFolder, dir)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			missing = append(missing, fullPath)
		}
	}

	return missing
}
