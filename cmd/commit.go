package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/uragirii/got/internals"
	"github.com/uragirii/got/internals/git/commit"
	"github.com/uragirii/got/internals/git/head"
	"github.com/uragirii/got/internals/git/index"
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

func Commit(cmd *internals.Command, gitPath string) {

	if cmd.GetFlag("message") == "" {
		fmt.Println("Aborting commit due to empty commit message")
		return
	}

	message := cmd.GetFlag("message")

	gitDir, err := internals.GetGitDir()

	if err != nil {
		panic(err)
	}

	indexFile, err := os.Open(path.Join(gitDir, "index"))

	if err != nil {
		panic(err)
	}

	i, err := index.New(indexFile)

	if err != nil {
		panic(err)
	}

	err = i.Hydrate()

	if err != nil {
		panic(err)
	}

	err = i.WriteToFile()

	if err != nil {
		panic(err)
	}

	gitFs := os.DirFS(gitDir)

	c, err := commit.New(gitFs, message)

	if err != nil {
		panic(err)
	}

	h, err := head.New(gitFs)

	if err != nil {
		panic(err)
	}

	h.SetTo(c.GetSHA(), h.Mode)

	err = h.WriteToFile()

	if err != nil {
		panic(err)
	}

	err = c.WriteToFile()

	if err != nil {
		panic(err)
	}

	fmt.Printf("[%s %s] %s\n", h.Branch, c.GetSHA().String()[:7], message)
}
