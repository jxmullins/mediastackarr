package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// StackVariant represents the docker-compose variant to use
type StackVariant string

const (
	VariantFull  StackVariant = "full-download-vpn"
	VariantMini  StackVariant = "mini-download-vpn"
	VariantNoVPN StackVariant = "no-download-vpn"
)

// Config holds all configuration for the media stack
type Config struct {
	// Paths
	ConfigDir     string // Directory containing .env and yaml files
	MediaFolder   string // FOLDER_FOR_MEDIA - where media is stored
	DataFolder    string // FOLDER_FOR_DATA - where app data is stored

	// User/Group
	PUID int
	PGID int

	// Stack settings
	Variant     string // Which compose variant to use
	ProjectName string // Docker compose project name

	// Network
	DockerSubnet  string
	DockerGateway string
	LocalSubnet   string
	Timezone      string

	// Database
	PostgresPassword string

	// All environment variables (raw)
	Env map[string]string
}

// Load reads configuration from the specified directory
func Load(configDir string) (*Config, error) {
	envPath := filepath.Join(configDir, ".env")

	env, err := ParseEnvFile(envPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse .env file: %w", err)
	}

	cfg := &Config{
		ConfigDir: configDir,
		Env:       env,
	}

	// Required fields
	required := []string{"FOLDER_FOR_MEDIA", "FOLDER_FOR_DATA", "PUID", "PGID"}
	missing := []string{}
	for _, key := range required {
		if _, ok := env[key]; !ok {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required environment variables: %v", missing)
	}

	// Parse required fields
	cfg.MediaFolder = env["FOLDER_FOR_MEDIA"]
	cfg.DataFolder = env["FOLDER_FOR_DATA"]

	puid, err := strconv.Atoi(env["PUID"])
	if err != nil {
		return nil, fmt.Errorf("invalid PUID value: %w", err)
	}
	cfg.PUID = puid

	pgid, err := strconv.Atoi(env["PGID"])
	if err != nil {
		return nil, fmt.Errorf("invalid PGID value: %w", err)
	}
	cfg.PGID = pgid

	// Optional fields with defaults
	cfg.Timezone = getEnvDefault(env, "TIMEZONE", "UTC")
	cfg.DockerSubnet = getEnvDefault(env, "DOCKER_SUBNET", "172.28.0.0/16")
	cfg.DockerGateway = getEnvDefault(env, "DOCKER_GATEWAY", "172.28.0.1")
	cfg.LocalSubnet = getEnvDefault(env, "LOCAL_SUBNET", "192.168.0.0/16")
	cfg.ProjectName = getEnvDefault(env, "COMPOSE_PROJECT_NAME", "mediastack")
	cfg.PostgresPassword = getEnvDefault(env, "POSTGRESQL_PASSWORD", "")

	// Determine variant from directory structure
	cfg.Variant = detectVariant(configDir)

	return cfg, nil
}

// detectVariant determines which compose variant exists
func detectVariant(configDir string) string {
	parentDir := filepath.Dir(configDir)

	// Check for compose files in order of preference
	variants := []string{"full-download-vpn", "mini-download-vpn", "no-download-vpn"}
	for _, v := range variants {
		composePath := filepath.Join(parentDir, v, "docker-compose.yaml")
		if _, err := os.Stat(composePath); err == nil {
			return v
		}
	}

	return "full-download-vpn" // default
}

// ComposeFile returns the path to the docker-compose file for the current variant
func (c *Config) ComposeFile() string {
	parentDir := filepath.Dir(c.ConfigDir)
	return filepath.Join(parentDir, c.Variant, "docker-compose.yaml")
}

// VariantDir returns the directory containing the compose file
func (c *Config) VariantDir() string {
	parentDir := filepath.Dir(c.ConfigDir)
	return filepath.Join(parentDir, c.Variant)
}

// Validate checks that all required configuration is present and valid
func (c *Config) Validate() []error {
	var errors []error

	// Check directories exist or can be created
	if c.MediaFolder == "" {
		errors = append(errors, fmt.Errorf("FOLDER_FOR_MEDIA is not set"))
	}
	if c.DataFolder == "" {
		errors = append(errors, fmt.Errorf("FOLDER_FOR_DATA is not set"))
	}

	// Check compose file exists
	if _, err := os.Stat(c.ComposeFile()); os.IsNotExist(err) {
		errors = append(errors, fmt.Errorf("compose file not found: %s", c.ComposeFile()))
	}

	// Check variant is valid
	validVariants := map[string]bool{
		"full-download-vpn": true,
		"mini-download-vpn": true,
		"no-download-vpn":   true,
		"full":              true,
		"mini":              true,
		"no-vpn":            true,
	}
	if !validVariants[c.Variant] {
		errors = append(errors, fmt.Errorf("invalid variant: %s", c.Variant))
	}

	return errors
}

// NormalizeVariant converts short variant names to full directory names
func NormalizeVariant(v string) string {
	switch v {
	case "full":
		return "full-download-vpn"
	case "mini":
		return "mini-download-vpn"
	case "no-vpn":
		return "no-download-vpn"
	default:
		return v
	}
}

func getEnvDefault(env map[string]string, key, defaultVal string) string {
	if val, ok := env[key]; ok && val != "" {
		return val
	}
	return defaultVal
}
