package object

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"io"
	"os"
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
	objectType GitObjectType
	// Also contains header for the object
	uncompressedContents *[]byte
	sha                  *internals.SHA
}

func UnmarshallGitObject(sha *internals.SHA) (*GitObject, error) {
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

	return &GitObject{
		sha:                  sha,
		objectType:           objType,
		uncompressedContents: decompressedContents,
	}, nil
}

func (obj *GitObject) String() string {
	return fmt.Sprintf("%s", *(obj.uncompressedContents))
}

func (obj *GitObject) RawString() string {
	return fmt.Sprintf("%s", *(obj.uncompressedContents))
}

func (obj *GitObject) GetObjType() GitObjectType {
	return obj.objectType
}

func (obj *GitObject) Write() error {
	objPath, err := getObjectPath(obj.sha)

	if err != nil {
		return err
	}

	var compressBytes bytes.Buffer

	writer := zlib.NewWriter(&compressBytes)

	_, err = writer.Write(*obj.uncompressedContents)

	if err != nil {
		return err
	}

	writer.Close()

	return os.WriteFile(objPath, compressBytes.Bytes(), 0444)

}

func (obj *GitObject) GetSHA() *internals.SHA {
	return obj.sha
}

func NewGitObject(filePath string) (*GitObject, error) {
	data, err := os.ReadFile(filePath)

	if err != nil {
		return nil, err
	}

	header := []byte(fmt.Sprintf(_GitBlobHeader, len(data)))

	contents := append(header, data...)

	fmt.Println(string(contents))

	hash := sha1.Sum(contents)

	hashSlice := hash[:]

	sha, err := internals.SHAFromByteSlice(&hashSlice)

	if err != nil {
		return nil, err
	}

	return &GitObject{
		objectType:           BlobObj,
		uncompressedContents: &contents,
		sha:                  sha,
	}, nil
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
