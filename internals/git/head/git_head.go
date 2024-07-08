package head

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/uragirii/got/internals"
	"github.com/uragirii/got/internals/git/sha"
)

const _HeadFile string = "HEAD"
const _Ref string = "ref: "
const _BranchPrefix = "refs/heads/"

type Mode int

const (
	Detached Mode = iota
	Branch
	Tag
)

type Head struct {
	SHA    *sha.SHA
	Mode   Mode
	Branch string
}

var ErrInvalidHead = fmt.Errorf("invalid HEAD")

func parseRefHead(headContents string) (*Head, error) {
	gitDir, err := internals.GetGitDir()

	if err != nil {
		return nil, err
	}

	refPath := headContents[len(_Ref):]

	shaBytes, err := os.ReadFile(path.Join(gitDir, refPath))

	if err != nil {
		return nil, err
	}

	var branch string

	headMode := Tag

	if strings.HasPrefix(refPath, _BranchPrefix) {
		headMode = Branch
		branch = refPath[len(_BranchPrefix):]
	}

	// Need to convert to string as the SHA in hex is stored as string
	// in the file and the bytes are not ASCII for those hex characters
	sha, err := sha.FromString(string(shaBytes))

	if err != nil {
		return nil, err
	}

	return &Head{
		SHA:    sha,
		Mode:   headMode,
		Branch: branch,
	}, nil
}

func New() (*Head, error) {
	gitDir, err := internals.GetGitDir()

	if err != nil {
		return nil, err
	}

	headFilePath := path.Join(gitDir, _HeadFile)

	headByteContents, err := os.ReadFile(headFilePath)

	if err != nil {
		return nil, err
	}

	headContents := strings.Trim(string(headByteContents), "\n")

	if strings.HasPrefix(headContents, _Ref) {
		return parseRefHead(headContents)
	}

	// Detached head mode
	if len(headContents) != 20 {
		return nil, ErrInvalidHead
	} else {
		sha, err := sha.FromString(headContents)

		if err != nil {
			return nil, err
		}

		return &Head{
			SHA:  sha,
			Mode: Detached,
		}, nil
	}

}