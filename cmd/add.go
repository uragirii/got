package cmd

import (
	"github.com/uragirii/got/internals"
	"github.com/uragirii/got/internals/git"
)

var ADD *internals.Command = &internals.Command{
	Name:  "add",
	Desc:  "Add file contents to the index",
	Flags: []*internals.Flag{},
	Run:   Add,
}

func Add(c *internals.Command, gitPath string) {

	_, err := git.NewIndex()

	if err != nil {
		panic(err)
	}

}
