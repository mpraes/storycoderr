package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const DefaultYAML = `version: 1

repository:
  include:
    - "**/*.py"
    - "tests/**/*.py"
    - "docs/**/*.md"
  exclude:
    - ".git/**"
    - ".venv/**"
    - "venv/**"
    - "__pycache__/**"
    - "node_modules/**"

analysis:
  languages:
    - python
  follow_symlinks: false
  max_file_size_bytes: 5242880

storage:
  mode: repository
  engine: sqlite
`

type Settings struct {
	Version          int
	Include          []string
	Exclude          []string
	Languages        []string
	FollowSymlinks   bool
	MaxFileSizeBytes int64
}

// Defaults returns the built-in StoryCode analysis settings.
//
//	s := config.Defaults()
func Defaults() Settings {
	s, err := Parse(DefaultYAML)
	if err != nil {
		return Settings{
			Version:          1,
			Include:          []string{"**/*.py"},
			MaxFileSizeBytes: 5242880,
		}
	}
	return s
}

// LoadFile reads settings from path, or Defaults when the file is missing.
//
//	s, err := config.LoadFile("/repo/.storycode/config.yaml")
func LoadFile(path string) (Settings, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Defaults(), nil
		}
		return Settings{}, fmt.Errorf("cannot read config %q: %w (expected a readable YAML file)", path, err)
	}
	return Parse(string(body))
}

func Parse(text string) (Settings, error) {
	var s Settings
	var list *[]string
	for i, raw := range strings.Split(text, "\n") {
		if err := applyConfigLine(&s, &list, strings.TrimSpace(strings.TrimSuffix(raw, "\r")), i+1); err != nil {
			return Settings{}, err
		}
	}
	return validateSettings(s)
}

func applyConfigLine(s *Settings, list **[]string, line string, n int) error {
	if line == "" || strings.HasPrefix(line, "#") {
		return nil
	}
	if strings.HasPrefix(line, "- ") {
		return appendList(list, unquote(strings.TrimSpace(line[2:])), n)
	}
	key, value, ok := strings.Cut(line, ":")
	if !ok {
		return fmt.Errorf("cannot parse config line %d %q, expected key: value or list item", n, line)
	}
	return applyConfigKey(s, list, strings.TrimSpace(key), strings.TrimSpace(value), n)
}

func appendList(list **[]string, item string, n int) error {
	if *list == nil {
		return fmt.Errorf("cannot parse list item %q on line %d, expected it under include, exclude, or languages", item, n)
	}
	**list = append(**list, item)
	return nil
}

func applyConfigKey(s *Settings, list **[]string, key, value string, n int) error {
	*list = nil
	switch key {
	case "include":
		s.Include = nil
		*list = &s.Include
	case "exclude":
		s.Exclude = nil
		*list = &s.Exclude
	case "languages":
		s.Languages = nil
		*list = &s.Languages
	case "version":
		return setVersion(s, value, n)
	case "follow_symlinks":
		return setFollow(s, value, n)
	case "max_file_size_bytes":
		return setMaxSize(s, value, n)
	}
	return nil
}

func setVersion(s *Settings, value string, n int) error {
	v, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("cannot parse version %q on line %d, expected an integer", value, n)
	}
	s.Version = v
	return nil
}

func setFollow(s *Settings, value string, n int) error {
	v, err := strconv.ParseBool(value)
	if err != nil {
		return fmt.Errorf("cannot parse follow_symlinks %q on line %d, expected true or false", value, n)
	}
	s.FollowSymlinks = v
	return nil
}

func setMaxSize(s *Settings, value string, n int) error {
	v, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fmt.Errorf("cannot parse max_file_size_bytes %q on line %d, expected an integer byte count", value, n)
	}
	s.MaxFileSizeBytes = v
	return nil
}

func unquote(value string) string {
	if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
		return value[1 : len(value)-1]
	}
	return value
}

func validateSettings(s Settings) (Settings, error) {
	if s.Version == 0 {
		s.Version = 1
	}
	if len(s.Include) == 0 {
		return Settings{}, fmt.Errorf("config include %v is empty, expected at least one glob such as **/*.py", s.Include)
	}
	if s.MaxFileSizeBytes <= 0 {
		s.MaxFileSizeBytes = 5242880
	}
	return s, nil
}
