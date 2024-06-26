package cmd

import (
	"github.com/uragirii/got/internals"
)

var STATUS *internals.Command = &internals.Command{
	Name:  "status",
	Desc:  "Show the working tree status",
	Flags: []*internals.Flag{},
	Run:   Status,
}

func Status(c *internals.Command, gitPath string) {

	// head, err := internals.GetHeadSHA()

	// if err != nil {
	// 	panic(err)
	// }

	// commit, err := internals.ParseCommit(head.SHA)

	// if err != nil {
	// 	panic(err)
	// }

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
