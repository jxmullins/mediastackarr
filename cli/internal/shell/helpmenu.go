package shell

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	helpTitleStyle = lipgloss.NewStyle().
			MarginLeft(2).
			Foreground(lipgloss.Color("#7C3AED")).
			Bold(true)

	helpItemStyle = lipgloss.NewStyle().
			PaddingLeft(4)

	helpSelectedStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(lipgloss.Color("#7C3AED")).
				Bold(true)

	helpDescStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280"))

	helpCategoryStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(lipgloss.Color("#10B981")).
				Bold(true)
)

// helpItem represents a command in the help list
type helpItem struct {
	name        string
	description string
	category    string
	isCategory  bool
}

func (i helpItem) Title() string {
	if i.isCategory {
		return helpCategoryStyle.Render(i.name)
	}
	return "/" + i.name
}

func (i helpItem) Description() string {
	if i.isCategory {
		return ""
	}
	return i.description
}

func (i helpItem) FilterValue() string {
	return i.name
}

// helpDelegate handles rendering of list items
type helpDelegate struct{}

func (d helpDelegate) Height() int { return 1 }

func (d helpDelegate) Spacing() int { return 0 }

func (d helpDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d helpDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(helpItem)
	if !ok {
		return
	}

	if i.isCategory {
		fmt.Fprintf(w, "\n%s\n", helpCategoryStyle.Render(i.name))
		return
	}

	str := fmt.Sprintf("  /%-12s %s", i.name, helpDescStyle.Render(i.description))

	if index == m.Index() {
		str = helpSelectedStyle.Render("> ") + fmt.Sprintf("%-12s %s", "/"+i.name, i.description)
	}

	fmt.Fprint(w, str)
}

// helpModel is the Bubble Tea model for the help menu
type helpModel struct {
	list     list.Model
	quitting bool
	selected string
}

func newHelpModel(commands map[string]*Command) helpModel {
	// Define categories and their commands
	categories := []struct {
		Name     string
		Commands []string
	}{
		{"Stack Management", []string{"deploy", "stop", "restart", "pull"}},
		{"Monitoring", []string{"status", "logs", "services"}},
		{"Configuration", []string{"config", "validate", "apikeys"}},
		{"Shell", []string{"exec", "clear", "help", "quit"}},
	}

	var items []list.Item

	for _, cat := range categories {
		// Add category header
		items = append(items, helpItem{
			name:       cat.Name,
			isCategory: true,
		})

		// Add commands in this category
		for _, name := range cat.Commands {
			if cmd, ok := commands[name]; ok {
				items = append(items, helpItem{
					name:        cmd.Name,
					description: cmd.Description,
					category:    cat.Name,
				})
			}
		}
	}

	// Create list with custom delegate
	l := list.New(items, helpDelegate{}, 50, 20)
	l.Title = "Commands"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(true)
	l.Styles.Title = helpTitleStyle
	l.SetShowPagination(false)

	return helpModel{list: l}
}

func (m helpModel) Init() tea.Cmd {
	return nil
}

func (m helpModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			if item, ok := m.list.SelectedItem().(helpItem); ok && !item.isCategory {
				m.selected = item.name
				m.quitting = true
				return m, tea.Quit
			}
		}
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		// Fixed height to avoid large gap at bottom
		m.list.SetHeight(24)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m helpModel) View() string {
	if m.quitting {
		return ""
	}
	return "\n" + m.list.View()
}

// ShowHelpMenu displays the interactive help menu
func ShowHelpMenu(commands map[string]*Command) (string, error) {
	m := newHelpModel(commands)
	p := tea.NewProgram(m)

	finalModel, err := p.Run()
	if err != nil {
		return "", err
	}

	if fm, ok := finalModel.(helpModel); ok {
		return fm.selected, nil
	}
	return "", nil
}
