package blob

import (
	"bytes"
	"compress/zlib"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"

	"github.com/uragirii/got/internals"
	"github.com/uragirii/got/internals/git/object"
	"github.com/uragirii/got/internals/git/sha"
)

var ErrInvalidObjType = fmt.Errorf("object is not blob type")

type Blob struct {
	// Doesn't contain the header
	contents *[]byte
	sha      *sha.SHA
}

func FromSHA(sha *sha.SHA, fsys fs.FS) (*Blob, error) {

	objContents, err := object.FromSHA(sha, fsys)

	if err != nil {
		return nil, err
	}

	if objContents.ObjType != object.BlobObj {
		return nil, ErrInvalidObjType
	}

	return &Blob{
		sha:      sha,
		contents: objContents.Contents,
	}, nil
}

func (blob Blob) String() string {
	return string(*blob.contents)
}

// func (obj *Blob) RawString() string {
// 	return fmt.Sprint(string(*(obj.getContentWithoutHeader())))
// }

func (blob Blob) GetObjType() object.ObjectType {
	return object.BlobObj
}

func (blob Blob) getContentWithHeader() *[]byte {
	var buffer bytes.Buffer

	contentLn := len(*blob.contents)

	buffer.Write([]byte(fmt.Sprintf(object.BlobHeader, contentLn)))

	buffer.Write(*blob.contents)

	b := buffer.Bytes()

	return &b
}

func (blob Blob) Write(w io.Writer) error {

	writer := zlib.NewWriter(w)

	_, err := writer.Write(*blob.getContentWithHeader())

	if err != nil {
		return err
	}

	writer.Close()

	return nil
}

func (blob Blob) WriteToFile() error {
	blobPath, err := blob.sha.GetObjPath()

	if err != nil {
		return err
	}

	gitDir, err := internals.GetGitDir()

	if err != nil {
		return err
	}

	blobPath = path.Join(gitDir, blobPath)

	if _, err := os.Stat(blobPath); errors.Is(err, os.ErrNotExist) {

		var compressBytes bytes.Buffer

		err := os.MkdirAll(path.Join(blobPath, ".."), 0755)

		if err != nil {
			return err
		}

		err = blob.Write(&compressBytes)

		if err != nil {
			return err
		}

		return os.WriteFile(blobPath, compressBytes.Bytes(), 0444)
	}
	return nil
}

func (blob Blob) GetSHA() *sha.SHA {
	return blob.sha
}

func (blob Blob) Raw() string {
	return blob.String()
}

// Creates a new in memory blob from the raw file
// It doesn't read the existing object, instead hashes a file
func FromFile(reader io.Reader) (*Blob, error) {
	data, err := io.ReadAll(reader)

	if err != nil {
		return nil, err
	}

	blob := Blob{
		contents: &data,
	}

	sha, err := sha.FromData(blob.getContentWithHeader())

	if err != nil {
		return nil, err
	}

	blob.sha = sha

	return &blob, nil
}
