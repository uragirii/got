package tree

import (
	"io/fs"
	"path"
	"sync"

	"github.com/uragirii/got/internals/git/sha"
)

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
func (tree Tree) Compare(other *Tree, fsys fs.FS) ([]ChangeItem, error) {

	if tree.SHA.Eq(other.SHA) {
		return []ChangeItem{}, nil
	}

	treeNames := make(map[string]*TreeEntry, len(tree.entries))
	otherTreeNames := make(map[string]*TreeEntry, len(other.entries))

	changeChan := make(chan ChangeItem)

	for _, entry := range tree.entries {
		treeNames[entry.Name] = &entry
	}

	for _, entry := range other.entries {
		otherTreeNames[entry.Name] = &entry
	}

	var wg sync.WaitGroup

	visited := make(map[string]bool)

	for _, entry := range tree.entries {
		otherEntry, ok := otherTreeNames[entry.Name]

		visited[entry.Name] = true

		if !ok {
			// Item is deleted
			wg.Add(1)
			// Need to wrap this in goroutine as channel is unbuffered
			go func(entry TreeEntry) {
				defer wg.Done()

				changeChan <- ChangeItem{
					Status:  StatusDeleted,
					RelPath: entry.Name,
				}
			}(entry)

		} else if !entry.SHA.Eq(otherEntry.SHA) {
			// Item is modified

			if entry.Mode != ModeDir {
				wg.Add(1)
				go func(entry TreeEntry) {
					defer wg.Done()

					changeChan <- ChangeItem{
						Status:  StatusModified,
						RelPath: entry.Name,
						SHA:     otherEntry.SHA,
					}
				}(entry)

			} else {
				entryTree, err := entry.GetTree(fsys)

				if err != nil {
					return nil, err
				}

				wg.Add(1)

				go func(entry TreeEntry) {
					defer wg.Done()

					subChanges, err := entryTree.Compare(otherEntry.tree, fsys)

					if err != nil {
						panic(err)
					}

					for _, change := range subChanges {
						changeChan <- ChangeItem{
							Status:  change.Status,
							RelPath: path.Join(entry.Name, change.RelPath),
							SHA:     change.SHA,
						}
					}

				}(entry)

			}

		}
	}

	for _, otherEntry := range other.entries {
		entry, ok := treeNames[otherEntry.Name]

		if visited[otherEntry.Name] {
			continue
		}

		if !ok {
			// Item is Added
			wg.Add(1)
			go func(otherEntry TreeEntry) {
				defer wg.Done()
				changeChan <- ChangeItem{
					Status:  StatusAdded,
					RelPath: otherEntry.Name,
					SHA:     otherEntry.SHA,
				}
			}(otherEntry)

		} else if !entry.SHA.Eq(otherEntry.SHA) {
			// Item is modified

			if entry.Mode != ModeDir {
				wg.Add(1)
				go func(otherEntry TreeEntry) {
					defer wg.Done()

					changeChan <- ChangeItem{
						Status:  StatusModified,
						RelPath: otherEntry.Name,
						SHA:     otherEntry.SHA,
					}
				}(otherEntry)

			} else {
				entryTree, err := entry.GetTree(fsys)

				if err != nil {
					return nil, err
				}

				wg.Add(1)

				go func(otherEntry TreeEntry) {
					defer wg.Done()

					subChanges, err := entryTree.Compare(otherEntry.tree, fsys)

					if err != nil {
						panic(err)
					}

					for _, change := range subChanges {
						changeChan <- ChangeItem{
							Status:  change.Status,
							RelPath: path.Join(otherEntry.Name, change.RelPath),
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
