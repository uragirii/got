package cmd

import (
	"fmt"

	"github.com/uragirii/got/internals"
	"github.com/uragirii/got/internals/color"
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

	var modifiedFiles []object.ChangeItem
	var untrackedFiles []string

	for _, change := range changes {
		if change.Status != object.StatusAdded {
			modifiedFiles = append(modifiedFiles, change)
		} else {
			untrackedFiles = append(untrackedFiles, change.RelPath)
		}
	}

	fmt.Println("On branch", head.Branch)

	if len(modifiedFiles) > 0 {
		fmt.Println("Changes not staged for commit:")
		fmt.Println(`  (use "git add <file>..." to update what will be committed)`)
		fmt.Println(`  (use "git add <file>..." to update what will be committed)`)
		for _, file := range modifiedFiles {
			fmt.Printf("\t%s\n", color.RedString(fmt.Sprintf("%s:   %s", file.Status.String(), file.RelPath)))
		}
		fmt.Println()
	}
	if len(untrackedFiles) > 0 {
		fmt.Println("Untracked files:")
		fmt.Println(`  (use "git add <file>..." to include in what will be committed)`)
		for _, file := range untrackedFiles {
			fmt.Printf("\t%s\n", color.RedString(file))
		}
		fmt.Println()
	}
	fmt.Println(`no changes added to commit (use "git add" and/or "git commit -a")`)
}
