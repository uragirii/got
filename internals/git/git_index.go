package git

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

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

type CacheTree struct {
	relPath    string
	entryCount int
	subTrees   []*CacheTree
	sha        *SHA
}

func (tree CacheTree) String() string {
	var sb strings.Builder

	sb.WriteString(tree.relPath)

	for _, subTree := range tree.subTrees {
		sb.WriteRune('\t')
		sb.WriteString(subTree.String())
	}

	sb.WriteString("\t\n")

	return sb.String()

}

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

func parseCacheTreeEntry(treeContents *[]byte) (*CacheTree, int, error) {

	fmt.Println("\t", "******")
	fmt.Println("\t", string(*treeContents))
	fmt.Println("\t", "******")
	fmt.Println()

	startIdx := 0
	idx := 0

	for ; (*treeContents)[idx] != 0x00; idx++ {
	}

	relPath := string((*treeContents)[startIdx:idx])

	idx++

	startIdx = idx

	for ; (*treeContents)[idx] != ' '; idx++ {

	}

	entryCount, err := strconv.Atoi(string((*treeContents)[startIdx:idx]))

	if err != nil {
		return nil, 0, err
	}

	idx++

	startIdx = idx

	for ; (*treeContents)[idx] != '\n'; idx++ {

	}

	subTreeCount, err := strconv.Atoi(string((*treeContents)[startIdx:idx]))
	idx++

	if err != nil {
		return nil, 0, err
	}

	shaSlice := (*treeContents)[idx : idx+SHA_BYTES_LEN]

	idx += SHA_BYTES_LEN

	sha, err := SHAFromByteSlice(&shaSlice)

	if err != nil {
		return nil, 0, err
	}

	subTrees := make([]*CacheTree, 0, subTreeCount)

	treeSlice := (*treeContents)[idx:]

	for range subTreeCount {

		subTree, endIdx, err := parseCacheTreeEntry(&treeSlice)

		if err != nil {
			return nil, 0, err
		}

		fmt.Println("------")
		fmt.Println(subTree.relPath, "-->", string(treeSlice))
		fmt.Println("------")

		idx = endIdx
		treeSlice = treeSlice[idx:]

		subTrees = append(subTrees, subTree)
	}

	fmt.Println(relPath, subTreeCount)

	return &CacheTree{
		relPath:    relPath,
		entryCount: entryCount,
		subTrees:   subTrees,
		sha:        sha,
	}, idx, nil
}

func newCacheTree(treeContents *[]byte) (*CacheTree, error) {
	tree, _, err := parseCacheTreeEntry(treeContents)

	return tree, err
}

// @see https://git-scm.com/docs/index-format
type Index struct {
	fileMap   map[string]*IndexEntry
	CacheTree *CacheTree
}

var ErrInvalidIndex = fmt.Errorf("invalid index file")
var ErrVersionNotSupported = fmt.Errorf("index file version not supported")

func byteSliceToInt(bytesSlice *[]byte) (int64, error) {
	return strconv.ParseInt(fmt.Sprintf("%x", *bytesSlice), 16, 64)
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

	indexEntryMap := make(map[string]*IndexEntry, numFiles)

	for _, indexEntry := range indexEnteries {
		indexEntryMap[indexEntry.Filepath] = indexEntry
	}

	return &Index{
		fileMap:   indexEntryMap,
		CacheTree: cacheTree,
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
