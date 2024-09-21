package config

import (
	"io"
	"os"
	"path"
	"strings"
)

type User struct {
	Name  string
	Email string
}

type Config struct {
	User User
}

const _GLOBAL_CONFIG_PATH = ".gitconfig"

func userSectionParser(sectionLines []string, config *Config) error {

	for _, line := range sectionLines {
		line = strings.Trim(line, "\t")

		splittedLine := strings.Split(line, " = ")

		if len(splittedLine) != 2 {
			break
		}

		key := splittedLine[0]
		value := splittedLine[1]

		switch key {
		case "name":
			config.User.Name = value
		case "email":
			config.User.Email = value
		}
	}
	return nil
}

var _SECTION_PARSER_MAP = map[string]func(sectionLines []string, c *Config) error{
	"user": userSectionParser,
}

func New(reader io.Reader) (*Config, error) {
	configFileBytes, err := io.ReadAll(reader)

	if err != nil {
		return nil, err
	}

	// TODO: i want to use bufio here
	// but the config files are really small imo
	// plus idc about perf rn, we can improve later
	configFile := string(configFileBytes)

	config := &Config{}

	lines := strings.Split(configFile, "\n")

	for idx := range len(lines) {
		line := lines[idx]

		if line == "" {
			break
		}

		if strings.HasPrefix(line, "[") {
			// section header
			sectionName := line[1 : len(line)-1]

			sectionParser, hasParser := _SECTION_PARSER_MAP[sectionName]

			if hasParser {

				var section []string

				idx++
				for ; idx < len(lines)-1 && lines[idx][0] != '['; idx++ {
					section = append(section, lines[idx])
				}

				// need to do it coz range also autoincrements
				idx--

				err = sectionParser(section, config)

				if err != nil {
					return config, err
				}

			}
		}

	}

	return config, nil
}

func FromFile() (*Config, error) {
	dirName, err := os.UserHomeDir()

	if err != nil {
		return nil, err
	}

	configPath := path.Join(dirName, _GLOBAL_CONFIG_PATH)

	configFile, err := os.Open(configPath)

	if err != nil {
		return nil, err
	}

	return New(configFile)

}
