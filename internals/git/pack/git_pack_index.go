package pack

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io/fs"

	"github.com/uragirii/got/internals/git/sha"
)

const _SUPPORTED_VERSION uint32 = 2
const _HeaderSize = 8
const _FanoutTableLen = 0x100
const _FanoutTableSize = _FanoutTableLen * 4 // 256 4-byte fanout enteries

var _MagicHeaderBytes = []byte{0xff, 0x74, 0x4f, 0x63} // \377tOc

var ErrVersionNotSupported = errors.New("only v2 index file is supported")
var ErrIndexParsing = errors.New("index file parsing went wrong")

type idxItem struct {
	Offset         uint32
	CompressedSize uint32
}

type PackIndex struct {
	offsetMap   map[string]idxItem
	offsetOrder []*sha.SHA
}

func (idx PackIndex) GetObjOffset(sha *sha.SHA) (idxItem, bool) {
	offset, ok := idx.offsetMap[sha.String()]
	return offset, ok
}

func verifyHeader(header []byte) error {
	if len(header) != 8 {
		panic("not-reachable: header should be 8 length")
	}

	magicBytes := header[0:4]

	// A 4-byte magic number \377tOc which is an unreasonable fanout[0] value
	for i, b := range _MagicHeaderBytes {
		if magicBytes[i] != b {
			return ErrVersionNotSupported
		}
	}

	versionHeaderBytes := header[4:8]

	// A 4-byte version number (= 2)
	if binary.BigEndian.Uint32(versionHeaderBytes) != _SUPPORTED_VERSION {
		return ErrVersionNotSupported
	}

	return nil
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

	idxBytes := buf.Bytes()

	verifyHeader(idxBytes[:_HeaderSize])

	fanoutTableBytes := idxBytes[_HeaderSize : _HeaderSize+_FanoutTableSize]

	idxBytes = idxBytes[_HeaderSize+_FanoutTableSize:]

	// The header consists of 256 4-byte network byte order integers.
	// N-th entry of this table records the number of objects in the
	// corresponding pack, the first byte of whose object name is
	// less than or equal to N. This is called the first-level
	// fan-out table.

	// for ex: 0x04th = 40 fanout means 40 enteries have
	// first byte of sha <= 0x04
	// Total enteries = last entry

	// TODO: use this to verify if the idx file is correct
	fanoutTable := make([]uint32, _FanoutTableLen)

	var prevCount uint32

	for idx := range _FanoutTableLen {
		count := binary.BigEndian.Uint32(fanoutTableBytes[idx*4 : (idx+1)*4])

		fanoutTable[idx] = count - prevCount

		prevCount = count
	}

	totalEnteries := prevCount

	shaList := make([]*sha.SHA, totalEnteries)

	// A table of sorted object names. These are packed together without offset
	// values to reduce the cache footprint of the binary search for a specific object name.
	shaTable := idxBytes[:totalEnteries*sha.BYTES_LEN]

	idxBytes = idxBytes[totalEnteries*sha.BYTES_LEN:]

	for idx := range totalEnteries {
		sl := shaTable[idx*sha.BYTES_LEN : (idx+1)*sha.BYTES_LEN]
		s, err := sha.FromByteSlice(&sl)

		if err != nil {
			return nil, err
		}

		shaList[idx] = s
	}

	// Next is CRC, which we are skipping for now

	// A table of 4-byte CRC32 values of the packed object data.
	// This is new in v2 so compressed data can be copied directly
	// from pack to pack during repacking without undetected
	// data corruption.

	// TODO: Verify CRC checks

	idxBytes = idxBytes[totalEnteries*4:]

	offsetBytes := idxBytes[:totalEnteries*4]

	// Verify Trailer bytes
	// TODO: verify index file checksum properly

	trailerBytes := idxBytes[totalEnteries*4:]

	// A copy of the pack checksum at the end of corresponding packfile.
	// Index checksum of all of the above.
	if len(trailerBytes) != sha.BYTES_LEN*2 {
		return nil, ErrIndexParsing
	}

	// A table of 4-byte offset values (in network byte order).
	// These are usually 31-bit pack file offsets, but large
	// offsets are encoded as an index into the next table with the msbit set.

	offsetMap := make(map[string]idxItem, totalEnteries)

	// FIXME: This doesn't support files larger than 2GB, but ig thats fine.
	// TODO: add tests for larger pack files.
	for idx := range totalEnteries {
		sl := offsetBytes[idx*4 : (idx+1)*4]
		offset := binary.BigEndian.Uint32(sl)

		// TODO: Do we need compressed size?
		// Offsets are in sorted SHA order and
		// NOT sorted by offset.

		// var compressedSize uint32

		// if idx < totalEnteries-1 {

		// 	nextOffset := binary.BigEndian.Uint32(offsetBytes[(idx+1)*4 : (idx+2)*4])

		// 	compressedSize = nextOffset - offset
		// }

		offsetMap[shaList[idx].String()] = idxItem{
			Offset:         offset,
			CompressedSize: 0,
		}
	}

	return &PackIndex{
		offsetMap:   offsetMap,
		offsetOrder: shaList,
	}, nil

}
