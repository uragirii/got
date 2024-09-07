package commit

import (
	"testing"
	"time"
)

const TEST_AUTHOR_LINE = "author Apoorv Kansal <apoorvkansalak@gmail.com> 1720643686 +0530"
const TEST_COMMITTER_LINE = "committer Apoorv Kansal <apoorvkansalak@gmail.com> 1720643686 +0530"

const TIME_STR = "Thu, 11 Jul 2024 02:04:46 +0530"
const TIME_FORMATTED = "1720643686 +0530"

func TestParseAuthorLine(t *testing.T) {
	correctTime, _ := time.Parse(time.RFC1123Z, TIME_STR)
	authorPerson, authorTime := parseAuthorLine(TEST_AUTHOR_LINE)

	if authorPerson.email != "apoorvkansalak@gmail.com" {
		t.Errorf("got wrong email address %s", authorPerson.email)
	}

	if authorPerson.name != "Apoorv Kansal" {
		t.Errorf("got wrong name %s", authorPerson.name)
	}

	if correctTime.Format(time.RFC1123Z) != authorTime.Format(time.RFC1123Z) {
		t.Errorf("got wrong time expected: %s got :%s", correctTime.Format(time.RFC1123Z), authorTime.Format(time.RFC1123Z))
	}

}

func TestParseCommitterLine(t *testing.T) {
	correctTime, _ := time.Parse(time.RFC1123Z, TIME_STR)

	authorPerson, authorTime := parseCommitterLine(TEST_COMMITTER_LINE)

	if authorPerson.email != "apoorvkansalak@gmail.com" {
		t.Errorf("got wrong email address %s", authorPerson.email)
	}

	if authorPerson.name != "Apoorv Kansal" {
		t.Errorf("got wrong name %s", authorPerson.name)
	}

	if correctTime.Format(time.RFC1123Z) != authorTime.Format(time.RFC1123Z) {
		t.Errorf("got wrong time expected: %s got :%s", correctTime.Format(time.RFC1123Z), authorTime.Format(time.RFC1123Z))

	}

}

func TestToGitTime(t *testing.T) {
	ti, _ := time.Parse(time.RFC1123Z, TIME_STR)

	fmtTime := toGitTime(ti)

	if fmtTime != TIME_FORMATTED {
		t.Errorf("got wrong time expected: %s got :%s", TIME_FORMATTED, fmtTime)

	}

}
