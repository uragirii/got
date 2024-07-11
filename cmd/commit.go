package cmd

import (
	"fmt"

	"github.com/uragirii/got/internals"
)

var COMMIT *internals.Command = &internals.Command{
	Name: "commit",
	Desc: "Record changes to the repository",
	Flags: []*internals.Flag{
		{
			Name:  "message",
			Short: "m",
			Help:  "commit message",
			Key:   "message",
			Type:  internals.String,
		},
	},
	Run: Commit,
}

func Commit(c *internals.Command, gitPath string) {

	if c.GetFlag("message") == "" {
		fmt.Println("Aborting commit due to empty commit message")
		return
	}

}
