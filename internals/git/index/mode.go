package index

import (
	"io"
	"io/fs"
)

// binary are 1000 (regular file), 1010 (symbolic link) and 1110 (gitlink)
type modeType uint8

const (
	modeRegular modeType = 0b1000
	modeSymLink modeType = 0b1010
	modeGitLink modeType = 0b1110
)

type mode struct {
	// 4-bit object type
	fileType modeType
	// 9-bit unix permission. Only 0755 and 0644 are valid for regular files.
	// Symbolic links and gitlinks have value 0 in this field.
	perm uint32
}

func modeFromUint32(m uint32) (*mode, error) {

	permBits := m & 0b1_1111_1111

	if permBits != 0755 && permBits != 0644 && permBits != 0 {
		return nil, ErrInvalidEntryMode
	}

	fileType := modeType(m >> 12)

	mode := &mode{
		fileType: fileType,
		perm:     uint32(permBits),
	}

	return mode, nil
}

func modeFromFilePath(filePath string, fsys fs.FS) (*mode, error) {
	stat, err := fs.Stat(fsys, filePath)

	if err != nil {
		return nil, err
	}

	// fixme
	// TODO: check if file is symlink or git link
	return &mode{
		fileType: modeRegular,
		perm:     uint32(stat.Mode().Perm()),
	}, nil

}

func (m mode) Write(writer io.Writer) (int, error) {
	var num uint32 = 0

	num |= uint32(m.fileType)

	num = num << 12

	num |= uint32(m.perm)

	return writeUint32(num, writer)
}
