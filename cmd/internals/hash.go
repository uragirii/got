package internals

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"slices"
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

func HashTree(tree *dirTree) (*[20]byte, error) {

	items := make([]string, 0, len(tree.childFiles)+len(tree.childDirs))

	for file := range tree.childFiles {
		relPath, _ := filepath.Rel(tree.Path, file)

		items = append(items, relPath)
	}
	for subdir := range tree.childDirs {
		relPath, _ := filepath.Rel(tree.Path, subdir)

		items = append(items, relPath)
	}

	slices.Sort(items)

	var content strings.Builder

	for _, item := range items {
		absPath := path.Join(tree.Path, item)

		if tree.childFiles[absPath] != nil {
			// is file
			// get blob
			fileInfo, err := os.Stat(absPath)

			if err != nil {
				return nil, err
			}

			mode := "100644"

			if !fileInfo.Mode().IsRegular() {
				mode = "100755"
			}

			content.WriteString(fmt.Sprintf("%s %s\u0000", mode, item))

			decodedSha, err := hex.DecodeString(tree.childFiles[absPath].SHA)

			if err != nil {
				return nil, err
			}
			for _, b := range decodedSha {
				content.WriteByte(b)
			}

		} else {
			// is folder
			mode := "40000"
			content.WriteString(fmt.Sprintf("%s %s\u0000", mode, item))

			decodedSha, err := hex.DecodeString(tree.childDirs[absPath].SHA)

			if err != nil {
				return nil, err
			}
			for _, b := range decodedSha {
				content.WriteByte(b)
			}
		}
	}

	contentLen := content.Len()

	header := []byte(fmt.Sprintf("tree %d\u0000", contentLen))

	contentBytes := append(header, []byte(content.String())...)

	hash := sha1.Sum(contentBytes)

	return &hash, nil
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
