package cmd

import (
	"fmt"
	"os"
	"sync"

	"github.com/uragirii/got/internals"
	"github.com/uragirii/got/internals/git/blob"
)

var HASH_OBJECT *internals.Command = &internals.Command{
	Name: "hash-object",
	Desc: "Compute object ID and optionally create an object from a file",
	Flags: []*internals.Flag{
		{
			Name:  "t",
			Short: "",
			Help:  "object type",
			Key:   "type",
			Type:  internals.String,
		},
		{
			Name:  "w",
			Short: "",
			Help:  "write the object into the object database",
			Key:   "write",
			Type:  internals.Bool,
		},
	},
	Run: HashObject,
}

func HashObject(c *internals.Command, _ string) {

	var wg sync.WaitGroup
	results := make([]string, len(c.Args))

	compress := c.GetFlag("write") == "true"

	for idx, arg := range c.Args {
		wg.Add(1)
		go func(arg string, idx int) {
			defer wg.Done()

			file, err := os.Open(arg)

			if err != nil {
				panic(err)
			}

			obj, err := blob.FromFile(file)

			if err != nil {
				panic(err)
			}

			results[idx] = obj.GetSHA().String()

			if compress {
				obj.WriteToFile()
			}

		}(arg, idx)
	}

	wg.Wait()

	for _, result := range results {
		fmt.Println(result)
	}

}
