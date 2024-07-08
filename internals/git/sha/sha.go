package sha

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"
)

const BYTES_LEN = 20
const STR_LEN = BYTES_LEN * 2

type SHA struct {
	hash *[]byte
}

func (sha *SHA) Eq(other *SHA) bool {
	for i, b := range *sha.hash {
		if (*other.hash)[i] != b {
			return false
		}
	}

	return true
}

func FromByteSlice(byteSlice *[]byte) (*SHA, error) {

	trimmedBytes := bytes.Trim(*byteSlice, "\n")

	return &SHA{
		hash: &trimmedBytes,
	}, nil
}

func FromString(shaStr string) (*SHA, error) {
	byteSlice, err := hex.DecodeString(strings.Trim(shaStr, "\n"))

	if err != nil {
		return nil, err
	}

	return FromByteSlice(&byteSlice)
}

func (sha *SHA) MarshallToStr() string {
	return fmt.Sprintf("%x", *sha.hash)
}

func (sha *SHA) GetBytes() *[]byte {
	return sha.hash
}

func (sha *SHA) GetBinStr() string {
	return string(*sha.hash)
}
