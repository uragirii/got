package git

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"sync"

	"github.com/uragirii/got/internals"
)

const _IndexFileName string = "index"
const _NumFilesBytesLen int = 4

// File metadata like ctime, mtime etc
const _IndexEntryMetadataLen int = 62

const _IndexEntrySizeLoc int = 36
const _IndexEntrySHALoc int = _IndexEntrySizeLoc + 4
const _IndexEntryNameLoc int = _IndexEntrySHALoc + SHA_BYTES_LEN + 2 // 2 bytes for 16 bits flags
const _IndexEntryPaddingBytes int = 8

var _IndexFileHeader [4]byte = [4]byte{0x44, 0x49, 0x52, 0x43}           // DIRC
var _IndexFileSupportedVersion [4]byte = [4]byte{0x00, 0x00, 0x00, 0x02} // Version 2

type IndexEntry struct {

	// ctime    uint64
	// mtime    uint64
	// devId    uint64
	// inode    uint64
	// mode     uint32
	// uid      uint32
	// gid      uint32
	Size uint32
	SHA  *SHA
	// flag     uint16
	Filepath string
}

func unmarshalIndexEntry(entry *[]byte, start, end int) (*IndexEntry, error) {
	sizeBytes := (*entry)[start+_IndexEntrySizeLoc : start+_IndexEntrySizeLoc+4]

	shaBytes := (*entry)[start+_IndexEntrySHALoc : start+_IndexEntrySHALoc+20] // SHA is 20 bytes

	filepath := (*entry)[start+_IndexEntryNameLoc : end]

	size, err := byteSliceToInt(&sizeBytes)

	if err != nil {
		return nil, err
	}

	sha, err := SHAFromByteSlice(&shaBytes)
	if err != nil {
		return nil, err
	}

	return &IndexEntry{
		Size:     uint32(size),
		SHA:      sha,
		Filepath: string(filepath),
	}, nil

}

// @see https://git-scm.com/docs/index-format
type Index struct {
	fileMap map[string]*IndexEntry
}

var ErrInvalidIndex = fmt.Errorf("invalid index file")
var ErrVersionNotSupported = fmt.Errorf("index file version not supported")

func byteSliceToInt(bytesSlice *[]byte) (int64, error) {
	return strconv.ParseInt(fmt.Sprintf("%x", *bytesSlice), 16, 64)
}

func UnmarshallGitIndex() (*Index, error) {
	gitDir, err := internals.GetGitDir()

	if err != nil {
		return nil, err
	}

	fileContents, err := os.ReadFile(path.Join(gitDir, _IndexFileName))

	if err != nil {
		return nil, err
	}

	if len(fileContents) < (len(_IndexFileHeader) + len(_IndexFileSupportedVersion)) {
		return nil, ErrInvalidIndex
	}

	// Confirm header and version are correct

	headerBytes := fileContents[:len(_IndexFileHeader)]
	fileVersionBytes := fileContents[len(_IndexFileHeader) : len(_IndexFileHeader)+len(_IndexFileSupportedVersion)]

	for idx, b := range _IndexFileHeader {
		if headerBytes[idx] != b {
			return nil, ErrInvalidIndex
		}
	}

	for idx, b := range _IndexFileSupportedVersion {
		if fileVersionBytes[idx] != b {
			return nil, ErrVersionNotSupported
		}
	}

	numFilesBytes := fileContents[len(_IndexFileHeader)+len(_IndexFileSupportedVersion) : len(_IndexFileHeader)+len(_IndexFileSupportedVersion)+_NumFilesBytesLen]

	numFiles, err := byteSliceToInt(&numFilesBytes)

	if err != nil {
		return nil, err
	}

	actualContentStartIdx := len(_IndexFileHeader) + len(_IndexFileSupportedVersion) + _NumFilesBytesLen

	actualContent := fileContents[actualContentStartIdx:]

	indexEnteries := make([]*IndexEntry, numFiles)

	var wg sync.WaitGroup

	idx := 0

	for currFileIdx := 0; currFileIdx < int(numFiles); currFileIdx++ {
		startLoc := idx

		idx += _IndexEntryMetadataLen

		// After entry name we have at least one NULL byte
		for ; actualContent[idx] != 0x00; idx++ {
		}

		wg.Add(1)

		go func(currFileIdx, idx int) {
			defer wg.Done()

			indexEntry, err := unmarshalIndexEntry(&actualContent, startLoc, idx)

			// TODO: handle errors inside goroutines
			if err != nil {
				fmt.Println("WARN need to better handle errs")
				panic(err)
			}

			indexEnteries[currFileIdx] = indexEntry

		}(currFileIdx, idx)

		// @see https://git-scm.com/docs/index-format
		// 1-8 nul bytes as necessary to pad the entry to a multiple of eight bytes
		// while keeping the name NUL-terminated.

		// We need to calculate the padding

		fileNameLen := idx - startLoc

		padding := _IndexEntryPaddingBytes - (fileNameLen % _IndexEntryPaddingBytes)

		idx += padding
	}

	wg.Wait()

	indexEntryMap := make(map[string]*IndexEntry, numFiles)

	for _, indexEntry := range indexEnteries {
		indexEntryMap[indexEntry.Filepath] = indexEntry
	}

	return &Index{
		fileMap: indexEntryMap,
	}, nil

}

func (i *Index) Has(filePath string) bool {
	_, ok := i.fileMap[filePath]

	return ok
}

func (i *Index) Get(filePath string) *IndexEntry {
	return i.fileMap[filePath]
}

func (i *Index) GetTrackedFiles() []*IndexEntry {
	indexEnteries := make([]*IndexEntry, 0, len(i.fileMap))

	for _, entry := range i.fileMap {
		indexEnteries = append(indexEnteries, entry)
	}

	return indexEnteries
}
