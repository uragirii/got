package object

import (
	"compress/zlib"
	"fmt"
	"io"
	"io/fs"
	"slices"
	"strconv"

	"github.com/uragirii/got/internals/git/sha"
)

var ErrInvalidObj = fmt.Errorf("invalid git object")

const BlobHeader string = "blob %d\u0000"
const TreeHeader string = "tree %d\u0000"
const CommitHeader string = "commit %d\u0000"

type ObjectType string

const (
	BlobObj   ObjectType = "blob"
	TreeObj   ObjectType = "tree"
	CommitObj ObjectType = "commit"
)

type Object interface {
	GetSHA() *sha.SHA
	Write(io.Writer) error
	GetObjType() ObjectType
	// Pretty print the object
	String() string
	Read(io.Reader)
}

type ObjectContents struct {
	ObjType ObjectType
	// These are decompressed contents
	Contents *[]byte
}

func decompress(reader io.Reader) (*[]byte, error) {
	reader, err := zlib.NewReader(reader)

	if err != nil {
		return nil, err
	}

	uncompressed, err := io.ReadAll(reader)

	if err != nil {
		return nil, err
	}

	return &uncompressed, nil
}

func FromSHA(sha *sha.SHA, fsys fs.FS) (ObjectContents, error) {
	objPath, err := sha.GetObjPath()

	if err != nil {
		return ObjectContents{}, err
	}

	objFile, err := fsys.Open(objPath)

	if err != nil {
		return ObjectContents{}, err
	}

	defer objFile.Close()

	decompressedContents, err := decompress(objFile)

	if err != nil {
		return ObjectContents{}, err
	}

	return getContents(decompressedContents)

}

func getContents(decompressedContents *[]byte) (ObjectContents, error) {
	headerEndIdx := slices.Index(*decompressedContents, 0x00)

	if headerEndIdx == -1 {
		return ObjectContents{}, ErrInvalidObj
	}

	header := (*decompressedContents)[:headerEndIdx]
	contents := (*decompressedContents)[headerEndIdx+1:]

	headerSpaceIdx := slices.Index(header, 0x20) // SPACE

	if headerSpaceIdx == -1 {
		return ObjectContents{}, ErrInvalidObj
	}

	byteLenStr := string((*decompressedContents)[headerSpaceIdx+1 : headerEndIdx])

	byteLen, err := strconv.Atoi(byteLenStr)

	if err != nil {
		fmt.Println("invalid bytelen", err)
		return ObjectContents{}, ErrInvalidObj
	}

	if byteLen != len(contents) {
		fmt.Printf("expected bytelen to be %d but found %d\n", byteLen, len(contents))
		return ObjectContents{}, ErrInvalidObj
	}

	objType := string(header[:headerSpaceIdx])

	switch ObjectType(objType) {
	case BlobObj:
		return ObjectContents{
			ObjType:  BlobObj,
			Contents: &contents,
		}, nil

	case TreeObj:
		return ObjectContents{
			ObjType:  TreeObj,
			Contents: &contents,
		}, nil

	case CommitObj:
		return ObjectContents{
			ObjType:  CommitObj,
			Contents: &contents,
		}, nil

	default:
		return ObjectContents{}, ErrInvalidObj

	}
}
