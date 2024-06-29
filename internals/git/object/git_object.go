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
	"github.com/uragirii/got/internals/git"
)

type ObjectType string

const _ObjectsDir string = "objects"

const (
	BlobObj   ObjectType = "blob"
	TreeObj   ObjectType = "tree"
	CommitObj ObjectType = "commit"
)

var ErrInvalidObj = fmt.Errorf("invalid git object")

type Object struct {
	// Unmarshall(path string)
	// GetObjType() ObjectType
	// PrettyPrint()
	objectType ObjectType
	// Also contains header for the object
	uncompressedContents *[]byte
	sha                  *git.SHA
}

func NewObjectFromSHA(sha *git.SHA) (*Object, error) {
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

	return &Object{
		sha:                  sha,
		objectType:           objType,
		uncompressedContents: decompressedContents,
	}, nil
}

func (obj *Object) getContentWithoutHeader() *[]byte {
	for i := 0; i < len(*obj.uncompressedContents); i++ {
		if (*obj.uncompressedContents)[i] == '\u0000' {
			contents := (*obj.uncompressedContents)[i+1:]
			return &contents
		}
	}

	return nil
}

// TODO:
// Pretty print the obj
func (obj *Object) String() string {
	if obj.objectType == TreeObj {
		panic("pretty print not implemented")
	}
	return fmt.Sprint(string(*obj.getContentWithoutHeader()))
}

func (obj *Object) RawString() string {
	return fmt.Sprint(string(*(obj.getContentWithoutHeader())))
}

func (obj *Object) GetObjType() ObjectType {
	return obj.objectType
}

func (obj *Object) Write() error {
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

func (obj *Object) GetSHA() *git.SHA {
	return obj.sha
}

func NewObject(filePath string) (*Object, error) {
	data, err := os.ReadFile(filePath)

	if err != nil {
		return nil, err
	}

	header := []byte(fmt.Sprintf(_GitBlobHeader, len(data)))

	contents := append(header, data...)

	hash := sha1.Sum(contents)

	hashSlice := hash[:]

	sha, err := git.SHAFromByteSlice(&hashSlice)

	if err != nil {
		return nil, err
	}

	return &Object{
		objectType:           BlobObj,
		uncompressedContents: &contents,
		sha:                  sha,
	}, nil
}

func getObjectPath(sha *git.SHA) (string, error) {
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

func getObjType(decompressedContents *[]byte) (ObjectType, error) {
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
