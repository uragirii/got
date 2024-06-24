package object

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"path"
	"slices"

	"github.com/uragirii/got/internals"
)

type GitObjectType string

const _ObjectsDir string = "objects"

const (
	BlobObj   GitObjectType = "blob"
	TreeObj   GitObjectType = "tree"
	CommitObj GitObjectType = "commit"
)

var ErrInvalidObj = fmt.Errorf("invalid git object")

type GitObject struct {
	// Unmarshall(path string)
	// GetObjType() GitObjectType
	// PrettyPrint()
	objectType           GitObjectType
	uncompressedContents *[]byte
	sha                  *internals.SHA
}

func getObjectPath(sha *internals.SHA) (string, error) {
	gitDir, err := internals.GetGitDir()

	if err != nil {
		return "", err
	}

	objectsDir := path.Join(gitDir, _ObjectsDir)

	shaStr := sha.MarshallToStr()

	objPath := path.Join(objectsDir, shaStr[0:2], shaStr[2:])

	return objPath, nil
}

func decompressObj(contents *[]byte) (*[]byte, error) {
	b := bytes.NewReader(*contents)

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

func getObjType(decompressedContents *[]byte) (GitObjectType, error) {
	headerEndIdx := slices.Index(*decompressedContents, 0x00)

	if headerEndIdx == -1 {
		return BlobObj, ErrInvalidObj
	}

	header := (*decompressedContents)[:headerEndIdx]

	headerSpaceIdx := slices.Index(header, 0x20) // SPACE

	if headerSpaceIdx == -1 {
		return BlobObj, ErrInvalidObj
	}

	objType := string(header[:headerSpaceIdx])

	switch objType {
	case string(BlobObj):
		return BlobObj, nil

	case string(TreeObj):
		return TreeObj, nil

	case string(CommitObj):
		return CommitObj, nil

	default:
		return BlobObj, ErrInvalidObj

	}
}
