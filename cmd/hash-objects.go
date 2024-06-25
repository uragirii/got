package cmd

import (
	"fmt"
	"sync"

	"github.com/uragirii/got/internals"
	"github.com/uragirii/got/internals/git/object"
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

			obj, err := object.NewGitObject(arg)

			if err != nil {
				panic(err)
			}

			results[idx] = obj.GetSHA().MarshallToStr()

			if compress {
				obj.Write()
			}

		}(arg, idx)
	}

	wg.Wait()

	for _, result := range results {
		fmt.Println(result)
	}

}
