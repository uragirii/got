package cmd

import (
	"fmt"

	"github.com/uragirii/got/internals"
	"github.com/uragirii/got/internals/git"
	"github.com/uragirii/got/internals/git/object"
)

var STATUS *internals.Command = &internals.Command{
	Name:  "status",
	Desc:  "Show the working tree status",
	Flags: []*internals.Flag{},
	Run:   Status,
}

func Status(c *internals.Command, gitPath string) {

	head, err := git.NewHead()

	if err != nil {
		panic(err)
	}

	obj, err := object.NewObjectFromSHA(head.SHA)

	if err != nil {
		panic(err)
	}

	commit, err := object.ToCommit(obj)

	if err != nil {
		panic(err)
	}

	tree, err := object.NewTree()

	if err != nil {
		panic(err)
	}

	changes, err := commit.Tree.Compare(tree)

	if err != nil {
		panic(err)
	}

	fmt.Println(changes)

	// filesChan := make(chan *internals.FileStatus, 10)

	// rootDir := path.Join(gitPath, "..")

	// currDirTree := internals.GetTreeHash(rootDir)

	// var indexFile internals.GitIndex

	// err = indexFile.New(gitPath)

	// if err != nil {
	// 	panic(err)
	// }

	// go func() {
	// 	internals.CompareTree(commit.Tree, currDirTree, &indexFile, filesChan)
	// 	close(filesChan)
	// }()

	// for file := range filesChan {

	// 	relPath, _ := filepath.Rel(rootDir, file.Path)
	// 	fmt.Println(file.Status, relPath)
	// }
}
