package object

import (
	"crypto/sha1"
	"fmt"
	"io/fs"
	"os"
	"path"
	"sort"
	"strings"
	"sync"

	"github.com/uragirii/got/internals"
	"github.com/uragirii/got/internals/git"
)

type Mode string

const (
	ModeNormal     Mode = "100644"
	ModeExecutable Mode = "100755"
	// Sym link is not supported at the moment
	ModeSymLink Mode = "120000"
	ModeDir     Mode = "40000"
)

const _TreeHeader string = "blob %d\u0000"

func (mode Mode) Pretty() string {
	if mode == ModeDir {
		return fmt.Sprintf("0%s tree", mode)
	}
	return fmt.Sprintf("%s blob", mode)
}

type treeEntry struct {
	mode Mode
	// Its not the complete path
	name string
	sha  *git.SHA
	tree *Tree
}

func (entry treeEntry) String() string {
	return fmt.Sprintf("%s %s\t%s", entry.mode.Pretty(), entry.sha.MarshallToStr(), entry.name)
}

type Tree struct {
	entries []treeEntry
	sha     *git.SHA
}

var ErrInvalidTree = fmt.Errorf("invalid tree")

func ToTree(obj *Object) (*Tree, error) {
	if obj.GetObjType() != TreeObj {
		return nil, ErrInvalidTree
	}

	treeContents := string(*obj.getContentWithoutHeader())

	var entries []treeEntry

	for currIdx := 0; currIdx < len(treeContents); {

		startIdx := currIdx

		for ; treeContents[currIdx] != ' '; currIdx++ {

		}

		mode := Mode(treeContents[startIdx:currIdx])

		currIdx++

		startIdx = currIdx

		for ; treeContents[currIdx] != 0x00; currIdx++ {

		}

		name := treeContents[startIdx:currIdx]

		currIdx++
		startIdx = currIdx

		currIdx += git.SHA_BYTES_LEN

		shaBytes := []byte(treeContents[startIdx:currIdx])

		sha, err := git.SHAFromByteSlice(&shaBytes)

		if err != nil {
			return nil, err
		}

		entries = append(entries, treeEntry{
			mode: mode,
			name: name,
			sha:  sha,
		})
	}

	return &Tree{
		entries: entries,
		sha:     obj.sha,
	}, nil

}

func getSHAFromEntries(sortedEntries []treeEntry) (*git.SHA, error) {
	var content strings.Builder

	for _, entry := range sortedEntries {
		content.WriteString(fmt.Sprintf("%s %s\u0000%s", entry.mode, entry.name, entry.sha.GetBinStr()))
	}

	contentLen := content.Len()

	header := []byte(fmt.Sprintf(_TreeHeader, contentLen))

	contentBytes := append(header, []byte(content.String())...)

	hash := sha1.Sum(contentBytes)
	hashSlice := hash[:]

	return git.SHAFromByteSlice(&hashSlice)
}

func getModeFromAbsPath(absPath string) Mode {
	// FIXME: ignoring error
	fileInfo, _ := os.Stat(absPath)

	if !fileInfo.Mode().IsRegular() {
		return ModeExecutable
	}

	return ModeNormal
}

func getTreeForDir(dirPath string, ignore *git.Ignore) (*Tree, error) {
	items, err := os.ReadDir(dirPath)

	if err != nil {
		return nil, err
	}

	entries := make([]treeEntry, 0, len(items))

	ignoreFile, err := ignore.WithFile(path.Join(dirPath, ".gitignore"))

	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup

	for idx, entry := range items {

		absPath := path.Join(dirPath, entry.Name())

		if ignoreFile.Match(absPath) {
			continue
		}

		if entry.IsDir() {

			if entry.Name() != ".git" {
				wg.Add(1)
				go func(entry fs.DirEntry, idx int) {
					defer wg.Done()

					subTree, err := getTreeForDir(absPath, ignoreFile)

					if err != nil {
						panic(err)
					}

					entries = append(entries, treeEntry{
						mode: ModeDir,
						name: entry.Name(),
						sha:  subTree.sha,
						tree: subTree,
					})

				}(entry, idx)

			}
		} else {
			wg.Add(1)
			// is a file
			go func(entry fs.DirEntry, idx int) {
				defer wg.Done()

				obj, err := NewObject(absPath)

				if err != nil {
					panic(err)
				}

				entries = append(entries, treeEntry{
					name: entry.Name(),
					mode: getModeFromAbsPath(absPath),
					sha:  obj.sha,
				})

			}(entry, idx)
		}
	}

	wg.Wait()

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].name < entries[j].name
	})

	sha, err := getSHAFromEntries(entries)

	if err != nil {
		return nil, err
	}

	return &Tree{
		entries: entries,
		sha:     sha,
	}, nil

}

func NewTree() (*Tree, error) {

	gitDir, err := internals.GetGitDir()

	if err != nil {
		return nil, err
	}

	rootDir := path.Join(gitDir, "..")

	ignore, err := git.NewIgnore(path.Join(rootDir, ".gitignore"))

	if err != nil {
		return nil, err
	}

	return getTreeForDir(rootDir, ignore)
}

func (tree Tree) String() string {
	var sb strings.Builder

	for i, entry := range tree.entries {
		sb.WriteString(entry.String())
		if i != len(tree.entries)-1 {
			sb.WriteRune('\n')
		}
	}

	return sb.String()
}
