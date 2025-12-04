package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// ASCII art banner for MediaStack
var bannerArt = `
  __  __          _ _       ____  _             _
 |  \/  | ___  __| (_) __ _/ ___|| |_ __ _  ___| | __
 | |\/| |/ _ \/ _' | |/ _' \___ \| __/ _' |/ __| |/ /
 | |  | |  __/ (_| | | (_| |___) | || (_| | (__|   <
 |_|  |_|\___|\__,_|_|\__,_|____/ \__\__,_|\___|_|\_\
`

var (
	// Colors
	primaryColor   = lipgloss.Color("#7C3AED") // Purple
	secondaryColor = lipgloss.Color("#10B981") // Green
	accentColor    = lipgloss.Color("#F59E0B") // Amber
	mutedColor     = lipgloss.Color("#6B7280") // Gray
	errorColor     = lipgloss.Color("#EF4444") // Red
	successColor   = lipgloss.Color("#10B981") // Green

	// Styles
	BannerStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true)

	TitleStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Italic(true)

	PromptStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Bold(true)

	CommandStyle = lipgloss.NewStyle().
			Foreground(accentColor)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true)

	MutedStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true)

	HelpDescStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(0, 1)
)

// PrintBanner displays the startup banner
func PrintBanner(version string) {
	fmt.Println(BannerStyle.Render(bannerArt))
	fmt.Println(SubtitleStyle.Render(fmt.Sprintf("  Docker Media Stack Manager • v%s", version)))
	fmt.Println()
	fmt.Println(MutedStyle.Render("  Type /help for commands, /quit to exit"))
	fmt.Println()
}

// PrintWelcome displays a welcome message with current config
func PrintWelcome(variant, configDir string) {
	info := fmt.Sprintf("  Variant: %s  •  Config: %s", variant, configDir)
	fmt.Println(MutedStyle.Render(info))
	fmt.Println()
}

// PrintPrompt returns the styled prompt string
func PrintPrompt() string {
	return PromptStyle.Render("mediastack") + MutedStyle.Render(" > ")
}

// PrintError displays an error message
func PrintError(msg string) {
	fmt.Println(ErrorStyle.Render("✗ " + msg))
}

// PrintSuccess displays a success message
func PrintSuccess(msg string) {
	fmt.Println(SuccessStyle.Render("✓ " + msg))
}

// PrintInfo displays an info message
func PrintInfo(msg string) {
	fmt.Println(MutedStyle.Render("→ " + msg))
}

// PrintCommand displays a command being executed
func PrintCommand(cmd string) {
	fmt.Println(CommandStyle.Render("▶ " + cmd))
}
