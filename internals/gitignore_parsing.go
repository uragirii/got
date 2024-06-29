package internals

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// Deprecated: use git.Ignore instead
type IgnoreEntry struct {
	rule    string
	rootDir string
}

// Deprecated: use git.Ignore instead
func (entry *IgnoreEntry) Match(filePath string) bool {
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

// Deprecated: use git.Ignore instead
type GitIgnore struct {
	rules []*IgnoreEntry
}

// Deprecated: use git.Ignore instead
func (g *GitIgnore) New(ignoreFilePath string, rootDir string) error {
	contents, err := os.ReadFile(ignoreFilePath)

	if err != nil {
		return err
	}

	splitedLines := strings.Split(string(contents), "\n")

	rules := make([]*IgnoreEntry, len(splitedLines))

	for idx, line := range splitedLines {
		if line != "" {
			rules[idx] = &IgnoreEntry{
				rule:    line,
				rootDir: rootDir,
			}
		}
	}

	g.rules = rules

	return nil
}

// Deprecated: use git.Ignore instead
func (g *GitIgnore) WithFile(ignoreFilePath string, rootDir string) (*GitIgnore, error) {
	contents, err := os.ReadFile(ignoreFilePath)

	if err != nil {
		return nil, err
	}

	splitedLines := strings.Split(string(contents), "\n")

	rules := make([]*IgnoreEntry, len(g.rules)+len(splitedLines))

	for idx, line := range splitedLines {
		if line != "" {
			rules[idx] = &IgnoreEntry{
				rule:    line,
				rootDir: rootDir,
			}
		}
	}

	return &GitIgnore{
		rules: rules,
	}, nil
}

// Deprecated: use git.Ignore instead
func (g *GitIgnore) Match(filePath string) bool {
	for _, rule := range g.rules {
		if rule.Match(filePath) {
			return true
		}
	}

	return false
}
