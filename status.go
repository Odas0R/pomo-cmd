package pomo

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type statusModel struct {
	workDuration   time.Duration
	breakDuration  time.Duration
	workGoal       time.Duration
	restGoal       time.Duration
	workPercentage float64
	restPercentage float64
	quit           bool
	width          int
	height         int
}

var (
	sectionStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#3C3C3C")).
			Padding(1, 2).
			Align(lipgloss.Center)

	workStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#98C379")). // Green
			Align(lipgloss.Center)

	breakStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#61AFEF")). // Blue
			Align(lipgloss.Center)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			Align(lipgloss.Center)

	progressBarWidth = 30
)

func renderProgressBar(percentage float64, width int) string {
	filled := int(percentage * float64(width) / 100)
	if filled > width {
		filled = width
	}

	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "━"
		} else {
			bar += "─"
		}
	}

	return fmt.Sprintf("[%s] %.1f%%", bar, percentage)
}

func (m statusModel) Init() tea.Cmd {
	return tick()
}

func (m statusModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.quit = true
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case time.Time:
		// Update statistics
		sessions, err := ListSessions()
		if err != nil {
			return m, tick()
		}

		todaySessions, err := filterTodaySessions(sessions)
		if err != nil {
			return m, tick()
		}

		typeDurations, _ := summarizeSessions(todaySessions)

		m.workDuration = typeDurations[WorkSession]
		m.breakDuration = typeDurations[BreakSession]
		m.workPercentage = float64(m.workDuration) / float64(WorkGoal) * 100
		m.restPercentage = float64(m.breakDuration) / float64(RestGoal) * 100

		return m, tick()
	}
	return m, nil
}

func (m statusModel) View() string {
	if m.quit {
		return ""
	}

	// Calculate vertical padding
	verticalPad := (m.height - 8) / 2
	if verticalPad < 0 {
		verticalPad = 0
	}

	doc := strings.Builder{}

	if verticalPad > 0 {
		doc.WriteString(strings.Repeat("\n", verticalPad))
	}

	containerStyle := lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center)

	// Sessions Section
	sessionsContent := lipgloss.JoinVertical(lipgloss.Center,
		"Today's Sessions",
		workStyle.Render(fmt.Sprintf("Work:  %s", formatDurationHm(m.workDuration))),
		breakStyle.Render(fmt.Sprintf("Break: %s", formatDurationHm(m.breakDuration))),
	)

	// Goals Section
	workProgress := renderProgressBar(m.workPercentage, progressBarWidth)
	restProgress := renderProgressBar(m.restPercentage, progressBarWidth)

	goalsContent := lipgloss.JoinVertical(lipgloss.Center,
		"Daily Goals",
		workStyle.Render(fmt.Sprintf("Work:  %s", formatDurationHm(m.workGoal))),
		workStyle.Render(workProgress),
		breakStyle.Render(fmt.Sprintf("Break: %s", formatDurationHm(m.restGoal))),
		breakStyle.Render(restProgress),
	)

	sections := lipgloss.JoinHorizontal(
		lipgloss.Center,
		sectionStyle.Render(sessionsContent),
		"    ",
		sectionStyle.Render(goalsContent),
	)

	doc.WriteString(containerStyle.Render(sections))
	doc.WriteString("\n\n")
	doc.WriteString(containerStyle.Render(helpStyle.Render("Press 'q' to quit")))

	return doc.String()
}

func ShowStatus() error {
	sessions, err := ListSessions()
	if err != nil {
		return err
	}

	todaySessions, err := filterTodaySessions(sessions)
	if err != nil {
		return err
	}

	typeDurations, _ := summarizeSessions(todaySessions)

	workDuration := typeDurations[WorkSession]
	breakDuration := typeDurations[BreakSession]

	model := statusModel{
		workDuration:   workDuration,
		breakDuration:  breakDuration,
		workGoal:       WorkGoal,
		restGoal:       RestGoal,
		workPercentage: float64(workDuration) / float64(WorkGoal) * 100,
		restPercentage: float64(breakDuration) / float64(RestGoal) * 100,
	}

	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running status view: %v", err)
	}

	return nil
}

func filterTodaySessions(sessions []Session) ([]Session, error) {
	var todaySessions []Session
	now := time.Now()
	today := now.Format("2006-01-02") // YYYY-MM-DD format

	for _, session := range sessions {
		if session.StartTime.Format("2006-01-02") == today {
			todaySessions = append(todaySessions, session)
		}
	}

	return todaySessions, nil
}

func summarizeSessions(sessions []Session) (map[SessionType]time.Duration, map[string]time.Duration) {
	typeDurations := make(map[SessionType]time.Duration)
	projectDurations := make(map[string]time.Duration)

	for _, session := range sessions {
		var duration time.Duration
		if !session.EndTime.IsZero() {
			duration = session.EndTime.Sub(session.StartTime)
		} else {
			duration = time.Since(session.StartTime)
		}

		typeDurations[session.Type] += duration
		projectDurations[session.File] += duration
	}

	return typeDurations, projectDurations
}

func formatDurationHm(d time.Duration) string {
	hours := d / time.Hour
	minutes := (d % time.Hour) / time.Minute

	if hours == 0 {
		return fmt.Sprintf("%02dm", minutes)
	}

	return fmt.Sprintf("%dh:%02dm", hours, minutes)
}
