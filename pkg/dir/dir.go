package dir

import (
	"os"
)

func Create(path string) error {
	err := os.Mkdir(path, 0755)
	if err != nil {
		return err
	}

	return nil
}
