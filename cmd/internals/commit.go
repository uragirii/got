package internals

import (
	"fmt"
	"strings"
	"time"
)

type GitTree struct {
	SHA string
}

type GitPerson struct {
}

type GitCommit struct {
	SHA       string
	ParentSHA string
	Tree      *GitTree
	parent    *GitCommit
	commiter  *GitPerson
	timestamp time.Time
	message   string
}

const TREE_PREFIX string = "tree "
const PARENT_PREFIX string = "parent "

func parseTreeLine(treeLine string) (*GitTree, error) {

	fmt.Println(strings.HasPrefix(treeLine, TREE_PREFIX))

	if !(strings.HasPrefix(treeLine, TREE_PREFIX)) {
		return nil, fmt.Errorf("invalid commit file tree")
	}

	treeSha := treeLine[len(TREE_PREFIX):]

	if len(treeSha) != 40 {
		return nil, fmt.Errorf("invalid tree sha")
	}

	return &GitTree{
		SHA: treeSha,
	}, nil

}

func parseParentSha(parentLine string) (string, error) {
	if !(strings.HasPrefix(parentLine, PARENT_PREFIX)) {
		return "", fmt.Errorf("invalid commit file")
	}

	parentSha := parentLine[len(PARENT_PREFIX):]

	if len(parentSha) != 40 {
		return "", fmt.Errorf("invalid parent sha")
	}

	return parentSha, nil
}

func ParseCommit(sha string) (*GitCommit, error) {
	objType, contentBytes, err := ReadGitObject(sha)

	if err != nil {
		return nil, err
	}

	if objType != "commit" {
		return nil, fmt.Errorf("%s sha is not commit", sha)
	}

	content := string(*contentBytes)

	lines := strings.Split(content, "\n")

	tree, err := parseTreeLine(lines[0])

	if err != nil {
		return nil, err
	}

	parentSha, err := parseParentSha(lines[1])

	if err != nil {
		return nil, err
	}

	// TODO author and commiter parse

	message := strings.Join(lines[5:], "\n")

	return &GitCommit{
		SHA:       sha,
		ParentSHA: parentSha,
		Tree:      tree,
		message:   message,
	}, nil

}
