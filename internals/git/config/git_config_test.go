package config_test

import (
	"fmt"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/uragirii/got/internals/git/config"
	testutils "github.com/uragirii/got/internals/test_utils"
)

const TEST_USER_NAME = "Apoorv Kansal"
const TEST_USER_EMAIL = "dont_doxx_me@idc.com"

var TEST_CONFIG_FILE = fmt.Sprintf(`[credential "https://github.com"]
	helper = 
	helper = !/opt/homebrew/bin/gh auth git-credential
[credential "https://gist.github.com"]
	helper = 
	helper = !/opt/homebrew/bin/gh auth git-credential
[user]
	name = %s
	email = %s
`, TEST_USER_NAME, TEST_USER_EMAIL)

func assertConfig(c *config.Config, t *testing.T) {
	t.Helper()

	testutils.AssertString(t, "name", TEST_USER_NAME, c.User.Name)
	testutils.AssertString(t, "email", TEST_USER_EMAIL, c.User.Email)

}

func TestNewConfig(t *testing.T) {
	c, err := config.New(strings.NewReader(TEST_CONFIG_FILE))

	if err != nil {
		t.Errorf("Failed with err %v", err)
	}

	assertConfig(c, t)
}

func TestFromFile(t *testing.T) {
	tempDir := t.TempDir()

	randomName := fmt.Sprintf("%d", time.Now().UnixMicro())

	randomConfigFile := path.Join(tempDir, randomName)

	err := os.WriteFile(randomConfigFile, []byte(TEST_CONFIG_FILE), 0755)

	if err != nil {
		t.Errorf("Failed to create temp file %v", err)
	}

	t.Setenv("GIT_CONFIG", randomConfigFile)

	c, err := config.FromFile()

	if err != nil {
		t.Errorf("Failed with err %v", err)
	}
	assertConfig(c, t)

}
