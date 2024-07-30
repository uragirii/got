package tree_test

import (
	"testing"
	"testing/fstest"

	"github.com/uragirii/got/internals/git/tree"
)

const TREE_DIR_STR = "6d06cf3cdd7763dc6d938823dd3ca9cb9ccb8cb3"
const TREE_DIR_PRETTY_STR = `100644 blob b69bcf74597bd6e5a0d72137844ad3a4ee664615	.gitignore
040000 tree 8d81c5bd08c7105f3dd617849f68a1821ffbf61d	dir1
100644 blob 366f17ff507eeda97ee143e1ae7ef7933e52f89b	file1.txt
`

var TREE_DIR_RAW = []byte{0x31, 0x30, 0x30, 0x36, 0x34, 0x34, 0x20, 0x2e, 0x67, 0x69, 0x74, 0x69, 0x67, 0x6e, 0x6f, 0x72,
	0x65, 0x00, 0xb6, 0x9b, 0xcf, 0x74, 0x59, 0x7b, 0xd6, 0xe5, 0xa0, 0xd7, 0x21, 0x37, 0x84, 0x4a,
	0xd3, 0xa4, 0xee, 0x66, 0x46, 0x15, 0x34, 0x30, 0x30, 0x30, 0x30, 0x20, 0x64, 0x69, 0x72, 0x31,
	0x00, 0x8d, 0x81, 0xc5, 0xbd, 0x08, 0xc7, 0x10, 0x5f, 0x3d, 0xd6, 0x17, 0x84, 0x9f, 0x68, 0xa1,
	0x82, 0x1f, 0xfb, 0xf6, 0x1d, 0x31, 0x30, 0x30, 0x36, 0x34, 0x34, 0x20, 0x66, 0x69, 0x6c, 0x65,
	0x31, 0x2e, 0x74, 0x78, 0x74, 0x00, 0x36, 0x6f, 0x17, 0xff, 0x50, 0x7e, 0xed, 0xa9, 0x7e, 0xe1,
	0x43, 0xe1, 0xae, 0x7e, 0xf7, 0x93, 0x3e, 0x52, 0xf8, 0x9b}

func TestFromDir(t *testing.T) {
	t.Setenv("GIT_DIR", ".git")

	mapFS := fstest.MapFS{
		"ignore/ignore.txt": {Data: []byte(`This file should be ignored
`)},
		"file1.txt": {Data: []byte(`file 1
`)},
		"dir1/file2.txt": {Data: []byte(`file 2
`)},
		".gitignore": {Data: []byte(`/ignore
`)},
	}

	tree, err := tree.FromDir(mapFS)

	if err != nil {
		t.Errorf("expected error to be nil but got %s", err)
	}

	if tree.SHA.String() != TREE_DIR_STR {
		t.Errorf("expected error to be %s but got %s", TREE_DIR_STR, tree.SHA)
	}

	if tree.String() != TREE_DIR_PRETTY_STR {
		t.Errorf("expected error to be %s but got %s", TREE_DIR_PRETTY_STR, tree)
	}

	if tree.Raw() != string(TREE_DIR_RAW) {
		t.Errorf("expected error to be % x but got % x", tree.Raw(), TREE_DIR_RAW)
	}

}