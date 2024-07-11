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
	"github.com/uragirii/got/internals/git/sha"
)

type Mode string

const (
	ModeNormal     Mode = "100644"
	ModeExecutable Mode = "100755"
	// Deprecated: Sym link is not supported at the moment
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
	sha  *sha.SHA
	tree *Tree
}

func (entry treeEntry) String() string {
	return fmt.Sprintf("%s %s\t%s", entry.mode.Pretty(), entry.sha.MarshallToStr(), entry.name)
}

func (entry *treeEntry) GetTree() (*Tree, error) {
	if entry.mode != ModeDir {
		return nil, fmt.Errorf("GetTree called on File entry")
	}

	if entry.tree != nil {
		return entry.tree, nil
	}

	obj, err := NewObjectFromSHA(entry.sha)

	if err != nil {
		return nil, err
	}

	tree, err := ToTree(obj)

	if err != nil {
		return nil, err
	}

	entry.tree = tree

	return entry.tree, nil
}

type Tree struct {
	entries []treeEntry
	sha     *sha.SHA
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

		currIdx += sha.BYTES_LEN

		shaBytes := []byte(treeContents[startIdx:currIdx])

		sha, err := sha.FromByteSlice(&shaBytes)

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

func getSHAFromEntries(sortedEntries []treeEntry) (*sha.SHA, error) {
	var content strings.Builder

	for _, entry := range sortedEntries {
		content.WriteString(fmt.Sprintf("%s %s\u0000%s", entry.mode, entry.name, entry.sha.GetBinStr()))
	}

	contentLen := content.Len()

	header := []byte(fmt.Sprintf(_TreeHeader, contentLen))

	contentBytes := append(header, []byte(content.String())...)

	hash := sha1.Sum(contentBytes)
	hashSlice := hash[:]

	return sha.FromByteSlice(&hashSlice)
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

type ChangeStatus int

func (status ChangeStatus) String() string {
	switch status {
	case StatusModified:
		return "modified"
	case StatusDeleted:
		return "deleted "
	case StatusAdded:
		return "new file"
	}
	return "invalid"
}

const (
	StatusModified ChangeStatus = iota
	StatusAdded
	StatusDeleted
)

type ChangeItem struct {
	Status  ChangeStatus
	RelPath string
	SHA     *sha.SHA
}

// Call compare function on the Tree generated from git
// then compare it with the Live Tree
func (tree Tree) Compare(other *Tree) ([]ChangeItem, error) {

	if tree.sha.Eq(other.sha) {
		return []ChangeItem{}, nil
	}

	treeNames := make(map[string]*treeEntry, len(tree.entries))
	otherTreeNames := make(map[string]*treeEntry, len(other.entries))

	changeChan := make(chan ChangeItem)

	for _, entry := range tree.entries {
		treeNames[entry.name] = &entry
	}

	for _, entry := range other.entries {
		otherTreeNames[entry.name] = &entry
	}

	var wg sync.WaitGroup

	visited := make(map[string]bool)

	for _, entry := range tree.entries {
		otherEntry, ok := otherTreeNames[entry.name]

		visited[entry.name] = true

		if !ok {
			// Item is deleted
			wg.Add(1)
			// Need to wrap this in goroutine as channel is unbuffered
			go func(entry treeEntry) {
				defer wg.Done()

				changeChan <- ChangeItem{
					Status:  StatusDeleted,
					RelPath: entry.name,
				}
			}(entry)

		} else if !entry.sha.Eq(otherEntry.sha) {
			// Item is modified

			if entry.mode != ModeDir {
				wg.Add(1)
				go func(entry treeEntry) {
					defer wg.Done()

					changeChan <- ChangeItem{
						Status:  StatusModified,
						RelPath: entry.name,
						SHA:     otherEntry.sha,
					}
				}(entry)

			} else {
				entryTree, err := entry.GetTree()

				if err != nil {
					return nil, err
				}

				wg.Add(1)

				go func(entry treeEntry) {
					defer wg.Done()

					subChanges, err := entryTree.Compare(otherEntry.tree)

					if err != nil {
						panic(err)
					}

					for _, change := range subChanges {
						changeChan <- ChangeItem{
							Status:  change.Status,
							RelPath: path.Join(entry.name, change.RelPath),
							SHA:     change.SHA,
						}
					}

				}(entry)

			}

		}
	}

	for _, otherEntry := range other.entries {
		entry, ok := treeNames[otherEntry.name]

		if visited[otherEntry.name] {
			continue
		}

		if !ok {
			// Item is Added
			wg.Add(1)
			go func(otherEntry treeEntry) {
				defer wg.Done()
				changeChan <- ChangeItem{
					Status:  StatusAdded,
					RelPath: otherEntry.name,
					SHA:     otherEntry.sha,
				}
			}(otherEntry)

		} else if !entry.sha.Eq(otherEntry.sha) {
			// Item is modified

			if entry.mode != ModeDir {
				wg.Add(1)
				go func(otherEntry treeEntry) {
					defer wg.Done()

					changeChan <- ChangeItem{
						Status:  StatusModified,
						RelPath: otherEntry.name,
						SHA:     otherEntry.sha,
					}
				}(otherEntry)

			} else {
				entryTree, err := entry.GetTree()

				if err != nil {
					return nil, err
				}

				wg.Add(1)

				go func(otherEntry treeEntry) {
					defer wg.Done()

					subChanges, err := entryTree.Compare(otherEntry.tree)

					if err != nil {
						panic(err)
					}

					for _, change := range subChanges {
						changeChan <- ChangeItem{
							Status:  change.Status,
							RelPath: path.Join(otherEntry.name, change.RelPath),
							SHA:     change.SHA,
						}
					}

				}(otherEntry)

			}

		}
	}

	go func() {
		wg.Wait()
		close(changeChan)
	}()

	var changeItems []ChangeItem

	for change := range changeChan {
		changeItems = append(changeItems, change)
	}

	return changeItems, nil
}
