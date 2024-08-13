package testutils

import (
	"fmt"
	"strings"

	"github.com/uragirii/got/internals/color"
)

func diffLine(correctLine, incorrectLine string) string {
	var sb strings.Builder

	for i, correctRune := range correctLine {
		incorrectRune := incorrectLine[i]

		if correctRune == rune(incorrectRune) {
			sb.WriteRune(correctRune)
		} else {
			sb.WriteString(color.RedString(fmt.Sprintf("%q", correctRune)))
			sb.WriteString(color.GreenString(fmt.Sprintf("%q", incorrectRune)))
		}
	}

	return sb.String()
}

func Diff(correct, incorrect string) string {
	// Very simple, split the str1 and str2 and then color the line which dont match
	// TODO: Char by char and not line by line
	// TODO: what if no of lines are not the same

	var sb strings.Builder

	correctSplitted := strings.Split(correct, "\n")
	incorrectSplitted := strings.Split(incorrect, "\n")

	for i, correctLine := range correctSplitted {
		incorrectLine := incorrectSplitted[i]

		if incorrectLine == correctLine {
			sb.WriteString(correctLine)
			sb.WriteRune('\n')
		} else {
			sb.WriteString(color.RedString(correctLine))
			sb.WriteRune('\n')
			sb.WriteString(color.GreenString(incorrectLine))
			sb.WriteRune('\n')
			// sb.WriteRune('|')
			// sb.WriteString(diffLine(correctLine, incorrectLine))
			// sb.WriteRune('|')
			// sb.WriteRune('\n')

		}
	}

	return sb.String()
}

func DiffBytes(correct, incorrect *[]byte) string {
	var correctStr strings.Builder
	var incorrectStr strings.Builder

	for i, r := range *correct {

		correctStr.WriteString(fmt.Sprintf("% x", r))

		if (i+1)%10 == 0 {
			correctStr.WriteRune('\n')
		}
	}

	for i, r := range *incorrect {

		incorrectStr.WriteString(fmt.Sprintf("% x", r))

		if (i+1)%10 == 0 {
			incorrectStr.WriteRune('\n')
		}
	}

	return Diff(correctStr.String(), incorrectStr.String())
}
