package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/infosecstreams/subsleuth/utils"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model is our main application model.
type Model struct {
	Tabs       []string
	TabContent []string
	activeTab  int
	table      table.Model
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		// Check for mouse events
		case "mouse":
			// Scroll depending on the mouse wheel event
			m.table, cmd = m.table.Update(msg)
			return m, cmd

		case "ctrl+c", "q":
			return m, tea.Quit

		case "right", "l", "n", "tab":
			// Tab should wrap around
			m.activeTab = (m.activeTab + 1) % len(m.Tabs)
			return m, nil

		case "left", "h", "p", "shift+tab":
			// Tab should wrap around
			m.activeTab = (m.activeTab - 1 + len(m.Tabs)) % len(m.Tabs)
			return m, nil

		// TODO(gpsy): Implement add dialog
		case "a":
			if m.activeTab == 1 {
				return m, tea.Batch(
					tea.Printf("You pressed a!"),
				)
			}

		// TODO(gpsy): Implement delete dialog
		case "d":
			// TODO(gpsy): Resolve username to id from cache
			if m.activeTab == 1 {
				return m, tea.Batch(
					tea.Printf("You pressed d with broadcaster id %s!", m.table.SelectedRow()[0]),
				)
			}
		case "enter":
			if m.activeTab == 1 {
				return m, tea.Batch(
					tea.Printf("You selected webhook with broadcaster id %s!", m.table.SelectedRow()[0]),
				)
			}
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// Styles
var (
	inactiveTabBorder = tabBorderWithBottom("┴", "─", "┴")
	activeTabBorder   = tabBorderWithBottom("┘", " ", "└")
	docStyle          = lipgloss.NewStyle().Padding(1, 2, 1, 2)
	highlightColor    = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	inactiveTabStyle  = lipgloss.NewStyle().Border(inactiveTabBorder, true).BorderForeground(highlightColor).Padding(0, 1)
	activeTabStyle    = inactiveTabStyle.Copy().Border(activeTabBorder, true)
	windowStyle       = lipgloss.NewStyle().BorderForeground(highlightColor).Padding(2, 0).Align(lipgloss.Center).Border(lipgloss.NormalBorder()).UnsetBorderTop()
	baseStyle         = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240"))
)

// Helper functions
func tabBorderWithBottom(left, middle, right string) lipgloss.Border {
	border := lipgloss.RoundedBorder()
	border.BottomLeft = left
	border.Bottom = middle
	border.BottomRight = right
	return border
}

// Render the view
func (m Model) View() string {
	doc := strings.Builder{}

	var renderedTabs []string

	for i, t := range m.Tabs {
		var style lipgloss.Style
		isFirst, isLast, isActive := i == 0, i == len(m.Tabs)-1, i == m.activeTab
		if isActive {
			style = activeTabStyle.Copy()
		} else {
			style = inactiveTabStyle.Copy()
		}
		border, _, _, _, _ := style.GetBorder()
		if isFirst && isActive {
			border.BottomLeft = "│"
		} else if isFirst && !isActive {
			border.BottomLeft = "├"
		} else if isLast && isActive {
			border.BottomRight = "│"
		} else if isLast && !isActive {
			border.BottomRight = "┤"
		}
		style = style.Border(border)
		renderedTabs = append(renderedTabs, style.Render(t))
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	doc.WriteString(row)
	doc.WriteString("\n")

	switch m.activeTab {
	case 0:
		doc.WriteString(windowStyle.Width((lipgloss.Width(row) - windowStyle.GetHorizontalFrameSize())).Render(m.TabContent[m.activeTab]))
	case 1:
		doc.WriteString(baseStyle.Render(m.table.View()))
	}

	return docStyle.Render(doc.String())
}

func main() {
	// Log all to file in cache dir
	logdir, _ := os.UserCacheDir()
	logdir = filepath.Join(logdir, "subsleuth")
	logPath := filepath.Join(logdir, "subsleuth.log")
	f, err := tea.LogToFile(logPath, "subsleuth:")
	if err != nil {
		log.Fatalf("could not create log file: %v", err)
	}
	log.Printf("Starting subsleuth at %s", time.Now().Format("2006-01-02 15:04:05"))
	defer f.Close()
	// Start timing
	start := time.Now()
	// Load the cache
	cache := utils.NewCache()
	cache.GetCache()
	cache.LoadUsers()
	log.Printf("Users loaded: %d\n", len(cache.Users.Users))
	end := time.Now()
	log.Printf("Time to load cache: %v\n", end.Sub(start))
	// log.Fatal("test")

	// Get windows size * 2/3
	// _, height := utils.GetWindowSize()
	var numEntries int
	flag.IntVar(&numEntries, "n", 15, "number of entries to display at once")
	flag.Parse()

	eventSubsLists, err := loadEventSubsLists(cache)
	log.Printf("EventSubs loaded: %d\n", len(eventSubsLists.Subscription))
	if err != nil {
		log.Fatalf("Failed to load event subscriptions: %v", err)
	}

	columns := []table.Column{
		{Title: "User", Width: 12},
		{Title: "Status", Width: 10},
		{Title: "Online", Width: 7},
		{Title: "Offline", Width: 7},
		{Title: "URL", Width: 30},
		{Title: "Created", Width: 30},
	}

	rows := make([]table.Row, 0, len(eventSubsLists.Subscription))
	broadcasterRows := make(map[string]*table.Row)

	for _, data := range eventSubsLists.Subscription {
		streamer := cache.GetUsernames(data.Condition.BroadcasterUserID)
		online := "x"
		offline := "x"
		if data.Type == "stream.online" {
			online = "✓"
		} else if data.Type == "stream.offline" {
			offline = "✓"
		}

		row, exists := broadcasterRows[streamer]
		if exists {
			// if the broadcaster already exists,
			// we update the online/offline status
			// in its corresponding row
			if data.Type == "stream.online" {
				(*row)[2] = "✓" // Online field
			} else if data.Type == "stream.offline" {
				(*row)[3] = "✓" // Offline field
			}
		} else {
			// else, add a new row
			newRow := []string{
				streamer,
				data.Status,
				online,
				offline,
				data.Transport.Callback,
				data.CreatedAt.Local().Format("2006-01-02 15:04"),
			}
			rows = append(rows, newRow)
			broadcasterRows[streamer] = &rows[len(rows)-1]
			// Sort the rows by username
			sort.Slice(rows, func(i, j int) bool {
				return rows[i][0] < rows[j][0]
			})
		}
	}

	_, height := utils.GetWindowSize()
	height = int(height * 1 / 3)
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(height),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	tabs := []string{"SubSleuth", "Subscriptions"}
	tabContent := []string{
		"Hello!",
		"This program displays a table of Twitch EventSubs.\nYou can add a new one with 'a' or select one to delete and press 'd'.",
	}
	m := Model{
		Tabs:       tabs,
		TabContent: tabContent,
		table:      t,
	}
	if _, err := tea.NewProgram(m).Run(); err != nil {
		log.Fatalf("error running program: %v", err)
	}
}

func loadEventSubsLists(c utils.Cache) (utils.EventSubsLists, error) {
	var eventSubsLists utils.EventSubsLists
	data, err := os.ReadFile(c.CacheFilePath)
	if err != nil {
		return eventSubsLists, fmt.Errorf("failed to read file: %w", err)
	}

	err = json.Unmarshal(data, &eventSubsLists)
	if err != nil {
		return eventSubsLists, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return eventSubsLists, nil
}
