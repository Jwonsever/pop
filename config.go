package pop

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/pop/logging"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// ErrConfigFileNotFound is returned when the pop config file can't be found,
// after looking for it.
var ErrConfigFileNotFound = errors.New("unable to find pop config file")

var lookupPaths = []string{"", "./config", "config"}

// ConfigName is the name of the YAML databases config file
var ConfigName = "database.yml"

func init() {
	SetLogger(defaultLogger)

	ap := os.Getenv("APP_PATH")
	if ap != "" {
		_ = AddLookupPaths(ap)
	}
	ap = os.Getenv("POP_PATH")
	if ap != "" {
		_ = AddLookupPaths(ap)
	}
}

// LoadConfigFile loads a POP config file from the configured lookup paths
func LoadConfigFile() error {
	path, err := findConfigPath()
	if err != nil {
		return err
	}
	Connections = map[string]*Connection{}
	log(logging.Debug, "Loading config file from %s", path)
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	return LoadFrom(f)
}

// LookupPaths returns the current configuration lookup paths
func LookupPaths() []string {
	return lookupPaths
}

// AddLookupPaths add paths to the current lookup paths list
func AddLookupPaths(paths ...string) error {
	lookupPaths = append(paths, lookupPaths...)
	return nil
}

func findConfigPath() (string, error) {
	for pathPrepend := ""; ; {
		for _, p := range LookupPaths() {
			path, _ := filepath.Abs(filepath.Join(pathPrepend, p, ConfigName))
			log(logging.Debug, "Looking for config at path %s", path)

			// If our path is outside of any goPath, stop looking.
			if !inGoPath(path) {
				return "", ErrConfigFileNotFound
			}

			if _, err := os.Stat(path); err == nil {
				return path, err
			}
		}
		pathPrepend = fmt.Sprintf("%s%s%c", pathPrepend, "..", filepath.Separator)
	}
}

func inGoPath(s string) bool {
	inGoPath := false
	for _, p := range envy.GoPaths() {
		if strings.HasPrefix(s, p) {
			inGoPath = true
		}
	}

	return inGoPath
}

// LoadFrom reads a configuration from the reader and sets up the connections
func LoadFrom(r io.Reader) error {
	envy.Load()
	deets, err := ParseConfig(r)
	if err != nil {
		return err
	}
	for n, d := range deets {
		con, err := NewConnection(d)
		if err != nil {
			log(logging.Warn, "unable to load connection %s: %v", n, err)
			continue
		}
		Connections[n] = con
	}
	return nil
}

// ParseConfig reads the pop config from the given io.Reader and returns
// the parsed ConnectionDetails map.
func ParseConfig(r io.Reader) (map[string]*ConnectionDetails, error) {
	tmpl := template.New("test")
	tmpl.Funcs(map[string]interface{}{
		"envOr": func(s1, s2 string) string {
			return envy.Get(s1, s2)
		},
		"env": func(s1 string) string {
			return envy.Get(s1, "")
		},
	})
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	t, err := tmpl.Parse(string(b))
	if err != nil {
		return nil, errors.Wrap(err, "couldn't parse config template")
	}

	var bb bytes.Buffer
	err = t.Execute(&bb, nil)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't execute config template")
	}

	deets := map[string]*ConnectionDetails{}
	err = yaml.Unmarshal(bb.Bytes(), &deets)
	return deets, errors.Wrap(err, "couldn't unmarshal config to yaml")
}
