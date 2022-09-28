package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	_fs "io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/odas0r/pomo-cmd/pkg/dir"
	"github.com/odas0r/pomo-cmd/pkg/editor"
	"github.com/odas0r/pomo-cmd/pkg/fs"
	"github.com/rogpeppe/go-internal/lockedfile"
	gojsonq "github.com/thedevsaddam/gojsonq/v2"
)

type Conf struct {
	Id   string // usually application name
	Dir  string // usually os.UserConfigDir
	File string // usually config.yaml
}

func (c Conf) DirPath() string { return filepath.Join(c.Dir, c.Id) }

func (c Conf) Path() string { return filepath.Join(c.Dir, c.Id, c.File) }

func (c Conf) Init() error {
	d := c.DirPath()

	if d == "" {
		return fmt.Errorf("could not resolve config path for %q", c.Id)
	}
	if len(c.Id) == 0 && len(c.Dir) == 0 {
		return fmt.Errorf("empty directory id")
	}

	if fs.Exists(d) {
		if err := os.RemoveAll(d); err != nil {
			return err
		}
	}

	if err := dir.Create(d); err != nil {
		return err
	}

	return fs.Cat("{}", c.Path())
}

func (c Conf) Data() string {
	buf, err := os.ReadFile(c.Path())
	if err != nil {
		log.Println(err)
		return ""
	}
	return string(buf)
}

func (c Conf) Set(key string, val any) error {
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(c.Data()), &config); err != nil {
		return err
	}

	// append key to json
	config[key] = val

	if err := c.OverWrite(config); err != nil {
		return err
	}

	return nil
}

func (c Conf) Del(key string) error {
	var config map[string]json.RawMessage
	if err := json.Unmarshal([]byte(c.Data()), &config); err != nil {
		return err
	}

	if _, exists := config[key]; exists {
		delete(config, key)
		if err := c.OverWrite(config); err != nil {
			return err
		}
	}

	return nil
}

func (c Conf) Print() error {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, []byte(c.Data()), "", "    "); err != nil {
		return err
	}

	fmt.Println(prettyJSON.String())

	return nil
}

func (c Conf) Edit() error {
	if err := c.mkdir(); err != nil {
		return err
	}
	path := c.Path()
	if path == "" {
		return fmt.Errorf("unable to locate config for %q", c.Id)
	}
	return editor.Edit(path)
}

func (c Conf) OverWrite(newconf any) error {
	buf, err := json.Marshal(newconf)
	if err != nil {
		return err
	}
	if err := c.mkdir(); err != nil {
		return err
	}
	return lockedfile.Write(c.Path(),
		bytes.NewReader(buf), _fs.FileMode(fs.DefaultPerms))
}

// Using github.com/thedevsaddam/gojsonq to query json files.
//
// Wiki: https://github.com/thedevsaddam/gojsonq/wiki/Queries
func (c Conf) Query(q string) string {
	result := gojsonq.New().File(c.Path()).Find(q)

	if result == nil {
		return ""
	}

	return fmt.Sprintf("%v", result)
}

// QueryPrint prints the output of Query.
func (c Conf) QueryPrint(q string) { fmt.Print(c.Query(q)) }

//////////////////////////////////////////////////////
// Private methods
//////////////////////////////////////////////////////

func (c Conf) mkdir() error {
	d := c.DirPath()
	if d == "" {
		return fmt.Errorf("failed to find config for %q", c.Id)
	}
	if fs.NotExists(d) {
		if err := dir.Create(d); err != nil {
			return err
		}
	}
	return nil
}
