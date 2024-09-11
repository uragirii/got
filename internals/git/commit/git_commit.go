package commit

import (
	"bytes"
	"compress/zlib"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"strings"
	"time"

	"github.com/uragirii/got/internals"
	"github.com/uragirii/got/internals/git/head"
	"github.com/uragirii/got/internals/git/index"
	"github.com/uragirii/got/internals/git/object"
	"github.com/uragirii/got/internals/git/sha"
	"github.com/uragirii/got/internals/git/tree"
)

type Commit struct {
	Tree      *tree.Tree
	parentSHA *sha.SHA
	sha       *sha.SHA
	message   string

	author   person
	commiter person

	authorTime time.Time
	commitTime time.Time
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
	authorLine := lines[2]
	commiterLine := lines[3]

	treeSha, err := sha.FromString(strings.Split(treeLine, " ")[1])

	if err != nil {
		return nil, err
	}

	parentSha, err := sha.FromString(strings.Split(parentLine, " ")[1])

	if err != nil {
		return nil, err
	}

	tree, err := tree.FromSHA(treeSha, gitFsys)

	author, authorTime := parseAuthorLine(authorLine)
	commiter, commiterTime := parseAuthorLine(commiterLine)

	if err != nil {
		return nil, err
	}

	return &Commit{
		Tree:       tree,
		sha:        SHA,
		message:    commitMsg,
		parentSHA:  parentSha,
		author:     author,
		authorTime: authorTime,
		commiter:   commiter,
		commitTime: commiterTime,
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
	sb.WriteString("author Apoorv Kansal <apoorvkansalak@gmail.com> ")
	sb.WriteString(toGitTime(commit.authorTime))
	sb.WriteRune('\n')
	sb.WriteString("committer Apoorv Kansal <apoorvkansalak@gmail.com> ")
	sb.WriteString(toGitTime(commit.commitTime))
	sb.WriteRune('\n')
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

func (commit Commit) WriteToFile() error {
	objPath, err := commit.sha.GetObjPath()

	if err != nil {
		return err
	}

	gitDir, err := internals.GetGitDir()

	if err != nil {
		return err
	}

	objPath = path.Join(gitDir, objPath)

	if _, err := os.Stat(objPath); errors.Is(err, os.ErrNotExist) {
		var buffer bytes.Buffer

		err = commit.Write(&buffer)

		if err != nil {
			return err
		}

		err = os.MkdirAll(path.Join(objPath, ".."), 0755)

		if err != nil {
			return err
		}

		return os.WriteFile(objPath, buffer.Bytes(), 0444) // Read only file

	}

	return nil
}

func (commit Commit) Raw() string {
	return commit.String()
}

func (commit *Commit) CalculateSha() error {
	raw := []byte(commit.Raw())
	header := []byte(fmt.Sprintf(object.CommitHeader, len(raw)))
	raw = append(header, raw...)

	sha, err := sha.FromData(&raw)

	if err != nil {
		return err
	}

	commit.sha = sha

	return nil
}

func New(gitFs fs.FS, message string) (*Commit, error) {

	indexFile, err := gitFs.Open(index.IndexFileName)

	if err != nil {
		return nil, err
	}

	defer indexFile.Close()

	index, err := index.New(indexFile)

	if err != nil {
		return nil, err
	}

	if err = index.Hydrate(); err != nil {
		return nil, err
	}

	treeSha := index.GetTreeSHA()

	head, err := head.New(gitFs)

	if err != nil {
		return nil, err
	}

	tree, err := tree.FromSHA(treeSha, gitFs)

	if err != nil {
		return nil, err
	}

	// TODO: Read these from global config
	p := person{
		name:  "Apoorv Kansal",
		email: "apoorvkansalak@gmail.com",
	}

	commit := &Commit{
		parentSHA:  head.SHA,
		message:    message,
		Tree:       tree,
		author:     p,
		authorTime: time.Now(),
		commiter:   p,
		commitTime: time.Now(),
	}

	if err = commit.CalculateSha(); err != nil {
		return nil, err
	}

	return commit, nil
}
