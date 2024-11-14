package pomo

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	prefix   string
	quit     bool
	notified bool
	width    int
	height   int
	session  Session
}

var (
	timerStyle = lipgloss.NewStyle().
			Bold(true).
			Padding(1, 4).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Align(lipgloss.Center)

	prefixStyle = lipgloss.NewStyle().
			Bold(true).
			Padding(0, 1).
			Align(lipgloss.Center)

	containerStyle = lipgloss.NewStyle().
			Align(lipgloss.Center).
			Padding(2)

	quitStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Align(lipgloss.Center)
)

// Init implements tea.Model
func (m model) Init() tea.Cmd {
	return tick()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			// !! do not stop the session
			// if m.session.isRunning() {
			// 	if err := m.session.Stop(); err != nil {
			// 		fmt.Printf("Failed to stop session: %v\n", err)
			// 	}
			// }
			m.quit = true
			return m, tea.Quit

		case "w": // Switch to work
			duration := conf.QueryString("duration")
			if duration == "" {
				duration = Break
			}
			dur, err := time.ParseDuration(duration)
			if err != nil {
				fmt.Printf("Failed to parse duration: %v\n", err)
			}

			// Stop current session
			if m.session.isRunning() {
				if err := m.session.Stop(); err != nil {
					fmt.Printf("Failed to stop session: %v\n", err)
				}
			}

			var session Session
			if err := session.Start(conf, dur, WorkSession); err != nil {
				fmt.Printf("Failed to start work session: %v\n", err)
			}

			if err := conf.Set("prefix", WorkPrefix); err != nil {
				fmt.Printf("Failed to set the prefix config %v\n", err)
			}

			m.session = session
			m.notified = false

			return m, tick()

		case "b": // Switch to break
			duration := conf.QueryString("break")
			if duration == "" {
				duration = Break
			}
			dur, err := time.ParseDuration(duration)
			if err != nil {
				fmt.Printf("Failed to parse duration: %v\n", err)
			}

			// Stop current session
			if m.session.isRunning() {
				if err := m.session.Stop(); err != nil {
					fmt.Printf("Failed to stop session: %v\n", err)
				}
			}

			var session Session
			if err := session.Start(conf, dur, BreakSession); err != nil {
				fmt.Printf("Failed to start break session: %v\n", err)
			}

			if err := conf.Set("prefix", BreakPrefix); err != nil {
				fmt.Printf("Failed to set the prefix config %v\n", err)
			}

			m.session = session
			m.notified = false

			return m, tick()

		case "r": // Reset current timer
			if err := m.session.Reset(); err != nil {
				fmt.Printf("Failed to stop session: %v\n", err)
			}
			m.notified = false
		}
		return m, tick()

	case time.Time:
		remaining := m.session.Elapsed()
		if !m.notified && remaining <= 0 && m.session.isRunning() {
			title := "Pomo Timer"
			message := "Time to take a break!"
			if m.session.Type == BreakSession {
				message = "Break is over! Time to focus!"
			}

			go func() {
				if err := sendNotification(title, message); err != nil {
					fmt.Printf("Failed to send notification: %v\n", err)
				}
				if err := playAlert(); err != nil {
					fmt.Printf("Failed to play alert: %v\n", err)
				}
			}()
			m.notified = true
		}
		return m, tick()
	}
	return m, nil
}

func (m model) View() string {
	if m.quit {
		return quitStyle.Render("Goodbye!") + "\n"
	}

	// Set colors based on session type and remaining time
	var timeStyle = timerStyle
	remaining := m.session.Elapsed()

	if remaining < 0 {
		timeStyle = timeStyle.Foreground(lipgloss.Color("196")) // Red
	} else if m.session.Type == WorkSession {
		timeStyle = timeStyle.Foreground(lipgloss.Color("35")) // Green
	} else {
		timeStyle = timeStyle.Foreground(lipgloss.Color("39")) // Blue
	}

	// Format the duration display
	duration := remaining
	sign := ""
	if duration < 0 {
		duration = -duration
		sign = "-"
	}

	remainingStr := sign + StopWatchFormat(duration)

	// ... (rest of the View formatting remains the same)
	contentWidth := m.width / 6
	if contentWidth < 30 {
		contentWidth = 30 // minimum width
	}

	timeStyle = timeStyle.Width(contentWidth)
	prefixStyle = prefixStyle.Width(contentWidth)
	containerStyle = containerStyle.Width(m.width)

	var sb strings.Builder

	verticalPad := (m.height - 9) / 2
	if verticalPad > 0 {
		sb.WriteString(strings.Repeat("\n", verticalPad))
	}

	prefixText := prefixStyle.Render(m.prefix)
	sb.WriteString(prefixText + "\n\n")

	timerText := timeStyle.Render(remainingStr)
	sb.WriteString(timerText + "\n\n")

	helpStyle := quitStyle
	sb.WriteString(helpStyle.Render("w: work • b: break • r: reset • q: quit") + "\n")

	return containerStyle.Render(sb.String())
}

func StartUI() error {
	var session Session
	if err := session.Get(); err != nil {
		return fmt.Errorf("failed to get current session: %v", err)
	}

	prefix := conf.QueryString("prefix")
	if prefix == "" {
		prefix = WorkPrefix
	}

	initialModel := model{
		prefix:  prefix,
		session: session,
	}

	p := tea.NewProgram(
		initialModel,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to start UI: %v", err)
	}

	return nil
}
