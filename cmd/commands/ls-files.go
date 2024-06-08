package commands

import (
	"fmt"

	"github.com/uragirii/got/cmd/internals"
)

var LS_FILES *internals.Command = &internals.Command{
	Name:  "ls-files",
	Desc:  "TBD",
	Flags: []*internals.Flag{},
	Run:   LsFiles,
}

func LsFiles(c *internals.Command) {
	_, err := internals.ParseIndex(".git/index")
	if err != nil {
		fmt.Println(err)
	}
}
