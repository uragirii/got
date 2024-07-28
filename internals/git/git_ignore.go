package git

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
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

	matched, err = path.Match(fmt.Sprintf("%s/*", entry.rule), rel)

	if err != nil {
		return false
	}

	return matched
}

type Ignore struct {
	rules []ignoreEntry
}

func NewIgnore(filePath string, fsys fs.FS) (*Ignore, error) {
	var rules []ignoreEntry

	ignore := Ignore{
		rules: rules,
	}

	return ignore.WithFile(filePath, fsys)
}

func (g *Ignore) Match(filePath string) bool {
	for _, rule := range g.rules {
		if rule.Match(filePath) {
			return true
		}
	}

	return false
}

func (g *Ignore) WithFile(ignoreFilePath string, fsys fs.FS) (*Ignore, error) {
	// If file doesn't exist return empty ignore list
	ignoreFile, err := fsys.Open(ignoreFilePath)

	if errors.Is(err, fs.ErrNotExist) {
		return g, nil
	}

	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(ignoreFile)

	// Just random number
	var rules []ignoreEntry
	rootDir := path.Join(ignoreFilePath, "..")

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "#") {
			continue
		}

		if line != "" {
			line, _ := strings.CutPrefix(line, "/")
			rules = append(rules, ignoreEntry{
				rule:    line,
				rootDir: rootDir,
			})

		}
	}

	return &Ignore{
		rules: rules,
	}, nil
}
