package sha

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"path"
	"strings"
)

const BYTES_LEN = 20
const STR_LEN = BYTES_LEN * 2
const _ObjectsDir string = "objects"

type SHA struct {
	hash *[]byte
}

func (sha *SHA) Eq(other *SHA) bool {
	return bytes.Equal(*sha.hash, *other.hash)
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

func (sha *SHA) String() string {
	return fmt.Sprintf("%x", *sha.hash)
}

func (sha *SHA) GetBytes() *[]byte {
	return sha.hash
}

func (sha *SHA) GetBinStr() string {
	return string(*sha.hash)
}

func (sha SHA) GetObjPath() (string, error) {

	shaStr := sha.String()

	objPath := path.Join(_ObjectsDir, shaStr[0:2], shaStr[2:])

	return objPath, nil
}

func FromData(data *[]byte) (*SHA, error) {
	hash := sha1.Sum(*data)

	hashSlice := hash[:]

	return FromByteSlice(&hashSlice)
}
