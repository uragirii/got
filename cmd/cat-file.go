package cmd

import (
	"fmt"

	"github.com/uragirii/got/internals"
	"github.com/uragirii/got/internals/git/object"
	"github.com/uragirii/got/internals/git/sha"
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

// func printTree(obj *[]byte) {
// 	for idx := 0; idx < len(*obj); {
// 		modeStartIdx := idx
// 		for ; (*obj)[idx] != 0x20; idx++ {
// 		}

// 		mode := string((*obj)[modeStartIdx:idx])

// 		if mode == "40000" {
// 			mode = "040000"
// 		}

// 		nameStartIdx := idx

// 		for ; (*obj)[idx] != 0x00; idx++ {
// 		}

// 		name := string((*obj)[nameStartIdx:idx])

// 		// get over the \0
// 		idx++

// 		sha := (*obj)[idx : idx+20]

// 		idx += 20

// 		shaStr := fmt.Sprintf("%x", sha)

// 		objType, _, err := internals.ReadObject(shaStr)

// 		if err != nil {
// 			panic(err)
// 		}

// 		fmt.Printf("%s %s %s\t%s\n", mode, objType, shaStr, name)

// 	}
// }

func CatFile(c *internals.Command, gitDir string) {
	argSha := c.Args[0]

	sha, err := sha.FromString(argSha)

	if err != nil {
		panic(err)
	}

	obj, err := object.NewObjectFromSHA(sha)

	if err != nil {
		panic(err)
	}

	if c.GetFlag("type") == "true" {
		fmt.Println(obj.GetObjType())
		return
	}

	if c.GetFlag("pretty") == "true" {
		fmt.Println(obj)
	} else {
		fmt.Println(obj.RawString())
	}

}
