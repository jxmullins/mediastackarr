package shell

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jxmullins/mediastack/internal/config"
	"github.com/jxmullins/mediastack/internal/docker"
	"github.com/jxmullins/mediastack/internal/stack"
	"github.com/jxmullins/mediastack/internal/ui"
)

// Command represents a slash command
type Command struct {
	Name        string
	Aliases     []string
	Description string
	Usage       string
	Handler     func(args []string) error
}

// Shell is the interactive REPL
type Shell struct {
	cfg      *config.Config
	commands map[string]*Command
	history  []string
	histIdx  int
}

// New creates a new interactive shell
func New(cfg *config.Config) *Shell {
	s := &Shell{
		cfg:      cfg,
		commands: make(map[string]*Command),
		history:  make([]string, 0),
		histIdx:  -1,
	}
	s.registerCommands()
	return s
}

// registerCommands sets up all slash commands
func (s *Shell) registerCommands() {
	commands := []*Command{
		{
			Name:        "help",
			Aliases:     []string{"h", "?"},
			Description: "Show available commands",
			Usage:       "/help [command]",
			Handler:     s.cmdHelp,
		},
		{
			Name:        "status",
			Aliases:     []string{"s", "ps"},
			Description: "Show container status",
			Usage:       "/status",
			Handler:     s.cmdStatus,
		},
		{
			Name:        "deploy",
			Aliases:     []string{"up", "start"},
			Description: "Deploy the media stack",
			Usage:       "/deploy [--pull]",
			Handler:     s.cmdDeploy,
		},
		{
			Name:        "stop",
			Aliases:     []string{"down"},
			Description: "Stop the media stack",
			Usage:       "/stop [service]",
			Handler:     s.cmdStop,
		},
		{
			Name:        "restart",
			Aliases:     []string{"r"},
			Description: "Restart the media stack",
			Usage:       "/restart [service]",
			Handler:     s.cmdRestart,
		},
		{
			Name:        "logs",
			Aliases:     []string{"l", "log"},
			Description: "View container logs",
			Usage:       "/logs <service>",
			Handler:     s.cmdLogs,
		},
		{
			Name:        "pull",
			Aliases:     []string{"update"},
			Description: "Pull latest images",
			Usage:       "/pull [service]",
			Handler:     s.cmdPull,
		},
		{
			Name:        "validate",
			Aliases:     []string{"check", "v"},
			Description: "Validate configuration",
			Usage:       "/validate",
			Handler:     s.cmdValidate,
		},
		{
			Name:        "apikeys",
			Aliases:     []string{"keys", "api"},
			Description: "Show API keys",
			Usage:       "/apikeys [service]",
			Handler:     s.cmdApikeys,
		},
		{
			Name:        "config",
			Aliases:     []string{"cfg"},
			Description: "Show current configuration",
			Usage:       "/config",
			Handler:     s.cmdConfig,
		},
		{
			Name:        "services",
			Aliases:     []string{"svc"},
			Description: "List all services",
			Usage:       "/services",
			Handler:     s.cmdServices,
		},
		{
			Name:        "exec",
			Aliases:     []string{"sh", "shell"},
			Description: "Execute command in container",
			Usage:       "/exec <service> <command>",
			Handler:     s.cmdExec,
		},
		{
			Name:        "clear",
			Aliases:     []string{"cls"},
			Description: "Clear the screen",
			Usage:       "/clear",
			Handler:     s.cmdClear,
		},
		{
			Name:        "quit",
			Aliases:     []string{"exit", "q"},
			Description: "Exit the shell",
			Usage:       "/quit",
			Handler:     s.cmdQuit,
		},
	}

	for _, cmd := range commands {
		s.commands[cmd.Name] = cmd
		for _, alias := range cmd.Aliases {
			s.commands[alias] = cmd
		}
	}
}

// model is the Bubble Tea model for the input
type model struct {
	textInput textinput.Model
	shell     *Shell
	quitting  bool
	err       error
}

