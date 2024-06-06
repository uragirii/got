package internals

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"os"
)

func HashBlob(path string, compress bool) (*[20]byte, *bytes.Buffer, error) {
	data, err := os.ReadFile(path)

	if err != nil {
		return nil, nil, err
	}

	header := []byte(fmt.Sprintf("blob %d\u0000", len(data)))

	contents := append(header, data...)

	hash := sha1.Sum(contents)

	if !compress {
		return &hash, nil, nil
	}

	var compressBytes bytes.Buffer

	writer := zlib.NewWriter(&compressBytes)

	_, err = writer.Write(contents)

	if err != nil {
		//Should we return hash here?
		return nil, nil, err
	}

	writer.Flush()

	return &hash, &compressBytes, nil

}
