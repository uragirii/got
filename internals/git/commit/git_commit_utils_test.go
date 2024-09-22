package commit

import (
	"fmt"
	"testing"
	"time"

	testutils "github.com/uragirii/got/internals/test_utils"
)

const TEST_USER_NAME = "Apoorv Kansal"
const TEST_USER_EMAIL = "dont_doxx_me@idc.com"

var TEST_AUTHOR_LINE = fmt.Sprintf("author %s <%s> 1720643686 +0530", TEST_USER_NAME, TEST_USER_EMAIL)
var TEST_COMMITTER_LINE = fmt.Sprintf("committer %s <%s> 1720643686 +0530", TEST_USER_NAME, TEST_USER_EMAIL)

const TIME_STR = "Thu, 11 Jul 2024 02:04:46 +0530"
const TIME_FORMATTED = "1720643686 +0530"

func TestParseAuthorLine(t *testing.T) {
	correctTime, _ := time.Parse(time.RFC1123Z, TIME_STR)
	authorPerson, authorTime := parseAuthorLine(TEST_AUTHOR_LINE)

	testutils.AssertString(t, "email", TEST_USER_EMAIL, authorPerson.Email)
	testutils.AssertString(t, "name", TEST_USER_NAME, authorPerson.Name)

	if correctTime.Format(time.RFC1123Z) != authorTime.Format(time.RFC1123Z) {
		t.Errorf("got wrong time expected: %s got :%s", correctTime.Format(time.RFC1123Z), authorTime.Format(time.RFC1123Z))
	}

}

func TestParseCommitterLine(t *testing.T) {
	correctTime, _ := time.Parse(time.RFC1123Z, TIME_STR)

	authorPerson, authorTime := parseCommitterLine(TEST_COMMITTER_LINE)

	testutils.AssertString(t, "email", TEST_USER_EMAIL, authorPerson.Email)
	testutils.AssertString(t, "name", TEST_USER_NAME, authorPerson.Name)

	if authorPerson.Email != TEST_USER_EMAIL {
		t.Errorf("got wrong email address %s", authorPerson.Email)
	}

	if correctTime.Format(time.RFC1123Z) != authorTime.Format(time.RFC1123Z) {
		t.Errorf("got wrong time expected: %s got :%s", correctTime.Format(time.RFC1123Z), authorTime.Format(time.RFC1123Z))

	}

}

func TestToGitTime(t *testing.T) {
	ti, _ := time.Parse(time.RFC1123Z, TIME_STR)

	fmtTime := toGitTime(ti)

	testutils.AssertString(t, "time", TIME_FORMATTED, fmtTime)

}
