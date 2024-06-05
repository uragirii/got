package main

import (
	"fmt"
	"os"
	"slices"

	"github.com/uragirii/got/cmd/commands"
)

var COMMAND_TO_FUNC = map[string]func(){
	"init": commands.Init,
}

var SUPPORTED_COMMANDS = []string{"init"}

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		fmt.Println("no arguments were provided")
		return
	}

	command := args[0]

	isValidCmd := slices.Contains(SUPPORTED_COMMANDS, command)

	if !isValidCmd {
		fmt.Printf("%s is not a supported argument\n", command)
		return
	}

	commandFunc := COMMAND_TO_FUNC[command]

	commandFunc()

}
