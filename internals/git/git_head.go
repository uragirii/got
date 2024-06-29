package git

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/uragirii/got/internals"
)

const _HEAD_FILE string = "HEAD"
const _REF string = "ref: "

type HeadMode int

const (
	Detached HeadMode = iota
	Branch
	Tag
)

type Head struct {
	SHA  *SHA
	Mode HeadMode
}

var ErrInvalidHead = fmt.Errorf("invalid HEAD")

func parseRefHead(headContents string) (*Head, error) {
	gitDir, err := internals.GetGitDir()

	if err != nil {
		return nil, err
	}

	refPath := headContents[len(_REF):]

	shaBytes, err := os.ReadFile(path.Join(gitDir, refPath))

	if err != nil {
		return nil, err
	}

	headMode := Branch

	if strings.Contains(refPath, "tags") {
		headMode = Tag
	}

	// Need to convert to string as the SHA in hex is stored as string
	// in the file and the bytes are not ASCII for those hex characters
	sha, err := SHAFromString(string(shaBytes))

	if err != nil {
		return nil, err
	}

	return &Head{
		SHA:  sha,
		Mode: headMode,
	}, nil
}

func NewHead() (*Head, error) {
	gitDir, err := internals.GetGitDir()

	if err != nil {
		return nil, err
	}

	headFilePath := path.Join(gitDir, _HEAD_FILE)

	headByteContents, err := os.ReadFile(headFilePath)

	if err != nil {
		return nil, err
	}

	headContents := strings.Trim(string(headByteContents), "\n")

	if strings.HasPrefix(headContents, _REF) {
		return parseRefHead(headContents)
	}

	// Detached head mode
	if len(headContents) != 20 {
		return nil, ErrInvalidHead
	} else {
		sha, err := SHAFromString(headContents)

		if err != nil {
			return nil, err
		}

		return &Head{
			SHA:  sha,
			Mode: Detached,
		}, nil
	}

}
