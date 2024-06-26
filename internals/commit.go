package internals

// import (
// 	"fmt"
// 	"path"
// 	"strings"
// 	"time"
// )

// type GitTreeFile struct {
// 	Name string
// 	SHA  string
// }

// type GitTree struct {
// 	SHA      string
// 	Path     string
// 	Files    map[string]*GitTreeFile
// 	SubTrees map[string]*GitTree
// }

// func (gt *GitTree) LoadChildren() {
// 	objType, contentBytes, err := ReadGitObject(gt.SHA)

// 	if err != nil {
// 		panic(err)
// 	}

// 	if objType != "tree" {
// 		panic("invalid sha hash for tree")
// 	}

// 	// TODO: tweak this value for perf
// 	gt.Files = make(map[string]*GitTreeFile)
// 	gt.SubTrees = make(map[string]*GitTree)

// 	for idx := 0; idx < len(*contentBytes); {
// 		modeStartIdx := idx
// 		for ; (*contentBytes)[idx] != ' '; idx++ {
// 		}

// 		mode := string((*contentBytes)[modeStartIdx:idx])

// 		isDir := mode == "40000"

// 		idx++
// 		nameStartIdx := idx
// 		for ; (*contentBytes)[idx] != 0x00; idx++ {
// 		}

// 		name := (*contentBytes)[nameStartIdx:idx]
// 		idx++

// 		sha := (*contentBytes)[idx : idx+20]

// 		idx += 20

// 		if isDir {
// 			gt.SubTrees[path.Join(gt.Path, string(name))] = &GitTree{
// 				Path: path.Join(gt.Path, string(name)),
// 				SHA:  fmt.Sprintf("%x", sha),
// 			}
// 		} else {

// 			gt.Files[path.Join(gt.Path, string(name))] = &GitTreeFile{
// 				Name: string(name),
// 				SHA:  fmt.Sprintf("%x", sha),
// 			}
// 		}
// 	}

// }

// type GitPerson struct {
// }

// type GitCommit struct {
// 	SHA       string
// 	ParentSHA string
// 	Tree      *GitTree
// 	parent    *GitCommit
// 	commiter  *GitPerson
// 	timestamp time.Time
// 	message   string
// }

// const TREE_PREFIX string = "tree "
// const PARENT_PREFIX string = "parent "

// func parseTreeLine(treeLine string) (*GitTree, error) {

// 	if !(strings.HasPrefix(treeLine, TREE_PREFIX)) {
// 		return nil, fmt.Errorf("invalid commit file tree")
// 	}

// 	treeSha := treeLine[len(TREE_PREFIX):]

// 	if len(treeSha) != 40 {
// 		return nil, fmt.Errorf("invalid tree sha")
// 	}

// 	return &GitTree{
// 		SHA: treeSha,
// 	}, nil

// }

// func parseParentSha(parentLine string) (string, error) {
// 	if !(strings.HasPrefix(parentLine, PARENT_PREFIX)) {
// 		return "", fmt.Errorf("invalid commit file")
// 	}

// 	parentSha := parentLine[len(PARENT_PREFIX):]

// 	if len(parentSha) != 40 {
// 		return "", fmt.Errorf("invalid parent sha")
// 	}

// 	return parentSha, nil
// }

// func ParseCommit(sha string) (*GitCommit, error) {
// 	objType, contentBytes, err := ReadGitObject(sha)

// 	if err != nil {
// 		return nil, err
// 	}

// 	if objType != "commit" {
// 		return nil, fmt.Errorf("%s sha is not commit", sha)
// 	}

// 	content := string(*contentBytes)

// 	lines := strings.Split(content, "\n")

// 	tree, err := parseTreeLine(lines[0])

// 	if err != nil {
// 		return nil, err
// 	}
// 	gitDir, err := GetGitDir()
// 	if err != nil {
// 		return nil, err
// 	}

// 	tree.Path = path.Join(gitDir, "..")

// 	parentSha, err := parseParentSha(lines[1])

// 	if err != nil {
// 		return nil, err
// 	}

// 	// TODO author and commiter parse

// 	message := strings.Join(lines[5:], "\n")

// 	return &GitCommit{
// 		SHA:       sha,
// 		ParentSHA: parentSha,
// 		Tree:      tree,
// 		message:   message,
// 	}, nil

// }
