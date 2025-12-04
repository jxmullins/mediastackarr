package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/fatih/color"
	"github.com/jxmullins/mediastack/internal/docker"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show container status",
	Long: `Display the status of all MediaStack containers.

Shows container name, image, state, health status, and ports.`,
	RunE: runStatus,
}

func init() {
	statusCmd.Flags().BoolP("all", "a", false, "Show all containers (including stopped)")
	statusCmd.Flags().Bool("health", false, "Show only health status")
	statusCmd.Flags().Bool("json", false, "Output as JSON")
	statusCmd.Flags().BoolP("watch", "w", false, "Continuously watch status")
}

func runStatus(cmd *cobra.Command, args []string) error {
	showAll, _ := cmd.Flags().GetBool("all")
	healthOnly, _ := cmd.Flags().GetBool("health")
	jsonOutput, _ := cmd.Flags().GetBool("json")
	watch, _ := cmd.Flags().GetBool("watch")

	ctx := context.Background()

	client, err := docker.NewClient(cfg.ProjectName)
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer client.Close()

	if watch {
		return watchStatus(ctx, client, showAll, healthOnly)
	}

	return showStatus(ctx, client, showAll, healthOnly, jsonOutput)
}

func showStatus(ctx context.Context, client *docker.Client, showAll, healthOnly, jsonOutput bool) error {
	containers, err := client.ListContainers(ctx, showAll)
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	if len(containers) == 0 {
		color.Yellow("No containers found for project: %s", cfg.ProjectName)
		return nil
	}

	// Sort by name
	sort.Slice(containers, func(i, j int) bool {
		return containers[i].Name < containers[j].Name
	})

	if jsonOutput {
		return outputJSON(containers)
	}

	if healthOnly {
		return outputHealthTable(containers)
	}

	return outputFullTable(containers)
}

func outputJSON(containers []docker.ContainerInfo) error {
	data, err := json.MarshalIndent(containers, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func outputFullTable(containers []docker.ContainerInfo) error {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Image", "State", "Health", "Status"})
	table.SetAutoWrapText(false)
	table.SetBorder(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetTablePadding("  ")
	table.SetNoWhiteSpace(true)

	for _, c := range containers {
		stateColor := getStateColor(c.State)
		healthColor := getHealthColor(c.Health)

		row := []string{
			c.Name,
			truncateString(c.Image, 40),
			stateColor(c.State),
			healthColor(c.Health),
			c.Status,
		}
		table.Append(row)
	}

	fmt.Printf("\nMediaStack Status (%s)\n\n", cfg.ProjectName)
	table.Render()

	// Summary
	running := 0
	stopped := 0
	healthy := 0
	unhealthy := 0

	for _, c := range containers {
		if c.State == "running" {
			running++
			if c.Health == "healthy" {
				healthy++
			} else if c.Health == "unhealthy" {
				unhealthy++
			}
		} else {
			stopped++
		}
	}

	fmt.Printf("\nTotal: %d | ", len(containers))
	color.Green("Running: %d", running)
	fmt.Print(" | ")
	if stopped > 0 {
		color.Red("Stopped: %d", stopped)
	} else {
		fmt.Printf("Stopped: %d", stopped)
	}
	fmt.Print(" | ")
	if unhealthy > 0 {
		color.Red("Unhealthy: %d", unhealthy)
	} else {
		color.Green("Healthy: %d", healthy)
	}
	fmt.Println()

	return nil
}

func outputHealthTable(containers []docker.ContainerInfo) error {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Health", "Status"})
	table.SetAutoWrapText(false)
	table.SetBorder(false)

	for _, c := range containers {
		if c.State != "running" {
			continue
		}

		healthColor := getHealthColor(c.Health)
		health := c.Health
		if health == "" {
			health = "n/a"
		}

		row := []string{
			c.Name,
			healthColor(health),
			c.Status,
		}
		table.Append(row)
	}

	fmt.Printf("\nHealth Status (%s)\n\n", cfg.ProjectName)
	table.Render()
	return nil
}

func watchStatus(ctx context.Context, client *docker.Client, showAll, healthOnly bool) error {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		// Clear screen
		fmt.Print("\033[H\033[2J")

		if err := showStatus(ctx, client, showAll, healthOnly, false); err != nil {
			color.Red("Error: %v", err)
		}

		fmt.Printf("\nPress Ctrl+C to exit (updating every 2s)\n")

		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			continue
		}
	}
}

func getStateColor(state string) func(format string, a ...interface{}) string {
	switch state {
	case "running":
		return color.GreenString
	case "exited", "dead":
		return color.RedString
	case "paused":
		return color.YellowString
	default:
		return fmt.Sprintf
	}
}

func getHealthColor(health string) func(format string, a ...interface{}) string {
	switch health {
	case "healthy":
		return color.GreenString
	case "unhealthy":
		return color.RedString
	case "starting":
		return color.YellowString
	default:
		return func(format string, a ...interface{}) string {
			if len(a) == 0 {
				return format
			}
			return fmt.Sprintf(format, a...)
		}
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
