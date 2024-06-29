package git

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type ignoreEntry struct {
	rule    string
	rootDir string
}

func (entry *ignoreEntry) Match(filePath string) bool {
	rel, err := filepath.Rel(entry.rootDir, filePath)

	if err != nil {
		// We dont care
		return false
	}

	matched, err := path.Match(entry.rule, rel)

	if err != nil {
		return false
	}

	if matched {
		return true
	}

	matched, err = path.Match(fmt.Sprintf("%s/", entry.rule), rel)

	if err != nil {
		return false
	}

	return matched
}

type Ignore struct {
	rules []ignoreEntry
}

func NewIgnore(filePath string) (*Ignore, error) {

	var ignore *Ignore

	return ignore.WithFile(filePath)
}

func (g *Ignore) Match(filePath string) bool {
	for _, rule := range g.rules {
		if rule.Match(filePath) {
			return true
		}
	}

	return false
}

func (g *Ignore) WithFile(ignoreFilePath string) (*Ignore, error) {
	contents, err := os.ReadFile(ignoreFilePath)

	if err != nil {
		return nil, err
	}

	splitedLines := strings.Split(string(contents), "\n")

	rules := make([]ignoreEntry, len(g.rules)+len(splitedLines))

	rootDir := path.Join(ignoreFilePath, "..")

	for idx, line := range splitedLines {
		if line != "" {
			rules[idx] = ignoreEntry{
				rule:    line,
				rootDir: rootDir,
			}
		}
	}

	return &Ignore{
		rules: rules,
	}, nil
}
