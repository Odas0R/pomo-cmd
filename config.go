package pomo

import (
	"bytes"
	"encoding/json"
	"fmt"
	_fs "io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/rogpeppe/go-internal/lockedfile"
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

	if err := Mkdir(d); err != nil {
		return err
	}

	if !Exists(c.Path()) {
		return WriteAppend("{}", c.Path())
	}

	return nil
}

func (c Conf) Data() []byte {
	buf, err := os.ReadFile(c.Path())
	if err != nil {
		log.Println(err)
		return []byte("{}")
	}
	return buf
}

func (c Conf) Set(key string, val any) error {
	var config map[string]interface{}
	if err := json.Unmarshal(c.Data(), &config); err != nil {
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
	if err := json.Unmarshal(c.Data(), &config); err != nil {
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

func (c Conf) Query(key string) any {
	var config map[string]interface{}
	if err := json.Unmarshal(c.Data(), &config); err != nil {
		return nil
	}

	val, exists := config[key]
	if !exists {
		return nil
	}

	return val
}

func (c Conf) QueryString(key string) string {
	result := c.Query(key)
	if val, ok := result.(string); !ok {
		log.Fatalf("QueryString: %s key value is not a string", key)
	} else {
		return val
	}
	return ""
}

func (c Conf) Print() error {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, c.Data(), "", "    "); err != nil {
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
	return Editor(path)
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
		bytes.NewReader(buf), _fs.FileMode(DefaultPerms))
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
	return Mkdir(d)
}
