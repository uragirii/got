package commands

import (
	"crypto/sha1"
	"flag"
	"fmt"
	"os"

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
	filename := flag.Arg(1)

	if filename == "" {
		return
	}

	data, err := os.ReadFile(filename)

	if err != nil {
		fmt.Println(err)
		fmt.Println("error while reading file")
		return
	}

	fmt.Println(c.Args)

	header := []byte(fmt.Sprintf("blob %d\u0000", len(data)))

	contents := append(header, data...)

	hash := fmt.Sprintf("%x", sha1.Sum(contents))

	fmt.Printf("%s\n", hash)

	// var compressBytes bytes.Buffer

	// writer := zlib.NewWriter(&compressBytes)

	// writer.Write(contents)

	// writer.Flush()

	// os.WriteFile("hashed", compressBytes.Bytes(), 0660)

}
