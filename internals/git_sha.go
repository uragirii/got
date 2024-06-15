package internals

import (
	"encoding/hex"
	"fmt"
)

type SHA struct {
	hash *[]byte
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

func (sha *SHA) UnmarshallToStr() string {
	return fmt.Sprintf("%x", *sha.hash)
}

func (sha *SHA) GetBytes() *[]byte {
	return sha.hash
}
