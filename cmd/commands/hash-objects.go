package commands

import (
	"crypto/sha1"
	"flag"
	"fmt"
	"os"
)

func HashObject() {
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
