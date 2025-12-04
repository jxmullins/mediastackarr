package cli

import (
	"bufio"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var apikeysCmd = &cobra.Command{
	Use:   "apikeys",
	Short: "Extract API keys from running services",
	Long: `Extract API keys from *ARR and other services.

Reads configuration files from the data directory to extract
API keys for services like Radarr, Sonarr, Bazarr, etc.`,
	RunE: runApikeys,
}

func init() {
	apikeysCmd.Flags().Bool("json", false, "Output as JSON")
	apikeysCmd.Flags().String("service", "", "Get key for specific service only")
}

// APIKeyInfo holds information about an extracted API key
type APIKeyInfo struct {
	Service  string `json:"service"`
	APIKey   string `json:"api_key"`
	Location string `json:"location"`
}

// ServiceConfig defines how to extract API key for a service
type ServiceConfig struct {
	Name       string
	ConfigPath string
	Format     string // xml, yaml, or ini
	KeyPath    string // XPath for XML, key for YAML, or key= for INI
}

var serviceConfigs = []ServiceConfig{
	// XML-based services (*ARR apps)
	{Name: "Lidarr", ConfigPath: "lidarr/config.xml", Format: "xml", KeyPath: "ApiKey"},
	{Name: "Prowlarr", ConfigPath: "prowlarr/config.xml", Format: "xml", KeyPath: "ApiKey"},
	{Name: "Radarr", ConfigPath: "radarr/config.xml", Format: "xml", KeyPath: "ApiKey"},
	{Name: "Readarr", ConfigPath: "readarr/config.xml", Format: "xml", KeyPath: "ApiKey"},
	{Name: "Sonarr", ConfigPath: "sonarr/config.xml", Format: "xml", KeyPath: "ApiKey"},
	{Name: "Whisparr", ConfigPath: "whisparr/config.xml", Format: "xml", KeyPath: "ApiKey"},
	// YAML-based services
	{Name: "Bazarr", ConfigPath: "bazarr/config/config.yaml", Format: "yaml", KeyPath: "auth.apikey"},
	// INI-based services
	{Name: "Mylar", ConfigPath: "mylar/mylar/config.ini", Format: "ini", KeyPath: "api_key"},
}

func runApikeys(cmd *cobra.Command, args []string) error {
	jsonOutput, _ := cmd.Flags().GetBool("json")
	serviceFilter, _ := cmd.Flags().GetString("service")

	var keys []APIKeyInfo

	for _, svc := range serviceConfigs {
		// Filter by service if specified
		if serviceFilter != "" && !strings.EqualFold(svc.Name, serviceFilter) {
			continue
		}

		fullPath := filepath.Join(cfg.DataFolder, svc.ConfigPath)

		// Check if file exists
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			if verbose {
				color.Yellow("Config not found for %s: %s", svc.Name, fullPath)
			}
			continue
		}

		// Extract API key based on format
		var apiKey string
		var err error

		switch svc.Format {
		case "xml":
			apiKey, err = extractXMLKey(fullPath, svc.KeyPath)
		case "yaml":
			apiKey, err = extractYAMLKey(fullPath, svc.KeyPath)
		case "ini":
			apiKey, err = extractINIKey(fullPath, svc.KeyPath)
		}

		if err != nil {
			if verbose {
				color.Yellow("Failed to extract key for %s: %v", svc.Name, err)
			}
			continue
		}

		if apiKey != "" {
			keys = append(keys, APIKeyInfo{
				Service:  svc.Name,
				APIKey:   apiKey,
				Location: fullPath,
			})
		}
	}

	if len(keys) == 0 {
		color.Yellow("No API keys found")
		return nil
	}

	if jsonOutput {
		return outputKeysJSON(keys)
	}

	return outputKeysTable(keys)
}

func outputKeysJSON(keys []APIKeyInfo) error {
	data, err := json.MarshalIndent(keys, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func outputKeysTable(keys []APIKeyInfo) error {
	fmt.Println()
	color.Cyan("Extracted API Keys:")
	fmt.Println()

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Service", "API Key", "Location"})
	table.SetAutoWrapText(false)
	table.SetBorder(false)

	for _, k := range keys {
		table.Append([]string{
			k.Service,
			k.APIKey,
			k.Location,
		})
	}

	table.Render()
	return nil
}

// extractXMLKey extracts an API key from an XML config file
func extractXMLKey(path, keyName string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	// Parse XML to find the key
	// The *ARR apps use a simple Config structure
	type Config struct {
		XMLName xml.Name
		ApiKey  string `xml:"ApiKey"`
	}

	var config Config
	if err := xml.Unmarshal(data, &config); err != nil {
		return "", err
	}

	return config.ApiKey, nil
}

// extractYAMLKey extracts an API key from a YAML config file
func extractYAMLKey(path, keyPath string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	var config map[string]interface{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return "", err
	}

	// Navigate nested keys (e.g., "auth.apikey")
	keys := strings.Split(keyPath, ".")
	current := config

	for i, key := range keys {
		if i == len(keys)-1 {
			// Last key - get the value
			if val, ok := current[key]; ok {
				return fmt.Sprintf("%v", val), nil
			}
			return "", fmt.Errorf("key not found: %s", keyPath)
		}

		// Navigate to next level
		if next, ok := current[key].(map[string]interface{}); ok {
			current = next
		} else {
			return "", fmt.Errorf("key not found: %s", keyPath)
		}
	}

	return "", fmt.Errorf("key not found: %s", keyPath)
}

// extractINIKey extracts an API key from an INI config file
func extractINIKey(path, keyName string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	prefix := keyName + "="
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, prefix) {
			value := strings.TrimPrefix(line, prefix)
			value = strings.TrimSpace(value)
			// Remove quotes if present
			value = strings.Trim(value, "\"'")
			return value, nil
		}
	}

	return "", fmt.Errorf("key not found: %s", keyName)
}
