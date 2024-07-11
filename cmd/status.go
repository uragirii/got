package cmd

import (
	"fmt"

	"github.com/uragirii/got/internals"
	"github.com/uragirii/got/internals/color"
	"github.com/uragirii/got/internals/git/head"
	"github.com/uragirii/got/internals/git/index"
	"github.com/uragirii/got/internals/git/object"
)

var STATUS *internals.Command = &internals.Command{
	Name:  "status",
	Desc:  "Show the working tree status",
	Flags: []*internals.Flag{},
	Run:   Status,
}

func Status(c *internals.Command, gitPath string) {

	head, err := head.New()

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

	indexFile, err := index.New()

	if err != nil {
		panic(err)
	}

	var modifiedFiles []object.ChangeItem
	var untrackedFiles []string
	var stagedFiles []object.ChangeItem

	// TODO: FIXME when file is staged and then changed, it is shown as not staged
	// comapare index sha with live sha and commit sha, if both are different, its above case
	// ignoring for now
	for _, change := range changes {
		if change.Status != object.StatusAdded {
			indexEntry := indexFile.Get(change.RelPath)

			if indexEntry.SHA.Eq(change.SHA) {
				// file is staged
				stagedFiles = append(stagedFiles, change)
			} else {
				modifiedFiles = append(modifiedFiles, change)
			}
		} else {
			if indexFile.Has(change.RelPath) {
				stagedFiles = append(stagedFiles, change)

			} else {
				untrackedFiles = append(untrackedFiles, change.RelPath)
			}
		}
	}

	fmt.Println("On branch", head.Branch)

	if len(stagedFiles) > 0 {
		fmt.Println("Changes to be committed:")
		fmt.Println(`  (use "git restore --staged <file>..." to unstage)`)
		for _, file := range stagedFiles {
			fmt.Printf("\t%s\n", color.GreenString(fmt.Sprintf("%s:   %s", file.Status.String(), file.RelPath)))
		}
		fmt.Println()
	}

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

	if len(stagedFiles) > 0 {
		return
	}
	if len(changes) > 0 {
		fmt.Println(`no changes added to commit (use "git add" and/or "git commit -a")`)
	} else {
		fmt.Println("nothing to commit, working tree clean")
	}
}
