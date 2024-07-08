package object

import (
	"crypto/sha1"
	"fmt"
	"os"
	"slices"

	"github.com/uragirii/got/internals/git/sha"
)

type GitBlob struct {
	contents *[]byte
	SHA      *sha.SHA
}

const _GitBlobHeader string = "blob %d\u0000"

func MarshalGitBlobFromSHA(sha *sha.SHA) (*GitBlob, error) {
	objPath, err := getObjectPath(sha)

	if err != nil {
		return nil, err
	}

	contents, err := os.ReadFile(objPath)

	if err != nil {
		return nil, err
	}

	decompressedContents, err := decompressObj(&contents)

	if err != nil {
		return nil, err
	}

	objType, err := getObjType(decompressedContents)

	if err != nil {
		return nil, err
	}

	if objType != BlobObj {
		return nil, ErrInvalidObj
	}

	headerEndIdx := slices.Index(*decompressedContents, 0x00)

	// No need to check headerEndIdx as it will be >0
	// check is already done inside getObjType

	actualContents := (*decompressedContents)[headerEndIdx+1:]

	return &GitBlob{
		contents: &actualContents,
		SHA:      sha,
	}, nil
}

func (blob *GitBlob) PrettyPrint() {
	fmt.Println(string(*blob.contents))
}

func MarshalGitBlobFromFile(filePath string) (*GitBlob, error) {
	fileContents, err := os.ReadFile(filePath)

	if err != nil {
		return nil, err
	}

	header := []byte(fmt.Sprintf(_GitBlobHeader, len(fileContents)))

	blobContents := append(header, fileContents...)

	hashArr := sha1.Sum(blobContents)
	hashSlice := hashArr[:]

	sha, err := sha.FromByteSlice(&hashSlice)

	if err != nil {
		return nil, err
	}

	return &GitBlob{
		contents: &fileContents,
		SHA:      sha,
	}, nil
}

func (blob *GitBlob) Write() {
	panic("not implemented")
}
