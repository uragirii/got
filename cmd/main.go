package main

import (
	"flag"
	"fmt"
	"slices"

	"github.com/uragirii/got/cmd/commands"
)

// TODO: use from git tags
var version string = "0.0.0-pre-alpha"

var COMMAND_TO_FUNC = map[string]func(){
	"init": commands.Init,
	"help": commands.Help,
}

var SUPPORTED_COMMANDS = []string{"init"}

func main() {

	isVersion := flag.Bool("v", false, "version")
	isHelp := flag.Bool("h", false, "help")

	flag.Parse()

	if *isVersion {
		fmt.Printf("got version %s\n", version)
		return
	}

	if *isHelp {
		commands.Help()
		return
	}

	args := flag.Args()

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
