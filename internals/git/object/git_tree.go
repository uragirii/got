package object

import (
	"fmt"
	"strings"

	"github.com/uragirii/got/internals/git"
)

type Mode string

const (
	ModeNormal     Mode = "100644"
	ModeExecutable Mode = "100755"
	ModeSymLink    Mode = "120000"
	ModeDir        Mode = "40000"
)

func (mode Mode) Pretty() string {
	if mode == ModeDir {
		return fmt.Sprintf("0%s tree", mode)
	}
	return fmt.Sprintf("%s blob", mode)
}

type treeEntry struct {
	mode Mode
	name string
	sha  *git.SHA
}

func (entry treeEntry) String() string {
	return fmt.Sprintf("%s %s\t%s", entry.mode.Pretty(), entry.sha.MarshallToStr(), entry.name)
}

type Tree struct {
	entries []treeEntry
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
	}, nil

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
