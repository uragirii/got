package cmd

import (
	"fmt"
	"os"

	"github.com/uragirii/got/internals"
	"github.com/uragirii/got/internals/git/blob"
	"github.com/uragirii/got/internals/git/index"
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

	gitIndex, err := index.New()

	if err != nil {
		panic(err)
	}

	for _, entry := range gitIndex.GetTrackedFiles() {
		if c.GetFlag("modified") == "true" {
			file, err := os.Open(entry.Filepath)

			if err != nil {
				panic(err)
			}

			obj, err := blob.FromFile(file)

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
