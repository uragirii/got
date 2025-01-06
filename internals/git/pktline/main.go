package pktline

import (
	"errors"
	"fmt"
	"strconv"
)

const FlushPacket = "0000"

var ErrInvalidPktLine = errors.New("invalid pkt line")
var ErrLengthMismatch = errors.New("pkt line didn't match")

func Encode(line string) string {
	if len(line) == 0 {
		return FlushPacket
	}

	return fmt.Sprintf("%04x%s\n", len(line)+5, line)
}

func EncodeBinary(data []byte) string {

	if len(data) == 0 {
		return FlushPacket
	}

	return fmt.Sprintf("%04x%s", len(data)+4, data)
}

func Decode(line string) (*[]byte, error) {

	if len(line) < 4 {
		return nil, ErrInvalidPktLine
	}

	if line == FlushPacket {
		empty := make([]byte, 0)
		return &empty, nil
	}

	lineLen, err := strconv.ParseInt(line[0:4], 16, 64)

	if err != nil {
		return nil, err
	}

	if len(line) != int(lineLen) {
		return nil, ErrLengthMismatch
	}

	lineBytes := []byte(line[4:])

	return &lineBytes, nil
}
