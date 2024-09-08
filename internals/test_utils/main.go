package testutils

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/uragirii/got/internals/color"
)

const CHARS_PER_LINE = 16

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

		incorrectLine := ""

		if i < len(incorrectSplitted) {
			incorrectLine = incorrectSplitted[i]

		}

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

func splitAfterN(b *[]byte, n int) [][]byte {
	ret := make([][]byte, 0, len(*b)/n)

	for i := 0; i < len(*b); {
		item := make([]byte, 0, n)

		for j := range n {
			if i+j >= len(*b) {
				ret = append(ret, item)
				return ret
			}

			item = append(item, (*b)[i+j])

		}

		ret = append(ret, item)

		i += n
	}

	return ret
}

func hexdumpC(input *[]byte) string {
	var formattedStr strings.Builder

	splitted := splitAfterN(input, CHARS_PER_LINE)

	for _, itemBuff := range splitted {

		for _, item := range itemBuff {
			formattedStr.WriteString(fmt.Sprintf("%02x ", item))
		}

		if len(itemBuff) < CHARS_PER_LINE {
			formattedStr.WriteString(strings.Repeat(" ", (CHARS_PER_LINE-len(itemBuff))*3))
		}

		formattedStr.WriteString(" | ")

		for _, item := range itemBuff {
			if unicode.IsPrint(rune(item)) {

				formattedStr.WriteString(string(item))

			} else {
				formattedStr.WriteString(".")
			}
		}

		formattedStr.WriteString("\n")

	}

	return formattedStr.String()
}

func DiffBytes(correct, incorrect *[]byte) string {
	correctStr := hexdumpC(correct)
	incorrectStr := hexdumpC(incorrect)

	return Diff(correctStr, incorrectStr)
}
