package config_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/uragirii/got/internals/git/config"
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

func TestNewConfig(t *testing.T) {
	c, err := config.New(strings.NewReader(TEST_CONFIG_FILE))

	if err != nil {
		t.Errorf("Failed with err %v", err)
	}

	if c.User.Name != TEST_USER_NAME {
		t.Errorf("Expected name to be `%s` but got `%s`", TEST_USER_NAME, c.User.Name)
	}
	if c.User.Email != TEST_USER_EMAIL {
		t.Errorf("Expected email to be `%s` but got `%s`", TEST_USER_EMAIL, c.User.Email)
	}
}
