package head

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"path"
	"strings"

	"github.com/uragirii/got/internals"
	"github.com/uragirii/got/internals/git/sha"
)

const _HeadFile string = "HEAD"

var _Ref = []byte("ref: ")

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

func newRefHead(headData []byte, fs fs.FS) (*Head, error) {
	gitDir, err := internals.GetGitDir()

	if err != nil {
		return nil, err
	}

	refPath := string(headData[len(_Ref):])

	branchFile, err := fs.Open(path.Join(gitDir, refPath))

	if err != nil {
		return nil, err
	}

	defer branchFile.Close()

	var buffer bytes.Buffer

	buffer.ReadFrom(branchFile)

	shaBytes := buffer.Bytes()

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

func newDetachedHead(headData []byte) (*Head, error) {
	if len(headData) != sha.STR_LEN {
		return nil, ErrInvalidHead
	}

	// the contents are SHA
	headSHA, err := sha.FromString(string(headData))

	if err != nil {
		return nil, err
	}

	return &Head{
		SHA:  headSHA,
		Mode: Detached,
	}, nil
}

func newHead(headFile io.Reader, fs fs.FS) (*Head, error) {
	headContents, err := io.ReadAll(headFile)

	if err != nil {
		return nil, err
	}

	headLine := bytes.TrimRight(headContents, "\n")

	// read the ref file
	if bytes.HasPrefix(headLine, _Ref) {
		return newRefHead(headLine, fs)
	}

	return newDetachedHead(headLine)

}

func New(fs fs.FS) (*Head, error) {
	gitDir, err := internals.GetGitDir()

	if err != nil {
		return nil, err
	}

	headFilePath := path.Join(gitDir, _HeadFile)

	headFile, err := fs.Open(headFilePath)

	if err != nil {
		return nil, err
	}

	defer headFile.Close()

	return newHead(headFile, fs)

}
