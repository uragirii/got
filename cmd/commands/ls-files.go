package commands

import (
	"fmt"
	"path"

	"github.com/uragirii/got/cmd/internals"
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

func LsFiles(c *internals.Command, rootPath string) {

	if rootPath == "" {
		fmt.Println("fatal: not a git repository (or any of the parent directories): .git")
		return
	}

	// TODO: get root dir of git
	indexFiles, err := internals.ParseIndex(path.Join(rootPath, "index"))
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, file := range indexFiles {
		if c.GetFlag("modified") == "true" {
			currSha, _, err := internals.HashBlob(file.Filepath, false)

			if err != nil {
				panic(err)
			}

			if string((*currSha)[:]) != string((*file.SHA1)[:]) {
				fmt.Println(file.Filepath)
			}
		} else {
			fmt.Println(file.Filepath)
		}
	}
}
