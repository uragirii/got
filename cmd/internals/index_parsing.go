package internals

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"sync"
)

var INDEX_FILE_HEADER [4]byte = [4]byte{0x44, 0x49, 0x52, 0x43}
var SUPPORTED_VERSION [4]byte = [4]byte{0x00, 0x00, 0x00, 0x02}

type IndexFile struct {
	// ctime    uint64
	// mtime    uint64
	// devId    uint64
	// inode    uint64
	// mode     uint32
	// uid      uint32
	// gid      uint32
	Size uint32
	SHA1 *[20]byte
	// flag     uint16
	Filepath string
	Start    int
	End      int
}

func byteSliceToInt(bytesSlice *[]byte) (int64, error) {
	return strconv.ParseInt(fmt.Sprintf("%x", *bytesSlice), 16, 64)
}

func parseIndexEntry(entry *[]byte, start, end int) (*IndexFile, error) {
	sizeBytes := (*entry)[start+36 : start+40]
	shaBytes := (*entry)[start+40 : start+60] // SHA1 is 20 bytes
	filepath := (*entry)[start+62 : end]

	size, err := byteSliceToInt(&sizeBytes)

	if err != nil {
		return nil, err
	}

	sha1 := [20]byte{}

	for idx := range len(sha1) {
		sha1[idx] = shaBytes[idx]
	}

	return &IndexFile{
		Size:     uint32(size),
		SHA1:     &sha1,
		Filepath: string(filepath),
		Start:    start,
		End:      end,
	}, nil

}

type GitIndex struct {
	entries []*IndexFile
	fileMap map[string]*IndexFile
}

func (i *GitIndex) New(gitDir string) error {
	contents, err := os.ReadFile(path.Join(gitDir, "index"))

	if err != nil {
		return err
	}

	if len(contents) == 0 {
		return fmt.Errorf("empty index file")
	}

	if len(contents) < 8 {
		return fmt.Errorf("file has invalid length")
	}

	header := contents[0:4]
	fileversion := contents[4:8]
	numFilesBytes := contents[8:12]

	for idx, b := range INDEX_FILE_HEADER {
		if header[idx] != b {
			return fmt.Errorf("invalid header")
		}
	}

	for idx, b := range SUPPORTED_VERSION {
		if fileversion[idx] != b {
			return fmt.Errorf("%d version not supported", fileversion)
		}
	}

	numFiles, err := byteSliceToInt(&numFilesBytes)

	if err != nil {
		return err
	}

	fileContentsBytes := contents[12:]

	indexEntryLocs := make([]int, numFiles)

	fileEntry := make([]*IndexFile, numFiles)

	currLoc := 0

	i.fileMap = make(map[string]*IndexFile, numFiles)

	var wg sync.WaitGroup

	for currIdx := 0; currIdx < int(numFiles); currIdx++ {
		// Other stuff is 62 bytes long
		startLoc := currLoc
		currLoc += 62

		for ; fileContentsBytes[currLoc] != 0; currLoc++ {
		}
		wg.Add(1)

		go func(entry *[]byte, start, end, currIdx int) {
			defer wg.Done()
			indexFile, err := parseIndexEntry(entry, start, end)

			if err != nil {
				panic(err)
			}

			fileEntry[currIdx] = indexFile
			i.fileMap[indexFile.Filepath] = indexFile

		}(&fileContentsBytes, startLoc, currLoc, currIdx)

		// @see https://git-scm.com/docs/index-format
		// 1-8 nul bytes as necessary to pad the entry to a multiple of eight bytes
		// while keeping the name NUL-terminated.

		// We need to calculate the padding

		filenameLenBytes := currLoc - startLoc

		padding := 8 - (filenameLenBytes % 8)

		currLoc += padding

		indexEntryLocs[currIdx] = currLoc

	}

	wg.Wait()

	i.entries = fileEntry

	return nil
}

func (i *GitIndex) Has(filePath string) bool {
	return i.fileMap[filePath] != nil
}

func (i *GitIndex) Get(filePath string) *IndexFile {
	return i.fileMap[filePath]
}

func (i *GitIndex) GetTrackedFiles() []*IndexFile {
	return i.entries
}
