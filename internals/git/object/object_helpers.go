package object

import (
	"compress/zlib"
	"io"
	"slices"
)

func Decompress(reader io.Reader) (*[]byte, error) {
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

func (obj *Object) getContentWithoutHeader() *[]byte {

	for i := 0; i < len(*obj.uncompressedContents); i++ {
		if (*obj.uncompressedContents)[i] == '\u0000' {
			contents := (*obj.uncompressedContents)[i+1:]
			return &contents
		}
	}

	return nil
}

func GetType(decompressedContents *[]byte) (ObjectType, error) {
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
