package commit

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

type person struct {
	name  string
	email string
}

func parseOffset(offset string) int {
	sign := 1

	if offset[0] == '-' {
		sign = -1
	}

	hours, _ := strconv.Atoi(offset[1:3])
	minutes, _ := strconv.Atoi(offset[3:])

	return sign * (hours*60*60 + minutes*60)
}

func parseTime(timeLine string) time.Time {
	authorTimeSplitted := strings.Split(timeLine, " ")

	unixTimestamp, err := strconv.Atoi(authorTimeSplitted[0])

	if err != nil {
		panic(err)
	}

	offset := parseOffset(authorTimeSplitted[1])

	location := time.FixedZone(authorTimeSplitted[1], offset)

	return time.UnixMilli(int64(unixTimestamp * 1000)).In(location)

}

func parseAuthorLine(authorLine string) (person, time.Time) {
	// remove "author" from start
	authorLine = authorLine[7:]

	// I'm assuming email is always set
	emailStartIdx := strings.IndexByte(authorLine, '<')
	emailEndIdx := strings.IndexByte(authorLine, '>')

	name := authorLine[:emailStartIdx-1]
	email := authorLine[emailStartIdx+1 : emailEndIdx]

	authorTime := parseTime(authorLine[emailEndIdx+2:])

	return person{
		name, email,
	}, authorTime
}

func parseCommitterLine(commiterLine string) (person, time.Time) {
	// remove "committer" from start
	commiterLine = commiterLine[10:]

	// I'm assuming email is always set
	emailStartIdx := strings.IndexByte(commiterLine, '<')
	emailEndIdx := strings.IndexByte(commiterLine, '>')

	name := commiterLine[:emailStartIdx-1]
	email := commiterLine[emailStartIdx+1 : emailEndIdx]

	authorTime := parseTime(commiterLine[emailEndIdx+2:])

	return person{
		name, email,
	}, authorTime
}

func toGitTime(t time.Time) string {
	_, offset := t.Zone()

	offsetMin := offset / 60

	hrs := int(math.Abs(float64(offsetMin / 60)))
	min := int(math.Abs(float64(offsetMin % 60)))

	sign := "+"

	if offsetMin < 0 {
		sign = "-"
	}

	return fmt.Sprintf("%d %s%02d%02d", t.Unix(), sign, hrs, min)

}
