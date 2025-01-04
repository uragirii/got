package pack

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"slices"

	"github.com/uragirii/got/internals/git/object"
	"github.com/uragirii/got/internals/git/sha"
)

type packObjType byte

const (
	_COMMIT    packObjType = 0b001
	_TREE      packObjType = 0b010
	_BLOB      packObjType = 0b011
	_TAG       packObjType = 0b100
	_OFS_DELTA packObjType = 0b110
	_REF_DELTA packObjType = 0b111
)

func (objType packObjType) String() string {
	switch objType {
	case _BLOB:
		return "blob"
	case _COMMIT:
		return "commit"
	case _TREE:
		return "tree"
	case _TAG:
		return "tag"
	case _OFS_DELTA:
		return "ofs"
	case _REF_DELTA:
		return "ref"
	}
	return "na"
}

func (objType packObjType) ToGitObject() object.ObjectType {
	switch objType {
	case _BLOB:
		return object.BlobObj
	case _COMMIT:
		return object.CommitObj
	case _TREE:
		return object.TreeObj
	default:
		panic(fmt.Sprintf("ToGitObject can be called only on blob, commit or tree but called on %s", objType.String()))
	}
}

// Pack would be used to get an object that is no longer loose.
// Ideally the API would be pack.GetObject(sha.SHA)(object, error)
// For now im doing one pack at a time, but i'd want to read all pack indexes and then
// store the map just in case

type Pack struct {
	idx        *PackIndex
	fileReader bytes.Reader
}

var ErrCantReadPackFile = errors.New("cannot read pack file")
var ErrObjNotFound = errors.New("object not found in pack file")
var ErrRefDeltaNotImplemented = errors.New("REF_DELTA not implemented")

func readOneByte(r *bytes.Reader, offset int64) byte {
	var b [1]byte

	_, err := r.ReadAt(b[:], int64(offset))

	if err != nil {
		fmt.Printf("non-reachable code, readOneByte:git_pack.go %v", err)
		panic(err)
	}

	return b[0]
}

func shouldReadMore(b byte) bool {
	// check is MSB is set, if set we need to read more
	return (b & 0b1000_0000) == 0b1000_0000
}

func getObjType(b byte) packObjType {
	return packObjType((b & 0b0111_0000) >> 4)
}

/**
* Pass the reader seeked to the correct offset
 */
func parseObjTypeAndSize(r *bytes.Reader) (packObjType, int, error) {

	offset, _ := r.Seek(0, io.SeekCurrent)

	firstByte := readOneByte(r, offset)
	offset++

	objType := getObjType(firstByte)

	b := firstByte

	firstByte = firstByte & 0b0000_1111

	// this is expanded size
	sizeBytes := []byte{}

	for ; shouldReadMore(b); offset++ {
		b = readOneByte(r, offset)

		sizeBytes = append(sizeBytes, b&0b0111_1111)
	}
	slices.Reverse(sizeBytes)

	size := 0

	for _, b := range sizeBytes {
		size = (size << 7) + int(b)
	}

	// this is UNCOMPRESSED size NOT compressed size
	// more below
	size = (size << 4) + int(firstByte)

	r.Seek(int64(offset), io.SeekStart)

	return objType, size, nil
}

func (pack Pack) parseOFSDeltaObj(r *bytes.Reader, ogOffset int64) (object.ObjectContents, error) {
	offsetBytes := []byte{}

	var b byte = 0x80

	for shouldReadMore(b) {
		b, _ = r.ReadByte()

		offsetBytes = append(offsetBytes, b&0b0111_1111)
	}

	baseObjOffsetDiff := 0
	correction := 0

	for idx, b := range offsetBytes {

		// n bytes with MSB set in all but the last one.
		// The offset is then the number constructed by
		// concatenating the lower 7 bit of each byte, and
		// for n >= 2 adding 2^7 + 2^14 + ... + 2^(7*(n-1))
		// to the result

		if idx > 0 {
			correction = (correction << 7) + 0x80
		}

		baseObjOffsetDiff = (baseObjOffsetDiff << 7) + int(b)
	}

	baseObjOffsetDiff += correction

	baseObjOffset := ogOffset - int64(baseObjOffsetDiff)

	baseObjContents, err := pack.GetObjAt(baseObjOffset)

	if err != nil {
		return object.ObjectContents{}, err
	}

	baseObj := *baseObjContents.Contents

	instructionsData, err := object.Decompress(r)

	if err != nil {
		return object.ObjectContents{}, err
	}

	var objData []byte

	instructionsReader := bytes.NewReader(*instructionsData)

	instructionsReader.Seek(4, io.SeekStart)

	for {
		instruction, err := instructionsReader.ReadByte()

		if err != nil {
			if errors.Is(err, io.EOF) {
				return object.ObjectContents{
					ObjType:  baseObjContents.ObjType,
					Contents: &objData,
				}, nil
			}

			return object.ObjectContents{}, err
		}

		if (instruction & 0b1000_0000) != 0b1000_0000 {
			// 0xxxxxxx means data to copy

			dataToCopy := make([]byte, instruction)

			instructionsReader.Read(dataToCopy)

			objData = append(objData, dataToCopy...)

			continue
		}

		offsetMask := instruction & 0b1111

		var offsetSlice [4]byte

		// i++
		for idx := range 4 {

			hasOffset := offsetMask & 0b1

			if hasOffset == 1 {
				offsetSlice[4-idx-1], _ = instructionsReader.ReadByte()
			}

			offsetMask = offsetMask >> 1
		}

		off := 0

		for _, o := range offsetSlice {
			off = (off << 8) + int(o)
		}

		sizeMask := (instruction & 0b0111_0000) >> 4

		var sizeSlice [3]byte

		for idx := range 3 {
			hasSize := sizeMask & 0b1

			if hasSize == 1 {
				sizeSlice[3-idx-1], _ = instructionsReader.ReadByte()
			}

			sizeMask = sizeMask >> 1
		}

		size := 0

		for _, o := range sizeSlice {
			size = (size << 8) + int(o)
		}

		if size == 0 {
			// size zero is automatically converted to 0x10000
			size = 0x10000
		}

		objData = append(objData, baseObj[off:off+size]...)

	}

}

func (pack Pack) GetObjAt(offset int64) (object.ObjectContents, error) {

	pack.fileReader.Seek(int64(offset), io.SeekStart)

	objType, size, err := parseObjTypeAndSize(&pack.fileReader)

	if err != nil {
		return object.ObjectContents{}, err
	}

	if objType == _REF_DELTA {
		fmt.Println("REF DELTA CASE")

		return object.ObjectContents{}, ErrRefDeltaNotImplemented
	}

	if objType == _OFS_DELTA {
		return pack.parseOFSDeltaObj(&pack.fileReader, offset)
	}

	data, err := object.Decompress(&pack.fileReader)

	if err != nil {
		return object.ObjectContents{}, err
	}

	if len(*data) != size {
		return object.ObjectContents{}, fmt.Errorf("expected decompressed size to be %d but got %d", size, len(*data))
	}

	return object.ObjectContents{
		Contents: data,
		ObjType:  objType.ToGitObject(),
	}, nil
}

func (pack Pack) GetObj(objSha *sha.SHA) (object.ObjectContents, error) {

	item, ok := pack.idx.GetObjOffset(objSha)

	if !ok {
		return object.ObjectContents{}, ErrObjNotFound
	}

	return pack.GetObjAt(int64(item.Offset))

}

func ParsePackFile(r bytes.Reader, idx *PackIndex) *Pack {
	return &Pack{
		idx:        idx,
		fileReader: r,
	}
}