func initialModel(s *Shell) model {
	ti := textinput.New()
	ti.Placeholder = "Enter a command..."
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 60
	ti.Prompt = ui.PrintPrompt()
	ti.PromptStyle = lipgloss.NewStyle()
	ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

	return model{
		textInput: ti,
		shell:     s,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			m.quitting = true
			return m, tea.Quit

		case tea.KeyUp:
			// History navigation
			if len(m.shell.history) > 0 && m.shell.histIdx < len(m.shell.history)-1 {
				m.shell.histIdx++
				m.textInput.SetValue(m.shell.history[len(m.shell.history)-1-m.shell.histIdx])
				m.textInput.CursorEnd()
			}
			return m, nil

		case tea.KeyDown:
			// History navigation
			if m.shell.histIdx > 0 {
				m.shell.histIdx--
				m.textInput.SetValue(m.shell.history[len(m.shell.history)-1-m.shell.histIdx])
				m.textInput.CursorEnd()
			} else if m.shell.histIdx == 0 {
				m.shell.histIdx = -1
				m.textInput.SetValue("")
			}
			return m, nil

		case tea.KeyTab:
			// Autocomplete
			input := m.textInput.Value()
			if strings.HasPrefix(input, "/") {
				completed := m.shell.autocomplete(input)
				if completed != input {
					m.textInput.SetValue(completed)
					m.textInput.CursorEnd()
				}
			}
			return m, nil

		case tea.KeyEnter:
			input := strings.TrimSpace(m.textInput.Value())
			if input == "" {
				return m, nil
			}

			// Add to history
			m.shell.history = append(m.shell.history, input)
			m.shell.histIdx = -1

			// Clear input
			m.textInput.SetValue("")

			// Process command
			fmt.Println() // New line after input
			if err := m.shell.processInput(input); err != nil {
				if err.Error() == "quit" {
					m.quitting = true
					return m, tea.Quit
				}
				ui.PrintError(err.Error())
			}
			fmt.Println() // Space before next prompt

			return m, nil
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.quitting {
		return ""
	}
	return m.textInput.View()
}

// Run starts the interactive shell
func (s *Shell) Run(version string) error {
	// Print banner
	ui.PrintBanner(version)

	// Print config info
	if s.cfg != nil {
		ui.PrintWelcome(s.cfg.Variant, s.cfg.ConfigDir)
	}

	// Start Bubble Tea program
	p := tea.NewProgram(initialModel(s))
	if _, err := p.Run(); err != nil {
		return err
	}

	fmt.Println(ui.MutedStyle.Render("Goodbye!"))
	return nil
}

// processInput handles a line of input
func (s *Shell) processInput(input string) error {
	if strings.HasPrefix(input, "/") {
		return s.executeCommand(input[1:])
	}

	// Treat bare input as a command without slash
	return s.executeCommand(input)
}

// executeCommand parses and executes a command
func (s *Shell) executeCommand(input string) error {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return nil
	}

	cmdName := strings.ToLower(parts[0])
	args := parts[1:]

	cmd, ok := s.commands[cmdName]
	if !ok {
		return fmt.Errorf("unknown command: %s (try /help)", cmdName)
	}

	return cmd.Handler(args)
}

// autocomplete provides tab completion for commands
func (s *Shell) autocomplete(input string) string {
	prefix := strings.TrimPrefix(input, "/")
	if prefix == "" {
		return input
	}

	var matches []string
	for name, cmd := range s.commands {
		if strings.HasPrefix(name, prefix) && name == cmd.Name {
			matches = append(matches, "/"+name)
		}
	}

	if len(matches) == 1 {
		return matches[0] + " "
	}

	return input
}

// Command handlers

func (s *Shell) cmdHelp(args []string) error {
	if len(args) > 0 {
		// Help for specific command
		cmd, ok := s.commands[args[0]]
		if !ok {
			return fmt.Errorf("unknown command: %s", args[0])
		}
		fmt.Println()
		fmt.Printf("  %s\n", ui.HelpKeyStyle.Render("/"+cmd.Name))
		fmt.Printf("  %s\n", cmd.Description)
		fmt.Println()
		fmt.Printf("  Usage:   %s\n", cmd.Usage)
		if len(cmd.Aliases) > 0 {
			fmt.Printf("  Aliases: %s\n", strings.Join(cmd.Aliases, ", "))
		}
		return nil
	}

	// Show interactive help menu
	selected, err := ShowHelpMenu(s.commands)
	if err != nil {
		return err
	}

	// If user selected a command, show its details
	if selected != "" {
		return s.cmdHelp([]string{selected})
	}

	return nil
}

