package cmd

import (
	"fmt"
	"os"

	"github.com/uragirii/got/internals"
	"github.com/uragirii/got/internals/git/index"
)

var ADD *internals.Command = &internals.Command{
	Name:  "add",
	Desc:  "Add file contents to the index",
	Flags: []*internals.Flag{},
	Run:   Add,
}

func Add(c *internals.Command, gitPath string) {

	if len(c.Args) == 0 {
		fmt.Println("Nothing specified, nothing added.")
		return
	}

	gitDir, err := internals.GetGitDir()

	if err != nil {
		panic(err)
	}

	gitFs := os.DirFS(gitDir)

	indexFile, err := gitFs.Open(index.IndexFileName)

	if err != nil {
		panic(err)
	}

	index, err := index.New(indexFile)

	if err != nil {
		panic(err)
	}

	index.Add(c.Args, os.DirFS("."))
	index.WriteToFile()

}
