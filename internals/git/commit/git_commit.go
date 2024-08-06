package commit

import (
	"compress/zlib"
	"fmt"
	"io"
	"io/fs"
	"strings"

	"github.com/uragirii/got/internals/git/object"
	"github.com/uragirii/got/internals/git/sha"
	"github.com/uragirii/got/internals/git/tree"
)

type Commit struct {
	Tree      *tree.Tree
	parentSHA *sha.SHA
	sha       *sha.SHA
	message   string

	// author Person
	// commiter Person
}

var ErrInvalidCommit = fmt.Errorf("invalid commit")

func FromSHA(SHA *sha.SHA, gitFsys fs.FS) (*Commit, error) {

	objContents, err := object.FromSHA(SHA, gitFsys)

	if err != nil {
		return nil, err
	}

	if objContents.ObjType != object.CommitObj {
		return nil, ErrInvalidCommit
	}

	commitDetails := string(*objContents.Contents)

	msgStartIndex := strings.Index(commitDetails, "\n\n")

	if msgStartIndex == -1 {
		return nil, ErrInvalidCommit
	}

	commitMsg := commitDetails[msgStartIndex+2:] // 2 for \n\n

	commitDetails = commitDetails[:msgStartIndex]

	lines := strings.Split(commitDetails, "\n")

	treeLine := lines[0]
	parentLine := lines[1]

	// TODO: parse author and committer

	treeSha, err := sha.FromString(strings.Split(treeLine, " ")[1])

	if err != nil {
		return nil, err
	}

	parentSha, err := sha.FromString(strings.Split(parentLine, " ")[1])

	if err != nil {
		return nil, err
	}

	tree, err := tree.FromSHA(treeSha, gitFsys)

	if err != nil {
		return nil, err
	}

	return &Commit{
		Tree:      tree,
		sha:       SHA,
		message:   commitMsg,
		parentSHA: parentSha,
	}, nil

}

func (commit Commit) GetSHA() *sha.SHA {
	return commit.sha
}

func (commit Commit) GetObjType() object.ObjectType {
	return object.CommitObj
}

func (commit Commit) String() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("tree %s\n", commit.Tree.SHA))
	sb.WriteString(fmt.Sprintf("parent %s\n", commit.parentSHA))
	// 	author Apoorv Kansal <apoorvkansalak@gmail.com> 1720643686 +0530
	// committer Apoorv Kansal <apoorvkansalak@gmail.com> 1720643686 +053
	fmt.Println("WARN person and commiter not parsed for commit")
	sb.WriteString("author Apoorv Kansal <apoorvkansalak@gmail.com> 1720643686 +0530\n")
	sb.WriteString("committer Apoorv Kansal <apoorvkansalak@gmail.com> 1720643686 +0530\n")
	sb.WriteString("\n")
	sb.WriteString(strings.Trim(commit.message, "\n"))
	sb.WriteString("\n")

	return sb.String()

}

func (commit Commit) Write(writer io.Writer) error {
	contents := commit.String()

	w := zlib.NewWriter(writer)

	header := fmt.Sprintf(object.CommitHeader, len(contents))

	_, err := w.Write([]byte(header))

	if err != nil {
		return err
	}

	_, err = w.Write([]byte(contents))

	if err != nil {
		return err
	}

	return w.Close()
}

func (commit Commit) Raw() string {
	return commit.String()
}
