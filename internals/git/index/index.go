package index

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"syscall"

	"github.com/uragirii/got/internals"
	"github.com/uragirii/got/internals/git/object"
	"github.com/uragirii/got/internals/git/sha"
)

// @see https://git-scm.com/docs/index-format
type Index struct {
	fileMap   map[string]*IndexEntry
	cacheTree *CacheTree
	sha       *sha.SHA
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

	shaSlice := (*fileContents)[len((*fileContents))-sha.BYTES_LEN:]

	sHA, err := sha.FromByteSlice(&shaSlice)

	if err != nil {
		return err
	}

	hashableBytes := (*fileContents)[:len((*fileContents))-sha.BYTES_LEN]

	fileShaBytes := sha1.Sum(hashableBytes)

	fileShaSlice := fileShaBytes[:]

	fileSha, err := sha.FromByteSlice(&fileShaSlice)

	if err != nil {
		return err
	}

	if !sHA.Eq(fileSha) {
		return ErrCorruptedIndex
	}

	return nil
}

func New() (*Index, error) {
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

	shaSlice := treeContents[len(treeContents)-sha.BYTES_LEN:]

	sha, err := sha.FromByteSlice(&shaSlice)

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

// Adds the files to the index
// provide rel Path for the file and not the matching pattern
func (i *Index) Add(filePaths []string) error {
	for _, filePath := range filePaths {
		obj, err := object.NewObject(filePath)

		if err != nil {
			return err
		}

		var fileStat syscall.Stat_t

		if err = syscall.Stat(filePath, &fileStat); err != nil {
			return err
		}

		mode, err := modeFromFilePath(filePath)

		if err != nil {
			return err
		}

		i.fileMap[filePath] = &IndexEntry{
			ctime: fileStat.Ctimespec,
			mtime: fileStat.Mtimespec,
			mode:  mode,
			devId: uint32(fileStat.Dev),
			uid:   uint32(fileStat.Uid),
			gid:   uint32(fileStat.Gid),
			inode: uint32(fileStat.Ino),
			// fixme
			// TODO: as other flag is always unset, this should be fine
			flag:     uint64(len(filePath)),
			Size:     uint32(fileStat.Size),
			SHA:      obj.GetSHA(),
			Filepath: filePath,
		}

		i.cacheTree.add(strings.Split(filepath.Dir(filePath), string(filepath.Separator)))
	}

	fmt.Println(i.cacheTree)

	return nil
}
