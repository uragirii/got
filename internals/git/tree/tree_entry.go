package tree

import (
	"bytes"
	"fmt"
	"io/fs"
	"sort"

	"github.com/uragirii/got/internals/git/object"
	"github.com/uragirii/got/internals/git/sha"
)

type TreeEntry struct {
	Mode Mode
	// Its not the complete path
	Name string
	SHA  *sha.SHA
	tree *Tree
}

func (entry TreeEntry) String() string {
	return fmt.Sprintf("%s %s\t%s", entry.Mode.Pretty(), entry.SHA, entry.Name)
}

func (entry *TreeEntry) GetTree(fsys fs.FS) (*Tree, error) {
	if entry.Mode != ModeDir {
		return nil, fmt.Errorf("GetTree called on File entry")
	}

	if entry.tree != nil {
		return entry.tree, nil
	}

	tree, err := FromSHA(entry.SHA, fsys)

	if err != nil {
		return nil, err
	}

	entry.tree = tree

	return entry.tree, nil
}

func FromEnteries(enteries []TreeEntry) (*Tree, error) {
	sort.Slice(enteries, func(i, j int) bool {
		return enteries[i].Name < enteries[j].Name
	})

	tree := Tree{
		entries: enteries,
	}

	var buffer bytes.Buffer

	contents := []byte(tree.Raw())

	buffer.WriteString(fmt.Sprintf(object.TreeHeader, len(contents)))
	buffer.Write(contents)

	contents = buffer.Bytes()

	sha, err := sha.FromData(&contents)

	if err != nil {
		return nil, err
	}

	tree.SHA = sha

	return &tree, nil
}
