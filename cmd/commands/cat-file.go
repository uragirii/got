package commands

import (
	"fmt"

	"github.com/uragirii/got/cmd/internals"
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

func CatFile(c *internals.Command, gitDir string) {
	sha := c.Args[0]

	decoded, err := internals.DecodeHash(gitDir, sha)

	if err != nil {
		fmt.Println(err)
		return
	}

	objType, content := internals.GetObj(decoded)

	if content == nil {
		panic("no content found")
	}

	if c.GetFlag("type") == "true" {
		fmt.Println(objType)
		return
	}

	fmt.Println(string(*content))

}
