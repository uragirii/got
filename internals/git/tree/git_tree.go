package tree

import (
	"bytes"
	"compress/zlib"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/uragirii/got/internals"
	"github.com/uragirii/got/internals/git/object"
	"github.com/uragirii/got/internals/git/sha"
)

type Tree struct {
	entries []TreeEntry
	SHA     *sha.SHA
}

func FromSHA(SHA *sha.SHA, fsys fs.FS) (*Tree, error) {

	treeObj, err := object.FromSHA(SHA, fsys)

	if err != nil {
		return nil, err
	}

	if treeObj.ObjType != object.TreeObj {
		return nil, ErrInvalidTree
	}

	treeContents := string(*treeObj.Contents)

	var entries []TreeEntry

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

		entries = append(entries, TreeEntry{
			Mode: mode,
			Name: name,
			SHA:  sha,
		})
	}

	return &Tree{
		entries: entries,
		SHA:     SHA,
	}, nil

}

func (tree Tree) GetSHA() *sha.SHA {
	return tree.SHA
}

func (tree Tree) Write(writer io.Writer) error {
	w := zlib.NewWriter(writer)

	contents := tree.Raw()

	contentLen := len(contents)

	header := fmt.Sprintf(object.TreeHeader, contentLen)

	w.Write([]byte(header))
	w.Write([]byte(contents))

	return w.Close()
}

func (tree Tree) WriteToFile() error {
	objPath, err := tree.SHA.GetObjPath()

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

		err = tree.Write(&buffer)

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

func (tree *Tree) sortEnteries() {
	sort.Slice(tree.entries, func(i, j int) bool {
		return tree.entries[i].Name < tree.entries[j].Name
	})
}

func (tree Tree) GetObjType() object.ObjectType {
	return object.TreeObj
}

func (tree Tree) String() string {
	var sb strings.Builder

	tree.sortEnteries()

	for _, entry := range tree.entries {
		sb.WriteString(entry.String())
		sb.WriteRune('\n')
	}

	return sb.String()
}

func (tree Tree) Raw() string {
	var sb strings.Builder

	tree.sortEnteries()

	for _, entry := range tree.entries {
		sb.WriteString(string(entry.Mode))
		sb.WriteRune(' ')
		sb.WriteString(entry.Name)
		sb.WriteByte(0x00)
		sb.WriteString(entry.SHA.GetBinStr())
	}

	return sb.String()

}
