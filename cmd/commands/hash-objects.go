package commands

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/uragirii/got/cmd/internals"
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

func HashObject(c *internals.Command) {

	var wg sync.WaitGroup
	var resultMtx sync.Mutex
	results := make([]string, len(c.Args))
	bytesBuffers := make([]*bytes.Buffer, len(c.Args))

	compress := c.GetFlag("write") == "true"

	for idx, arg := range c.Args {
		wg.Add(1)
		go func(arg string, idx int) {
			defer wg.Done()
			hash, bytesBuffer, err := internals.HashBlob(arg, compress)

			if err != nil {
				fmt.Println(err)
				// TODO: better error handing
				panic("error while hashing object")
			}
			resultMtx.Lock()
			results[idx] = fmt.Sprintf("%x", *hash)
			if compress {
				bytesBuffers[idx] = bytesBuffer
			}
			resultMtx.Unlock()
		}(arg, idx)
	}

	wg.Wait()
	// TODO: write the objects

	for _, result := range results {
		fmt.Println(result)
	}

}
