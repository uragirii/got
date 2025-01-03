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
var ErrOFSDeltaNotImplemented = errors.New("OFS_DELTA not implemented")

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

func (pack Pack) getOfsDeltaObj(r bytes.Reader, offset uint32) {
	offsetBytes := []byte{}

	r.Seek(int64(offset), io.SeekStart)

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

	// baseObjOffset := offset - uint32(baseObjOffsetDiff)

	instructions, err := object.Decompress(&r)

	if err != nil {
		panic(err)
	}

	var objData []byte

	// for _, ins := range *instructions {
	// 	fmt.Printf("%08b\n", ins)
	// }

	for i := 0; i < len((*instructions)); i++ {
		instruction := (*instructions)[i]

		fmt.Printf("Instruction %08b\n", instruction)

		if (instruction & 0b1000_0000) != 0b1000_0000 {
			// 0xxxxxxx means data to copy

			dataToCopy := (*instructions)[i+1 : i+int(instruction)]

			fmt.Printf("Copying data \"%s\"\n", dataToCopy)

			i += int(instruction)
			objData = append(objData, dataToCopy...)

			continue
		}

		offsetMask := instruction & 0b1111

		var offsetSlice [4]byte

		i++
		for idx := range 4 {

			hasOffset := offsetMask & 0b1

			if hasOffset == 1 {
				offsetSlice[4-idx-1] = (*instructions)[i]
				i++
			}

			offsetMask = offsetMask >> 1
		}

		fmt.Printf("%08b\n", offsetSlice)

		off := 0

		for _, o := range offsetSlice {
			off = (off << 8) + int(o)
		}

		fmt.Println("Offset", off)

		sizeMask := (instruction & 0b0111_0000) >> 4

		var sizeSlice [3]byte

		i++
		for idx := range 3 {
			hasSize := sizeMask & 0b1

			if hasSize == 1 {
				sizeSlice[3-idx-1] = (*instructions)[idx]
				i++
			}

			sizeMask = sizeMask >> 1
		}

		fmt.Println(sizeSlice)

		size := 0

		for _, o := range sizeSlice {
			size = (size << 8) + int(o)
		}

		fmt.Println("Size", size)

	}

	fmt.Println(string(objData))
}

func (pack Pack) GetObj(objSha *sha.SHA) (object.ObjectContents, error) {

	item, ok := pack.idx.GetObjOffset(objSha)

	if !ok {
		return object.ObjectContents{}, ErrObjNotFound
	}

	offset := item.Offset

	pack.fileReader.Seek(int64(offset), io.SeekStart)

	objType, size, err := parseObjTypeAndSize(&pack.fileReader)

	if err != nil {
		return object.ObjectContents{}, err
	}

	if objType == _REF_DELTA {
		fmt.Println("REF DELTA CASE")

		return object.ObjectContents{}, ErrOFSDeltaNotImplemented
	}

	if objType == _OFS_DELTA {
		fmt.Println("OFS DELTA CASE")
		// pack.getOfsDeltaObj(pack.fileReader, offset+2)

		return object.ObjectContents{}, ErrOFSDeltaNotImplemented
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

func ParsePackFile(r bytes.Reader, idx *PackIndex) *Pack {
	return &Pack{
		idx:        idx,
		fileReader: r,
	}
}
