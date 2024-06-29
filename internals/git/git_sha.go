package git

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"
)

const SHA_BYTES_LEN = 20
const SHA_STR_LEN = SHA_BYTES_LEN * 2

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

func SHAFromByteSlice(byteSlice *[]byte) (*SHA, error) {

	trimmedBytes := bytes.Trim(*byteSlice, "\n")

	return &SHA{
		hash: &trimmedBytes,
	}, nil
}

func SHAFromString(shaStr string) (*SHA, error) {
	byteSlice, err := hex.DecodeString(strings.Trim(shaStr, "\n"))

	if err != nil {
		return nil, err
	}

	return SHAFromByteSlice(&byteSlice)
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
