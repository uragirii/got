package index

import (
	"encoding/hex"
	"fmt"
	"io"
	"strconv"
)

func parse32bit(data *[]byte, startIdx int) (uint32, error) {
	num, err := strconv.ParseUint(fmt.Sprintf("%x", (*data)[startIdx:startIdx+_32BitToByte]), 16, 32)

	return uint32(num), err
}

func writeUint32(num uint32, writer io.Writer) (int, error) {
	bytes, _ := hex.DecodeString(fmt.Sprintf("%08x", num))

	return writer.Write(bytes)
}

func byteSliceToInt(bytesSlice *[]byte) (int64, error) {
	return strconv.ParseInt(fmt.Sprintf("%x", *bytesSlice), 16, 64)
}
