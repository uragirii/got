package internals

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
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

func DecodeHash(gitDir string, hash string) (*[]byte, error) {
	folder := hash[0:2]
	hashFile := hash[2:]

	files, err := os.ReadDir(path.Join(gitDir, "objects", folder))

	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if strings.HasPrefix(file.Name(), hashFile) || file.Name() == hashFile {
			contents, err := os.ReadFile(path.Join(gitDir, "objects", folder, file.Name()))

			if err != nil {
				return nil, err
			}

			b := bytes.NewReader(contents)

			reader, err := zlib.NewReader(b)

			if err != nil {
				return nil, err
			}

			uncompressed, err := io.ReadAll(reader)

			if err != nil {
				return nil, err
			}

			return &uncompressed, nil

		}
	}

	return nil, fmt.Errorf("hash not found")
}

func GetObjHeaderEnd(decodedBytes *[]byte) int {
	for idx, b := range *decodedBytes {
		if b == 0 {
			return idx
		}
	}

	return -1
}

func GetObj(decodedBytes *[]byte) (string, *[]byte) {
	headerSplitIdx := GetObjHeaderEnd(decodedBytes)

	header := (*decodedBytes)[:headerSplitIdx]
	content := (*decodedBytes)[headerSplitIdx:]

	for idx, b := range header {
		if b == 0x20 {
			objType := string(header[:idx])

			return objType, &content
		}
	}

	return "blob", nil
}