func (s *Shell) cmdStatus(args []string) error {
	ui.PrintCommand("Checking container status...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := docker.NewClient(s.cfg.ProjectName)
	if err != nil {
		return err
	}
	defer client.Close()

	containers, err := client.ListContainers(ctx, true)
	if err != nil {
		return err
	}

	if len(containers) == 0 {
		ui.PrintInfo("No containers found")
		return nil
	}

	// Sort by name
	sort.Slice(containers, func(i, j int) bool {
		return containers[i].Name < containers[j].Name
	})

	fmt.Println()
	for _, c := range containers {
		var stateStyle lipgloss.Style
		switch c.State {
		case "running":
			stateStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981"))
		case "exited":
			stateStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444"))
		default:
			stateStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#F59E0B"))
		}

		health := ""
		if c.Health != "" {
			health = fmt.Sprintf(" [%s]", c.Health)
		}

		fmt.Printf("  %s  %s%s\n",
			stateStyle.Render(fmt.Sprintf("%-8s", c.State)),
			c.Name,
			ui.MutedStyle.Render(health))
	}

	// Summary
	running := 0
	for _, c := range containers {
		if c.State == "running" {
			running++
		}
	}
	fmt.Println()
	fmt.Printf("  %s\n", ui.MutedStyle.Render(fmt.Sprintf("%d/%d containers running", running, len(containers))))

	return nil
}

func (s *Shell) cmdDeploy(args []string) error {
	pull := false
	for _, arg := range args {
		if arg == "--pull" || arg == "-p" {
			pull = true
		}
	}

	ui.PrintCommand("Deploying media stack...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// Create directories
	ui.PrintInfo("Creating directories...")
	if err := stack.CreateDirectories(s.cfg.DataFolder, s.cfg.MediaFolder, s.cfg.PUID, s.cfg.PGID, false, false); err != nil {
		return err
	}

	// Copy config files
	ui.PrintInfo("Copying configuration files...")
	if err := stack.CopyConfigFiles(s.cfg.ConfigDir, s.cfg.DataFolder, s.cfg.PUID, s.cfg.PGID, false, false); err != nil {
		return err
	}

	compose := docker.NewCompose(s.cfg.ProjectName, s.cfg.ConfigDir, s.cfg.ComposeFile())

	// Pull if requested
	if pull {
		ui.PrintInfo("Pulling images...")
		if err := compose.Pull(ctx); err != nil {
			return err
		}
	}

	// Start services
	ui.PrintInfo("Starting services...")
	if err := compose.Up(ctx, true, false); err != nil {
		return err
	}

	ui.PrintSuccess("Deployment complete!")
	return nil
}

func (s *Shell) cmdStop(args []string) error {
	ui.PrintCommand("Stopping media stack...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	compose := docker.NewCompose(s.cfg.ProjectName, s.cfg.ConfigDir, s.cfg.ComposeFile())

	if len(args) > 0 {
		for _, service := range args {
			ui.PrintInfo(fmt.Sprintf("Stopping %s...", service))
			if err := compose.StopService(ctx, service); err != nil {
				return err
			}
		}
	} else {
		if err := compose.Down(ctx, false, true); err != nil {
			return err
		}
	}

	ui.PrintSuccess("Stopped!")
	return nil
}

func (s *Shell) cmdRestart(args []string) error {
	ui.PrintCommand("Restarting media stack...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	compose := docker.NewCompose(s.cfg.ProjectName, s.cfg.ConfigDir, s.cfg.ComposeFile())

	if len(args) > 0 {
		for _, service := range args {
			ui.PrintInfo(fmt.Sprintf("Restarting %s...", service))
			if err := compose.RestartService(ctx, service); err != nil {
				return err
			}
		}
	} else {
		if err := compose.Restart(ctx); err != nil {
			return err
		}
	}

	ui.PrintSuccess("Restarted!")
	return nil
}

func (s *Shell) cmdLogs(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: /logs <service>")
	}

	service := args[0]
	ui.PrintCommand(fmt.Sprintf("Showing logs for %s (Ctrl+C to stop)...", service))

	ctx := context.Background()
	compose := docker.NewCompose(s.cfg.ProjectName, s.cfg.ConfigDir, s.cfg.ComposeFile())

	return compose.Logs(ctx, service, true, "50", false)
}

func (s *Shell) cmdPull(args []string) error {
	ui.PrintCommand("Pulling images...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	compose := docker.NewCompose(s.cfg.ProjectName, s.cfg.ConfigDir, s.cfg.ComposeFile())

	if len(args) > 0 {
		for _, service := range args {
			ui.PrintInfo(fmt.Sprintf("Pulling %s...", service))
			if err := compose.PullService(ctx, service); err != nil {
				return err
			}
		}
	} else {
		if err := compose.Pull(ctx); err != nil {
			return err
		}
	}

	ui.PrintSuccess("Pull complete!")
	return nil
}

func (s *Shell) cmdValidate(args []string) error {
	ui.PrintCommand("Validating configuration...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Check config
	errors := s.cfg.Validate()
	if len(errors) > 0 {
		for _, err := range errors {
			ui.PrintError(err.Error())
		}
		return fmt.Errorf("validation failed")
	}

	// Check compose
	compose := docker.NewCompose(s.cfg.ProjectName, s.cfg.ConfigDir, s.cfg.ComposeFile())
	if err := compose.Config(ctx); err != nil {
		return err
	}

	ui.PrintSuccess("Configuration is valid!")
	return nil
}

func (s *Shell) cmdApikeys(args []string) error {
	ui.PrintCommand("Extracting API keys...")
	ui.PrintInfo("(Check the apikeys command output)")
	// This would need the full apikeys implementation
	return fmt.Errorf("use the CLI command: mediastack apikeys")
}

func (s *Shell) cmdConfig(args []string) error {
	fmt.Println(ui.TitleStyle.Render("Current Configuration"))
	fmt.Println()
	fmt.Printf("  %s  %s\n", ui.HelpKeyStyle.Render("Config Dir:"), s.cfg.ConfigDir)
	fmt.Printf("  %s  %s\n", ui.HelpKeyStyle.Render("Variant:   "), s.cfg.Variant)
	fmt.Printf("  %s  %s\n", ui.HelpKeyStyle.Render("Data:      "), s.cfg.DataFolder)
	fmt.Printf("  %s  %s\n", ui.HelpKeyStyle.Render("Media:     "), s.cfg.MediaFolder)
	fmt.Printf("  %s  %d:%d\n", ui.HelpKeyStyle.Render("UID:GID:   "), s.cfg.PUID, s.cfg.PGID)
	fmt.Printf("  %s  %s\n", ui.HelpKeyStyle.Render("Compose:   "), s.cfg.ComposeFile())
	return nil
}

func (s *Shell) cmdServices(args []string) error {
	ui.PrintCommand("Listing services...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	compose := docker.NewCompose(s.cfg.ProjectName, s.cfg.ConfigDir, s.cfg.ComposeFile())
	services, err := compose.ConfigServices(ctx)
	if err != nil {
		return err
	}

	fmt.Println()
	sort.Strings(services)
	for _, svc := range services {
		fmt.Printf("  • %s\n", svc)
	}
	fmt.Println()
	fmt.Printf("  %s\n", ui.MutedStyle.Render(fmt.Sprintf("%d services defined", len(services))))

	return nil
}

func (s *Shell) cmdExec(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: /exec <service> <command>")
	}

	service := args[0]
	command := args[1:]

	ui.PrintCommand(fmt.Sprintf("Executing in %s: %s", service, strings.Join(command, " ")))

	ctx := context.Background()
	compose := docker.NewCompose(s.cfg.ProjectName, s.cfg.ConfigDir, s.cfg.ComposeFile())

	return compose.Exec(ctx, service, command, true)
}

func (s *Shell) cmdClear(args []string) error {
	fmt.Print("\033[H\033[2J")
	return nil
}

func (s *Shell) cmdQuit(args []string) error {
	return fmt.Errorf("quit")
}
