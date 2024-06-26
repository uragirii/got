package cmd

import (
	"fmt"

	"github.com/uragirii/got/internals"
	"github.com/uragirii/got/internals/git"
	"github.com/uragirii/got/internals/git/object"
)

var LS_FILES *internals.Command = &internals.Command{
	Name: "ls-files",
	Desc: "TBD",
	Flags: []*internals.Flag{
		{
			Name:  "cached",
			Short: "c",
			Key:   "cached",
			Help:  "show cached files in the output (default)",
			Type:  internals.Bool,
		},

		{
			Name:  "modified",
			Short: "m",
			Key:   "modified",
			Help:  "show modified files in the output",
			Type:  internals.Bool,
		},
	},
	Run: LsFiles,
}

func LsFiles(c *internals.Command, _ string) {

	gitIndex, err := git.UnmarshallGitIndex()

	if err != nil {
		panic(err)
	}

	for _, entry := range gitIndex.GetTrackedFiles() {
		if c.GetFlag("modified") == "true" {
			obj, err := object.NewGitObject(entry.Filepath)

			if err != nil {
				panic(err)
			}

			if !obj.GetSHA().Eq(entry.SHA) {
				fmt.Println(entry.Filepath)
			}
		} else {
			fmt.Println(entry.Filepath)
		}
	}
}
