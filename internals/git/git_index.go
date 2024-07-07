package git

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strconv"
	"syscall"

	"github.com/uragirii/got/internals"
)

const _IndexFileName string = "index"
const _NumFilesBytesLen int = 4

// File metadata like ctime, mtime etc
const _IndexEntryMetadataLen int = 62

const _IndexEntryPaddingBytes int = 8
const _32BitToByte = 32 / 8

var _IndexFileHeader [4]byte = [4]byte{0x44, 0x49, 0x52, 0x43}           // DIRC
var _IndexFileSupportedVersion [4]byte = [4]byte{0x00, 0x00, 0x00, 0x02} // Version 2

var _TreeExtensionHeader [4]byte = [4]byte{0x54, 0x52, 0x45, 0x45} // TREE
const _TreeExtensionSize int = 4                                   // 4 bytes reserved for tree size

type IndexEntry struct {
	ctime    syscall.Timespec
	mtime    syscall.Timespec
	devId    uint32
	inode    uint32
	mode     uint32
	uid      uint32
	gid      uint32
	Size     uint32
	SHA      *SHA
	flag     uint64
	Filepath string
}

func writeUint32(num uint32, writer io.Writer) (int, error) {
	bytes, _ := hex.DecodeString(fmt.Sprintf("%08x", num))

	return writer.Write(bytes)
}

func (entry IndexEntry) Write(writer io.Writer) (int, error) {
	bytesWritten := 0
	n, _ := writeUint32(uint32(entry.ctime.Sec), writer)
	bytesWritten += n
	n, _ = writeUint32(uint32(entry.ctime.Nsec), writer)
	bytesWritten += n
	n, _ = writeUint32(uint32(entry.mtime.Sec), writer)
	bytesWritten += n
	n, _ = writeUint32(uint32(entry.mtime.Nsec), writer)
	bytesWritten += n

	n, _ = writeUint32(entry.devId, writer)
	bytesWritten += n
	n, _ = writeUint32(entry.inode, writer)
	bytesWritten += n
	n, _ = writeUint32(entry.mode, writer)
	bytesWritten += n

	n, _ = writeUint32(entry.uid, writer)
	bytesWritten += n
	n, _ = writeUint32(entry.gid, writer)
	bytesWritten += n

	n, _ = writeUint32(entry.Size, writer)

	bytesWritten += n

	n, _ = writer.Write(*entry.SHA.hash)

	bytesWritten += n

	flagBytes, _ := hex.DecodeString(fmt.Sprintf("%04x", entry.flag))

	n, _ = writer.Write(flagBytes)
	bytesWritten += n

	n, _ = writer.Write([]byte(entry.Filepath))
	bytesWritten += n

	// The NULL byte is already accumulated inside the padding
	// as minimum padding is 1 and max is 8

	padding := _IndexEntryPaddingBytes - (bytesWritten % _IndexEntryPaddingBytes)

	paddingSlice := make([]byte, padding)

	n, _ = writer.Write(paddingSlice)

	return bytesWritten + n, nil
}

func parse32bit(data *[]byte, startIdx int) (uint32, error) {
	num, err := strconv.ParseUint(fmt.Sprintf("%x", (*data)[startIdx:startIdx+_32BitToByte]), 16, 32)

	return uint32(num), err
}

