package pomo

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Session struct {
	Type      string
	StartTime time.Time
	EndTime   time.Time
	Filepath  string
}

const SESSION_FILENAME = "session.log"

func (s *Session) Write() error {
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

	session := fmt.Sprintf("type=%s start=%s end=%s | %s", s.Type, s.StartTime, s.EndTime, buffer)

	return Write(session, filepath.Join(dir, SESSION_FILENAME))
}

func ListSessions() ([]Session, error) {
	dir := conf.DirPath()
	if !Exists(dir) {
		return nil, fmt.Errorf("could not resolve config path for %q", conf.Id)
	}

	lines, err := ReadLines(filepath.Join(dir, SESSION_FILENAME))
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
