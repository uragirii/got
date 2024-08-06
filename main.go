package main

import (
	"flag"
	"fmt"
	"path/filepath"

	"github.com/uragirii/got/cmd"
	"github.com/uragirii/got/internals"
)

// TODO: use from git tags
var version string = "0.0.0-pre-alpha"

var SUPPORTED_COMMANDS []*internals.Command = []*internals.Command{
	cmd.HASH_OBJECT,
	cmd.INIT,
	cmd.LS_FILES,
	cmd.STATUS,
	cmd.CAT_FILE,
	cmd.ADD,
	cmd.COMMIT,
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
		cmd.Help()
		return
	}

	args := flag.Args()

	if len(args) == 0 {
		cmd.Help()
		return
	}

	gitDir, err := internals.GetGitDir()

	if err != nil {
		panic(err)
	}

	root := filepath.Join(gitDir, "..")

	command := args[0]

	var isValidCmd bool = false

	for _, cmdDetails := range SUPPORTED_COMMANDS {
		if cmdDetails.Name == command {
			cmdDetails.ParseCommand(args[1:])
			cmdDetails.Run(cmdDetails, root)
			isValidCmd = true
		}
	}

	if !isValidCmd {
		fmt.Printf("%s is not a supported argument\n", command)
		return
	}
}
