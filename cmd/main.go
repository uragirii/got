package main

import (
	"flag"
	"fmt"

	"github.com/uragirii/got/cmd/commands"
	"github.com/uragirii/got/cmd/internals"
)

// TODO: use from git tags
var version string = "0.0.0-pre-alpha"

var SUPPORTED_COMMANDS []*internals.Command = []*internals.Command{
	commands.HASH_OBJECT,
	commands.INIT,
	commands.LS_FILES,
}

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

	var isValidCmd bool = false

	for _, cmdDetails := range SUPPORTED_COMMANDS {
		if cmdDetails.Name == command {
			cmdDetails.ParseCommand(args[1:])
			cmdDetails.Run(cmdDetails)
			isValidCmd = true
		}
	}

	if !isValidCmd {
		fmt.Printf("%s is not a supported argument\n", command)
		return
	}
}
