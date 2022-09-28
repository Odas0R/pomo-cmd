package fs

import (
	"errors"
	"io/fs"
	"os"
)

const (
	DefaultPerms = 0600
)

func Touch(path string) error {
	myfile, err := os.Create(path)
	if err != nil {
		return err
	}

	if err := myfile.Close(); err != nil {
		return err
	}

	return nil
}

func Cat(text string, path string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	if _, err := f.WriteString(text); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}

	return nil
}

func Exists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, fs.ErrNotExist)
}

func NotExists(path string) bool {
	_, err := os.Stat(path)
	return errors.Is(err, fs.ErrNotExist)
}
