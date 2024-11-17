package pack

import (
	"errors"
	"fmt"
	"os"
	"slices"

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

// Pack would be used to get an object that is no longer loose.
// Ideally the API would be pack.GetObject(sha.SHA)(object, error)
// For now im doing one pack at a time, but i'd want to read all pack indexes and then
// store the map just in case

type Pack struct {
	idx      *PackIndex
	packFile *os.File
}

var ErrCantReadPackFile = errors.New("cannot read pack file")

func readOneByte(file *os.File, offset int64) byte {
	one := make([]byte, 1)
	n, err := file.ReadAt(one, offset)

	if err != nil || n != 1 {
		panic(ErrCantReadPackFile)
	}

	return one[0]
}

func shouldReadMore(b byte) bool {
	// check is MSB is set, if set we need to read more
	return b > 0b1000_0000
}

func getObjType(b byte) packObjType {
	return packObjType((b & 0b0111_0000) >> 4)
}

func getObjTypeAndSize(file *os.File, offset int64) (packObjType, *[]byte, error) {

	firstByte := readOneByte(file, offset)
	offset++

	objType := getObjType(firstByte)

	b := firstByte

	firstByte = firstByte & 0b0000_1111

	sizeBytes := []byte{}

	for ; shouldReadMore(b); offset++ {
		b = readOneByte(file, offset)

		sizeBytes = append(sizeBytes, b&0b0111_1111)
	}
	slices.Reverse(sizeBytes)

	size := 0

	for _, b := range sizeBytes {
		size = (size << 7) + int(b)
	}

	size = (size << 4) + int(firstByte)

	data := make([]byte, size)

	fmt.Println(offset)

	file.ReadAt(data, offset)

	return objType, &data, nil
}

func (pack Pack) GetObj(objSha *sha.SHA) {

	item, ok := pack.idx.GetObjOffset(objSha)

	if !ok {
		panic("obj not found")
	}

	offset := item.offset

	objType, data, err := getObjTypeAndSize(pack.packFile, offset)

	if err != nil {
		panic(err)
	}

	switch objType {
	case _REF_DELTA:
		baseObjShaBytes := (*data)[:sha.BYTES_LEN]
		baseObjSha, err := sha.FromByteSlice(&baseObjShaBytes)
		if err != nil {
			panic(err)
		}
		fmt.Printf("% x\n", *data)
		fmt.Println(baseObjSha.String())
		break
	}

}

func ParsePackFile(file *os.File, idx *PackIndex) *Pack {
	return &Pack{
		idx:      idx,
		packFile: file,
	}
}
