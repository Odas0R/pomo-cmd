package pomo

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// create a enum for the session type
type SessionType string

const (
	WorkSession      SessionType = "work"
	BreakSession     SessionType = "break"
	LongBreakSession SessionType = "longbreak"
)

type Session struct {
	ID        uuid.UUID
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
	Type      SessionType
	File      string
}

const SESSION_FILENAME = "session.log"

func (s *Session) Elapsed() time.Duration {
	if !s.isRunning() {
		return 0
	}
	targetEnd := s.StartTime.Add(s.Duration)
	return targetEnd.Sub(time.Now())
}

func (s *Session) Start(conf Conf, dur time.Duration, mode SessionType) error {
	s.ID = uuid.New()
	s.StartTime = time.Now()
	s.Duration = dur
	s.EndTime = time.Time{} // empty time
	s.Type = mode

	dir := conf.DirPath()
	if !Exists(dir) {
		return fmt.Errorf("could not resolve config path for %q", conf.Id)
	}

	homedir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	currentFile, err := Read(filepath.Join(homedir, ".nvim-buf"))
	if err != nil {
		return err
	}

	s.File = currentFile

	sessionPath, err := sessionPath()
	if err != nil {
		return err
	}

	if err := InsertLine(sessionPath, s.String()); err != nil {
		return err
	}

	return nil
}

func (s *Session) Get() error {
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

	if err := s.Scan(lastLine); err != nil {
		return err
	}

	return nil
}

func (s *Session) String() string {
	return fmt.Sprintf(
		"id=%s type=%s start=%s end=%s duration=%s | %s",
		s.ID,
		s.Type,
		s.StartTime.Format(time.RFC3339),
		s.EndTime.Format(time.RFC3339),
		s.Duration,
		s.File,
	)
}

func (s *Session) Scan(line string) error {
	parts := strings.SplitN(line, " | ", 2)
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
		case "id":
			s.ID = uuid.MustParse(value)
		case "type":
			s.Type = SessionType(value)
		case "duration":
			dur, err := time.ParseDuration(value)
			if err != nil {
				return err
			}
			s.Duration = dur
		case "start":
			t, err := time.Parse(time.RFC3339, value)
			if err != nil {
				return err
			}
			s.StartTime = t
		case "end":
			if value == "" {
				s.EndTime = time.Time{}
			} else {
				t, err := time.Parse(time.RFC3339, value)
				if err != nil {
					return err
				}
				s.EndTime = t
			}
		}
	}
	s.File = parts[1]

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
		start := fmt.Sprintf("id=%s", s.ID)
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
	s.EndTime = time.Now()
	return s.Save()
}

func (s *Session) Reset() error {
	s.StartTime = time.Now()
	s.EndTime = time.Time{}
	return s.Save()
}

func (s *Session) Delete() error {
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
		start := fmt.Sprintf("id=%s", s.ID)
		if strings.Contains(line, start) {
			lines = append(lines[:i], lines[i+1:]...)
			break
		}
	}

	return Write(sessionPath, strings.Join(lines, "\n"))
}

func (s *Session) isRunning() bool {
	return s.EndTime.IsZero()
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
		session := Session{}

		if err := session.Scan(line); err != nil {
			return nil, err
		}

		sessions = append(sessions, session)
	}

	return sessions, nil
}
