package pomo

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// create a enum for the session type
type SessionType string

const (
	WorkSession      SessionType = "work"
	BreakSession     SessionType = "break"
	LongBreakSession SessionType = "longbreak"
)

type Session struct {
	StartTime string
	EndTime   string
	Type      SessionType
	Filepath  string
}

const SESSION_FILENAME = "session.log"

func (s *Session) String() string {
	return fmt.Sprintf("type=%s start=%s end=%s | %s", s.Type, s.StartTime, s.EndTime, s.Filepath)
}

func (s *Session) Start() error {
	dir := conf.DirPath()
	if !Exists(dir) {
		return fmt.Errorf("could not resolve config path for %q", conf.Id)
	}

	homedir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	buffer, err := Read(filepath.Join(homedir, ".nvim-buf"))
	if err != nil {
		return err
	}

	s.Filepath = buffer

	sessionPath, err := sessionPath()
	if err != nil {
		return err
	}

	return InsertLine(sessionPath, s.String())
}

func (s *Session) Current() error {
	sessionPath, err := sessionPath()
	if err != nil {
		return err
	}

	lines, err := ReadLines(sessionPath)
	if err != nil {
		return err
	}

	if len(lines) == 0 {
		return nil
	}

	lastLine := lines[len(lines)-1]
	// Split by "|"
	parts := strings.SplitN(lastLine, " | ", 2)
	if len(parts) < 2 {
		return fmt.Errorf("session log format error: '|' separator not found")
	}

	// Split session details
	details := strings.Fields(parts[0])
	for _, detail := range details {
		keyValue := strings.SplitN(detail, "=", 2)
		if len(keyValue) != 2 {
			return fmt.Errorf("session detail format error: '=' separator not found in %s", detail)
		}
		key, value := keyValue[0], keyValue[1]
		switch key {
		case "type":
			s.Type = SessionType(value)
		case "start":
			s.StartTime = value
		case "end":
			s.EndTime = value
		}
	}
	s.Filepath = parts[1]

	return nil
}

func (s *Session) Save() error {
	sessionPath, err := sessionPath()
	if err != nil {
		return err
	}

	lines, err := ReadLines(sessionPath)
	if err != nil {
		return err
	}

	found := false

	for i, line := range lines {
		start := fmt.Sprintf("start=%s", s.StartTime)
		if strings.Contains(line, start) {
			lines[i] = s.String()
			found = true
			break
		}
	}

	if !found {
		return nil
	}

	return Write(sessionPath, strings.Join(lines, "\n"))
}

func (s *Session) Stop() error {
	s.EndTime = time.Now().Format(time.RFC3339)
	return s.Save()
}

func (s *Session) Remove() error {
	sessionPath, err := sessionPath()
	if err != nil {
		return err
	}

	lines, err := ReadLines(sessionPath)
	if err != nil {
		return err
	}

	if len(lines) == 0 {
		return nil
	}

	for i, line := range lines {
		start := fmt.Sprintf("start=%s", s.StartTime)
		if strings.Contains(line, start) {
			lines = append(lines[:i], lines[i+1:]...)
			break
		}
	}

	return Write(sessionPath, strings.Join(lines, "\n"))
}

func (s *Session) isRunning() bool {
	return s.EndTime == ""
}

func sessionPath() (string, error) {
	dir := conf.DirPath()
	if !Exists(dir) {
		return "", fmt.Errorf("could not resolve config path for %q", conf.Id)
	}

	path := filepath.Join(dir, SESSION_FILENAME)

	if !Exists(path) {
		fmt.Println("Creating session file...")
		_, err := os.Create(path)
		if err != nil {
			return "", err
		}
	}

	return path, nil
}

func ListSessions() ([]Session, error) {
	sessionPath, err := sessionPath()
	if err != nil {
		return nil, err
	}

	lines, err := ReadLines(sessionPath)
	if err != nil {
		return nil, err
	}

	sessions := make([]Session, 0, len(lines))
	for _, line := range lines {
		var session Session
		_, err := fmt.Sscanf(line, "type=%s start=%s end=%s | %s", &session.Type, &session.StartTime, &session.EndTime, &session.Filepath)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}