func newIndexEntry(entry *[]byte, start, end int) (*IndexEntry, error) {
	ctimeSec, err := parse32bit(entry, start)

	if err != nil {
		return nil, err
	}

	start += _32BitToByte

	ctimeNanoSec, err := parse32bit(entry, start)

	if err != nil {
		return nil, err
	}
	start += _32BitToByte

	cTime := syscall.Timespec{
		Sec:  int64(ctimeSec),
		Nsec: int64(ctimeNanoSec),
	}

	mtimeSec, err := parse32bit(entry, start)

	if err != nil {
		return nil, err
	}
	start += _32BitToByte

	mtimeNanoSec, err := parse32bit(entry, start)

	if err != nil {
		return nil, err
	}

	start += _32BitToByte

	mTime := syscall.Timespec{
		Sec:  int64(mtimeSec),
		Nsec: int64(mtimeNanoSec),
	}

	dev, err := parse32bit(entry, start)

	if err != nil {
		return nil, err
	}

	start += _32BitToByte

	ino, err := parse32bit(entry, start)

	if err != nil {
		return nil, err
	}

	start += _32BitToByte

	mode, err := parse32bit(entry, start)

	if err != nil {
		return nil, err
	}

	start += _32BitToByte

	uid, err := parse32bit(entry, start)

	if err != nil {
		return nil, err
	}

	start += _32BitToByte

	gid, err := parse32bit(entry, start)

	if err != nil {
		return nil, err
	}

	start += _32BitToByte

	size, err := parse32bit(entry, start)

	if err != nil {
		return nil, err
	}

	start += _32BitToByte

	shaBytes := (*entry)[start : start+SHA_BYTES_LEN]

	start += SHA_BYTES_LEN

	flag, err := strconv.ParseUint(fmt.Sprintf("%x", (*entry)[start:start+2]), 16, 16)

	if err != nil {
		return nil, err
	}

	start += 2

	filepath := (*entry)[start:end]

	sha, err := SHAFromByteSlice(&shaBytes)
	if err != nil {
		return nil, err
	}

	return &IndexEntry{
		Size:     uint32(size),
		SHA:      sha,
		Filepath: string(filepath),
		ctime:    cTime,
		mtime:    mTime,
		devId:    dev,
		inode:    ino,
		mode:     mode,
		uid:      uid,
		gid:      gid,
		flag:     flag,
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

func (tree CacheTree) Write(writer io.Writer) (int, error) {
	bytesWritten := 0
	n, _ := writer.Write([]byte(tree.relPath))
	bytesWritten += n
	n, _ = writer.Write([]byte{0x00})
	bytesWritten += n

	n, _ = writer.Write([]byte(fmt.Sprintf("%d", tree.entryCount)))
	bytesWritten += n

	n, _ = writer.Write([]byte{' '})
	bytesWritten += n

	n, _ = writer.Write([]byte(fmt.Sprintf("%d", tree.subTreesCount)))
	bytesWritten += n

	n, _ = writer.Write([]byte{'\n'})
	bytesWritten += n

	n, _ = writer.Write(*tree.sha.hash)
	bytesWritten += n

	for _, subTree := range tree.subTrees {
		n, _ = subTree.Write(writer)
		bytesWritten += n
	}

	return bytesWritten, nil

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

func (i *Index) Write() error {
	var buffer bytes.Buffer

	buffer.Write(_IndexFileHeader[:])

	buffer.Write(_IndexFileSupportedVersion[:])

	fileEntries := i.GetTrackedFiles()

	lenFileEntries := len(fileEntries)

	lenFileEntriesBytes, _ := hex.DecodeString(fmt.Sprintf("%08x", lenFileEntries))

	buffer.Write(lenFileEntriesBytes)

	for _, entry := range fileEntries {
		entry.Write(&buffer)
	}

	buffer.Write(_TreeExtensionHeader[:])

	var cacheTreeBuffer bytes.Buffer

	i.cacheTree.Write(&cacheTreeBuffer)

	writeUint32(uint32(cacheTreeBuffer.Len()), &buffer)

	buffer.Write(cacheTreeBuffer.Bytes())

	indexBytes := buffer.Bytes()

	sha := sha1.Sum(indexBytes)

	fi, err := os.Create("index")

	if err != nil {
		return err
	}

	defer fi.Close()

	_, err = fi.Write(indexBytes)

	if err != nil {
		return err
	}

	_, err = fi.Write(sha[:])

	if err != nil {
		return err
	}

	return nil

}
