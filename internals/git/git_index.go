package git

import (
	"crypto/sha1"
	"fmt"
	"os"
	"path"
	"sort"
	"strconv"

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

var _TreeExtensionHeader [4]byte = [4]byte{0x54, 0x52, 0x45, 0x45} // TREE
const _TreeExtensionSize int = 4                                   // 4 bytes reserved for tree size

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

func newIndexEntry(entry *[]byte, start, end int) (*IndexEntry, error) {
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

type CacheTree struct {
	relPath       string
	entryCount    int
	subTrees      []*CacheTree
	sha           *SHA
	subTreesCount int
}

func parseCacheTreeArrToTree(cacheTreeArr *[]*CacheTree, startIdx int) (*CacheTree, int, error) {
	if startIdx > len(*cacheTreeArr) {
		return nil, 0, fmt.Errorf("invalid index %d", startIdx)
	}

	cacheTree := (*cacheTreeArr)[startIdx]

	subTrees := make([]*CacheTree, 0, cacheTree.subTreesCount)

	idx := startIdx + 1

	for range cacheTree.subTreesCount {
		subTree, newIdx, err := parseCacheTreeArrToTree(cacheTreeArr, idx)

		if err != nil {
			return nil, 0, err
		}

		subTrees = append(subTrees, subTree)

		idx = newIdx
	}

	cacheTree.subTrees = subTrees

	return cacheTree, idx, nil
}

func newCacheTree(treeContents *[]byte) (*CacheTree, error) {
	// tree, _, err := parseCacheTreeEntry(treeContents, 1)

	var cacheTrees []*CacheTree

	for idx := 0; idx < len(*treeContents)-SHA_BYTES_LEN; {
		startIdx := idx

		for ; (*treeContents)[idx] != 0x00; idx++ {
		}

		relPath := string((*treeContents)[startIdx:idx])

		idx++

		startIdx = idx

		for ; (*treeContents)[idx] != ' '; idx++ {
		}

		entryCount, err := strconv.Atoi(string((*treeContents)[startIdx:idx]))

		if err != nil {
			return nil, err
		}

		idx++

		startIdx = idx

		for ; (*treeContents)[idx] != '\n'; idx++ {
		}

		subTreeCount, err := strconv.Atoi(string((*treeContents)[startIdx:idx]))

		if err != nil {
			return nil, err
		}

		idx++

		shaSlice := (*treeContents)[idx : idx+SHA_BYTES_LEN]

		idx += SHA_BYTES_LEN

		sha, err := SHAFromByteSlice(&shaSlice)

		if err != nil {
			return nil, err
		}

		cacheTree := &CacheTree{
			relPath:       relPath,
			entryCount:    entryCount,
			sha:           sha,
			subTreesCount: subTreeCount,
		}

		cacheTrees = append(cacheTrees, cacheTree)
	}

	cacheTree, _, err := parseCacheTreeArrToTree(&cacheTrees, 0)

	return cacheTree, err
}

// @see https://git-scm.com/docs/index-format
type Index struct {
	fileMap   map[string]*IndexEntry
	cacheTree *CacheTree
	sha       *SHA
}

var ErrInvalidIndex = fmt.Errorf("invalid index file")
var ErrVersionNotSupported = fmt.Errorf("index file version not supported")
var ErrCorruptedIndex = fmt.Errorf("index file corrupted")

func byteSliceToInt(bytesSlice *[]byte) (int64, error) {
	return strconv.ParseInt(fmt.Sprintf("%x", *bytesSlice), 16, 64)
}

func verifyIndexFile(fileContents *[]byte) error {
	if len(*fileContents) < (len(_IndexFileHeader) + len(_IndexFileSupportedVersion)) {
		return ErrInvalidIndex
	}

	// Confirm header and version are correct

	headerBytes := (*fileContents)[:len(_IndexFileHeader)]
	fileVersionBytes := (*fileContents)[len(_IndexFileHeader) : len(_IndexFileHeader)+len(_IndexFileSupportedVersion)]

	for idx, b := range _IndexFileHeader {
		if headerBytes[idx] != b {
			return ErrInvalidIndex
		}
	}

	for idx, b := range _IndexFileSupportedVersion {
		if fileVersionBytes[idx] != b {
			return ErrVersionNotSupported
		}
	}

	shaSlice := (*fileContents)[len((*fileContents))-SHA_BYTES_LEN:]

	sha, err := SHAFromByteSlice(&shaSlice)

	if err != nil {
		return err
	}

	hashableBytes := (*fileContents)[:len((*fileContents))-SHA_BYTES_LEN]

	fileShaBytes := sha1.Sum(hashableBytes)

	fileShaSlice := fileShaBytes[:]

	fileSha, err := SHAFromByteSlice(&fileShaSlice)

	if err != nil {
		return err
	}

	if !sha.Eq(fileSha) {
		return ErrCorruptedIndex
	}

	return nil
}

func NewIndex() (*Index, error) {
	gitDir, err := internals.GetGitDir()

	if err != nil {
		return nil, err
	}

	fileContents, err := os.ReadFile(path.Join(gitDir, _IndexFileName))

	if err != nil {
		return nil, err
	}

	err = verifyIndexFile(&fileContents)

	if err != nil {
		return nil, err
	}

	numFilesBytes := fileContents[len(_IndexFileHeader)+len(_IndexFileSupportedVersion) : len(_IndexFileHeader)+len(_IndexFileSupportedVersion)+_NumFilesBytesLen]

	numFiles, err := byteSliceToInt(&numFilesBytes)

	if err != nil {
		return nil, err
	}

	actualContentStartIdx := len(_IndexFileHeader) + len(_IndexFileSupportedVersion) + _NumFilesBytesLen

	actualContent := fileContents[actualContentStartIdx:]

	indexEnteries := make([]*IndexEntry, numFiles)

	idx := 0

	for currFileIdx := 0; currFileIdx < int(numFiles); currFileIdx++ {
		startLoc := idx

		idx += _IndexEntryMetadataLen

		// After entry name we have at least one NULL byte
		for ; actualContent[idx] != 0x00; idx++ {
		}

		indexEntry, err := newIndexEntry(&actualContent, startLoc, idx)

		if err != nil {
			return nil, err
		}

		indexEnteries[currFileIdx] = indexEntry

		// @see https://git-scm.com/docs/index-format
		// 1-8 nul bytes as necessary to pad the entry to a multiple of eight bytes
		// while keeping the name NUL-terminated.

		// We need to calculate the padding

		fileNameLen := idx - startLoc

		padding := _IndexEntryPaddingBytes - (fileNameLen % _IndexEntryPaddingBytes)

		idx += padding
	}

	treeContents := actualContent[idx:]

	startIdx := 0

	for i := range treeContents {
		isStart := true

		for j := range _TreeExtensionHeader {
			if treeContents[startIdx+j] != _TreeExtensionHeader[j] {
				isStart = false
			}
		}

		if isStart {
			startIdx = i
			break
		}
	}

	treeContents = treeContents[startIdx+len(_TreeExtensionHeader)+_TreeExtensionSize:]

	cacheTree, err := newCacheTree(&treeContents)

	if err != nil {
		return nil, err
	}

	shaSlice := treeContents[len(treeContents)-SHA_BYTES_LEN:]

	sha, err := SHAFromByteSlice(&shaSlice)

	if err != nil {
		return nil, err
	}

	indexEntryMap := make(map[string]*IndexEntry, numFiles)

	for _, indexEntry := range indexEnteries {
		indexEntryMap[indexEntry.Filepath] = indexEntry
	}

	return &Index{
		fileMap:   indexEntryMap,
		cacheTree: cacheTree,
		sha:       sha,
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

	sort.Slice(indexEnteries, func(i, j int) bool {
		return indexEnteries[i].Filepath < indexEnteries[j].Filepath
	})

	return indexEnteries
}
