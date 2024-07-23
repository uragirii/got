package object

import (
	"bytes"
	"compress/zlib"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/uragirii/got/internals/git/sha"
)

var ErrInvalidObj = fmt.Errorf("invalid git object")

type Object struct {
	// Unmarshall(path string)
	// GetObjType() ObjectType
	// PrettyPrint()
	objectType ObjectType
	// Also contains header for the object
	uncompressedContents *[]byte
	sha                  *sha.SHA
}

func FromSHA(sha *sha.SHA, fs fs.FS) (*Object, error) {
	objPath, err := sha.GetObjPath()

	if err != nil {
		return nil, err
	}

	objFile, err := fs.Open(objPath)

	if err != nil {
		return nil, err
	}

	defer objFile.Close()

	decompressedContents, err := Decompress(objFile)

	if err != nil {
		return nil, err
	}

	objType, err := GetType(decompressedContents)

	if err != nil {
		return nil, err
	}

	return &Object{
		sha:                  sha,
		objectType:           objType,
		uncompressedContents: decompressedContents,
	}, nil
}

func (obj *Object) RawString() string {
	return fmt.Sprint(string(*(obj.getContentWithoutHeader())))
}

func (obj *Object) GetObjType() ObjectType {
	return obj.objectType
}

func (obj *Object) Write(w io.Writer) error {

	writer := zlib.NewWriter(w)

	_, err := writer.Write(*obj.uncompressedContents)

	if err != nil {
		return err
	}

	writer.Close()

	return nil
}

func (obj *Object) WriteToFile() error {
	objPath, err := obj.sha.GetObjPath()

	if err != nil {
		return err
	}

	if _, err := os.Stat(objPath); errors.Is(err, os.ErrNotExist) {

		var compressBytes bytes.Buffer

		err := obj.Write(&compressBytes)

		if err != nil {
			return err
		}

		return os.WriteFile(objPath, compressBytes.Bytes(), 0444)
	}
	return nil
}

func (obj *Object) GetSHA() *sha.SHA {
	return obj.sha
}

// Creates a new object from the raw file
// It doesn't read the existing object, instead hashes a file
func FromFile(reader io.Reader) (*Object, error) {
	data, err := io.ReadAll(reader)

	if err != nil {
		return nil, err
	}

	header := []byte(fmt.Sprintf(_GitBlobHeader, len(data)))

	contents := append(header, data...)

	sha, err := sha.FromData(&contents)

	if err != nil {
		return nil, err
	}

	return &Object{
		objectType:           BlobObj,
		uncompressedContents: &contents,
		sha:                  sha,
	}, nil
}
