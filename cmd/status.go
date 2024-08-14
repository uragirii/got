package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/uragirii/got/internals"
	"github.com/uragirii/got/internals/color"
	"github.com/uragirii/got/internals/git/commit"
	"github.com/uragirii/got/internals/git/head"
	"github.com/uragirii/got/internals/git/index"
	"github.com/uragirii/got/internals/git/tree"
)

var STATUS *internals.Command = &internals.Command{
	Name:  "status",
	Desc:  "Show the working tree status",
	Flags: []*internals.Flag{},
	Run:   Status,
}

func Status(c *internals.Command, _ string) {

	gitDir, err := internals.GetGitDir()

	if err != nil {
		panic(err)
	}

	gitFs := os.DirFS(gitDir)

	head, err := head.New(gitFs)

	if err != nil {
		panic(err)
	}

	commit, err := commit.FromSHA(head.SHA, gitFs)

	if err != nil {
		panic(err)
	}

	rootFs := os.DirFS(path.Join(gitDir, ".."))

	treeObj, err := tree.FromDir(rootFs)

	if err != nil {
		panic(err)
	}

	changes, err := commit.Tree.Compare(treeObj, gitFs)

	if err != nil {
		panic(err)
	}

	file, err := gitFs.Open(index.IndexFileName)

	if err != nil {
		panic(err)
	}

	indexFile, err := index.New(file)

	if err != nil {
		panic(err)
	}

	var modifiedFiles []tree.ChangeItem
	var untrackedFiles []string
	var stagedFiles []tree.ChangeItem

	// TODO: FIXME when file is staged and then changed, it is shown as not staged
	// comapare index sha with live sha and commit sha, if both are different, its above case
	// ignoring for now
	for _, change := range changes {
		if change.Status != tree.StatusAdded {
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
