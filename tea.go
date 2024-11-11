package pomo

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// tick generates a tick every second
func tick() tea.Cmd {
	return tea.Every(time.Second, func(t time.Time) tea.Msg {
		return t
	})
}
