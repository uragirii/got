package internals

import (
	"fmt"
	"os"
	"strconv"
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
	// size     uint32
	// sha1     *[20]byte
	// flag     uint16
	// filepath string
}

func byteSliceToInt(bytesSlice *[]byte) (int64, error) {
	return strconv.ParseInt(fmt.Sprintf("%x", *bytesSlice), 16, 64)
}

func ParseIndex(filepath string) (*[]IndexFile, error) {
	contents, err := os.ReadFile(filepath)

	if err != nil {
		return nil, err
	}

	if len(contents) == 0 {
		return nil, fmt.Errorf("empty index file")
	}

	if len(contents) < 8 {
		return nil, fmt.Errorf("file has invalid length")
	}

	header := contents[0:4]
	fileversion := contents[4:8]
	numFilesBytes := contents[8:12]

	for idx, b := range INDEX_FILE_HEADER {
		if header[idx] != b {
			return nil, fmt.Errorf("invalid header")
		}
	}

	for idx, b := range SUPPORTED_VERSION {
		if fileversion[idx] != b {
			return nil, fmt.Errorf("%d version not supported", fileversion)
		}
	}

	numFiles, err := byteSliceToInt(&numFilesBytes)

	if err != nil {
		return nil, err
	}

	fileContentsBytes := contents[12:]

	indexEntryLocs := make([]int, numFiles)

	currLoc := 0

	for currIdx := 0; currIdx < int(numFiles); currIdx++ {
		// Other stuff is 62 bytes long
		currLoc += 62

		startLoc := currLoc

		for ; fileContentsBytes[currLoc] != 0; currLoc++ {
		}

		fmt.Printf("%s\n", fileContentsBytes[startLoc:currLoc])

		for ; fileContentsBytes[currLoc] == 0; currLoc++ {
		}
		indexEntryLocs[currIdx] = currLoc

	}

	return nil, nil

}
