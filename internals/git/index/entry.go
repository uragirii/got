package index

import (
	"encoding/hex"
	"fmt"
	"io"
	"strconv"
	"strings"
	"syscall"

	"github.com/uragirii/got/internals/git/sha"
)

type IndexEntry struct {
	ctime    syscall.Timespec
	mtime    syscall.Timespec
	devId    uint32
	inode    uint32
	mode     *mode
	uid      uint32
	gid      uint32
	Size     uint32
	SHA      *sha.SHA
	flag     uint64
	Filepath string
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

	modeBits, err := parse32bit(entry, start)

	if err != nil {
		return nil, err
	}

	mode, err := modeFromUint32(modeBits)

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

	shaBytes := (*entry)[start : start+sha.BYTES_LEN]

	start += sha.BYTES_LEN

	flag, err := strconv.ParseUint(fmt.Sprintf("%x", (*entry)[start:start+2]), 16, 16)

	if err != nil {
		return nil, err
	}

	// last 12 bit are filepath len, choosing to ignore that
	// actual flag is first 4 bit
	// 1-bit assume-valid flag
	// 1-bit extended flag (must be zero in version 2)
	// 2-bit stage (during merge)

	flag = flag >> 12

	start += 2

	filepath := (*entry)[start:end]

	sha, err := sha.FromByteSlice(&shaBytes)
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

	n, _ = entry.mode.Write(writer)
	bytesWritten += n

	n, _ = writeUint32(entry.uid, writer)
	bytesWritten += n
	n, _ = writeUint32(entry.gid, writer)
	bytesWritten += n

	n, _ = writeUint32(entry.Size, writer)

	bytesWritten += n

	n, _ = writer.Write(*entry.SHA.GetBytes())

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

func (entry IndexEntry) Debug() string {
	var sb strings.Builder

	sb.WriteString(entry.Filepath)
	sb.WriteRune('\n')

	sb.WriteString(fmt.Sprintf("  ctime: %d:%d\n", entry.ctime.Sec, entry.ctime.Nsec))
	sb.WriteString(fmt.Sprintf("  mtime: %d:%d\n", entry.ctime.Sec, entry.ctime.Nsec))
	sb.WriteString(fmt.Sprintf("  dev: %d\tino: %d\n", entry.devId, entry.inode))
	sb.WriteString(fmt.Sprintf("  uid: %d\tgid: %d\n", entry.uid, entry.gid))
	sb.WriteString(fmt.Sprintf("  size: %d\tflags: %d\n", entry.Size, entry.flag))

	return sb.String()
}
