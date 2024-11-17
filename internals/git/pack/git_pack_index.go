package pack

import (
	"bytes"
	"io/fs"

	"github.com/uragirii/got/internals/git/sha"
)

type idxItem struct {
	offset         int64
	compressedSize int64
}

type PackIndex struct {
	offsetMap map[string]idxItem
}

func (idx PackIndex) GetObjOffset(sha *sha.SHA) (idxItem, bool) {
	offset, ok := idx.offsetMap[sha.String()]
	return offset, ok
}

func FromIdxFile(fsys fs.FS, path string) (*PackIndex, error) {
	file, err := fsys.Open(path)

	if err != nil {
		return nil, err
	}

	defer file.Close()

	var buf bytes.Buffer

	if _, err = buf.ReadFrom(file); err != nil {
		return nil, err
	}

	// TODO: verify idx file

	idxBytes := buf.Bytes()

	cummulativeCountBytes := idxBytes[8 : 8+0x100*4]

	idxBytes = idxBytes[8+0x100*4:]

	prefixCounters := make([]int64, 0x100)

	var prevCount int64

	for idx := range 0x100 {
		count, err := fourBytesToInt(cummulativeCountBytes[idx*4 : (idx+1)*4])

		if err != nil {
			return nil, err
		}

		prefixCounters[idx] = count - prevCount

		prevCount = count
	}

	shaList := make([]*sha.SHA, prevCount)

	cummulativeSHABytes := idxBytes[:prevCount*20]

	idxBytes = idxBytes[prevCount*20:]

	for idx := range prevCount {
		sl := cummulativeSHABytes[idx*20 : (idx+1)*20]
		s, err := sha.FromByteSlice(&sl)

		if err != nil {
			return nil, err
		}

		shaList[idx] = s
	}

	// Next is CRC, which we are skipping for now
	// TODO: Verify CRC checks

	idxBytes = idxBytes[prevCount*4:]

	offsetBytes := idxBytes[:prevCount*4]
	idxBytes = idxBytes[prevCount*4:]

	offsetMap := make(map[string]idxItem, prevCount)

	// now offset
	// FIXME: This doesn't support files larger than 2GB, but ig thats fine.
	// TODO: add tests for larger pack files.
	for idx := range prevCount {
		sl := offsetBytes[idx*4 : (idx+1)*4]
		offset, _ := fourBytesToInt(sl)

		var compressedSize int64

		if idx < prevCount-1 {
			//TODO: FIXME last item compressed size??

			nextOffset, _ := fourBytesToInt(offsetBytes[(idx+1)*4 : (idx+2)*4])
			compressedSize = nextOffset - offset
		}

		offsetMap[shaList[idx].String()] = idxItem{
			offset:         offset,
			compressedSize: compressedSize,
		}
		// fmt.Printf("%s %d\n", shaList[idx].String(), offset)
	}

	return &PackIndex{
		offsetMap: offsetMap,
	}, nil

}
