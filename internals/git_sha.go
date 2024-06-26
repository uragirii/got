package internals

import (
	"encoding/hex"
	"fmt"
)

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

	return &SHA{
		hash: byteSlice,
	}, nil
}

func SHAFromString(shaStr string) (*SHA, error) {
	byteSlice, err := hex.DecodeString(shaStr)

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
