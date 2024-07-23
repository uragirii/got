package object

import (
	"fmt"
	"strings"

	"github.com/uragirii/got/internals/git/sha"
)

type Commit struct {
	Tree      *Tree
	parentSHA *sha.SHA
	sha       *sha.SHA
	message   string

	// author Person
	// commiter Person
}

var ErrInvalidCommit = fmt.Errorf("invalid commit")

func ToCommit(obj *Object) (*Commit, error) {
	if obj.objectType != CommitObj {
		return nil, ErrInvalidCommit
	}

	commitDetails := string(*obj.getContentWithoutHeader())

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

	treeObj, err := NewObjectFromSHA(treeSha)

	if err != nil {
		return nil, err
	}

	tree, err := ToTree(treeObj)

	if err != nil {
		return nil, err
	}

	return &Commit{
		Tree:      tree,
		sha:       obj.sha,
		message:   commitMsg,
		parentSHA: parentSha,
	}, nil
}

func (commit Commit) String() string {
	return fmt.Sprintf("commit %s\n\n\t%s", commit.sha.MarshallToStr(), commit.message)
}
