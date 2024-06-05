package commands

import (
	"fmt"
	"strings"
)

var COMMANDS_HELP_DESC = map[string]string{
	"init": "Create an empty Git repository",
}

var FLAGS []string = []string{"-v", "-h"}

func Help() {
	var flagsStringBuilder strings.Builder

	for _, flag := range FLAGS {
		flagsStringBuilder.WriteString(fmt.Sprintf(" [%s]", flag))
	}

	flagsStringBuilder.WriteString(" <command> [<args>]")

	fmt.Printf("usage: got%s\n", flagsStringBuilder.String())

	fmt.Printf("\nThere are common Git commands implemented:\n\n")

	var commandsStringBuilder strings.Builder

	for cmd, desc := range COMMANDS_HELP_DESC {
		commandsStringBuilder.WriteString(fmt.Sprintf("\t%s\t%s\n", cmd, desc))
	}

	fmt.Println(commandsStringBuilder.String())

}
