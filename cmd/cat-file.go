package cmd

import (
	"fmt"
	"os"

	"github.com/uragirii/got/internals"
	"github.com/uragirii/got/internals/git/blob"
	"github.com/uragirii/got/internals/git/commit"
	"github.com/uragirii/got/internals/git/object"
	"github.com/uragirii/got/internals/git/sha"
	"github.com/uragirii/got/internals/git/tree"
)

var CAT_FILE *internals.Command = &internals.Command{
	Name: "cat-file",
	Desc: "Provide contents or details of repository objects",
	Flags: []*internals.Flag{
		{
			Name:  "t",
			Short: "",
			Help:  "Instead of the content, show the object type identified by <object>",
			Key:   "type",
			Type:  internals.Bool,
		},
		{
			Name:  "p",
			Short: "",
			Help:  "Pretty-print the contents of <object> based on its type.",
			Key:   "pretty",
			Type:  internals.Bool,
		},
	},
	Run: CatFile,
}

func CatFile(c *internals.Command, _ string) {

	var objType, argSha string

	if len(c.Args) == 2 {
		objType = c.Args[0]
		argSha = c.Args[1]
	} else if len(c.Args) == 1 {
		argSha = c.Args[0]
	} else {
		panic("fatal: invalid number of arguments")
	}

	gitDir, err := internals.GetGitDir()

	if err != nil {
		panic(err)
	}

	gitFs := os.DirFS(gitDir)

	sha, err := sha.FromString(argSha)

	if err != nil {
		panic(err)
	}

	obj, err := object.FromSHA(sha, gitFs)

	if err != nil {
		panic(err)
	}

	if c.GetFlag("type") == "true" {
		fmt.Println(obj.ObjType)
		return
	}

	if c.GetFlag("pretty") != "true" {
		if !object.IsValidObjectType(objType) {
			panic(fmt.Sprintf("fatal: invalid object type \"%s\"", objType))
		}
		if objType != string(obj.ObjType) {
			panic(fmt.Sprintf("fatal: git cat-file %s: bad file", argSha))
		}

		fmt.Printf("%s\n", *obj.Contents)
		return
	}

	switch obj.ObjType {
	case object.BlobObj:
		blob, err := blob.FromSHA(sha, gitFs)

		if err != nil {
			panic(err)
		}

		fmt.Println(blob.String())
	case object.CommitObj:
		commit, err := commit.FromSHA(sha, gitFs)

		if err != nil {
			panic(err)
		}

		fmt.Println(commit.String())

	case object.TreeObj:
		tree, err := tree.FromSHA(sha, gitFs)

		if err != nil {
			panic(err)
		}

		fmt.Println(tree.String())
	default:
		panic(fmt.Sprintf("fatal: git cat-file %s: bad file", argSha))
	}

}
