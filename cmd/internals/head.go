package internals

import (
	"fmt"
	"os"
	"path"
	"strings"
)

const HEAD_FILE string = "HEAD"
const REF string = "ref: "

type headMode int

const (
	Detached headMode = iota
	Branch
	Tag
)

type Head struct {
	SHA  string
	Mode headMode
}

func GetHeadSHA() (*Head, error) {
	gitDir, err := GetGitDir()

	if err != nil {
		return nil, err
	}

	headFilePath := path.Join(gitDir, HEAD_FILE)

	headByteContents, err := os.ReadFile(headFilePath)

	if err != nil {
		return nil, err
	}

	headContents := strings.Trim(string(headByteContents), "\n")

	if strings.HasPrefix(headContents, REF) {
		refPath := headContents[len(REF):]

		shaBytes, err := os.ReadFile(path.Join(gitDir, refPath))

		if err != nil {
			return nil, err
		}

		headMode := Branch

		if strings.Contains(refPath, "tags") {
			headMode = Tag
		}

		return &Head{
			SHA:  strings.Trim(string(shaBytes), "\n"),
			Mode: headMode,
		}, nil
	}

	// Detached head mode
	if len(headContents) != 20 {
		return nil, fmt.Errorf("invalid head file, expected SHA found %s", headContents)
	} else {
		return &Head{
			SHA:  headContents,
			Mode: Detached,
		}, nil
	}

}
