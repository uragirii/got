package tree_test

import (
	"bytes"
	"testing"
	"testing/fstest"

	"github.com/uragirii/got/internals/git/object"
	"github.com/uragirii/got/internals/git/sha"
	"github.com/uragirii/got/internals/git/tree"
	testutils "github.com/uragirii/got/internals/test_utils"
)

const TEST_TREE_STR = `100644 blob 66305f506530406dcb4bd5bf5534c00bbdb3cb35	.gitignore
100644 blob bf8d4172cc250966d75337d707533e1f8a3b9dd2	Makefile
100644 blob f6ab07254cbfd820ed21f45bf646c22807d0d60f	Readme.md
040000 tree c6d869535b23a83941da6d2dedfc692996f9f24e	cmd
100644 blob 35f59388fd65cc4a937f0bda95092edc3c594263	go.mod
040000 tree 7cd6e558afb9b64f0224ffe734ee18acddd53a40	internals
100644 blob 8c8aacee78c68befeed133efb792e02e1b5cfd83	main.go
040000 tree 2cb967b04075b3de9c8b2d71c04e95c6efc4c609	tests
`

const DATA_SHA_STR = "011bf1e3d368c58d3fd7ea584d69a9acb6b21c13"

var TREE_COMPRESSED_BYTES = []byte{0x78, 0x9c, 0x2a, 0x29, 0x4a, 0x4d, 0x55, 0x30, 0x32, 0xb7, 0x60, 0x30, 0x34, 0x30, 0x30, 0x33,
	0x31, 0x51, 0xd0, 0x4b, 0xcf, 0x2c, 0xc9, 0x4c, 0xcf, 0xcb, 0x2f, 0x4a, 0x65, 0x48, 0x33, 0x88,
	0x0f, 0x48, 0x35, 0x70, 0xc8, 0x3d, 0xed, 0x7d, 0x75, 0x7f, 0xa8, 0xc9, 0x01, 0xee, 0xbd, 0x9b,
	0x4f, 0x9b, 0x42, 0x55, 0xf9, 0x26, 0x66, 0xa7, 0xa6, 0x65, 0xe6, 0xa4, 0x32, 0xec, 0xef, 0x75,
	0x2c, 0x3a, 0xa3, 0xca, 0x99, 0x76, 0x3d, 0xd8, 0xfc, 0x3a, 0x7b, 0xb0, 0x9d, 0x7c, 0x97, 0xf5,
	0xdc, 0x4b, 0x50, 0x35, 0x41, 0xa9, 0x89, 0x29, 0xb9, 0xa9, 0x7a, 0xb9, 0x29, 0x0c, 0xdf, 0x56,
	0xb3, 0xab, 0xfa, 0xec, 0xbf, 0xa1, 0xf0, 0x56, 0xf1, 0x4b, 0xf4, 0x37, 0xb7, 0x43, 0x1a, 0xec,
	0x17, 0xae, 0xf1, 0x9b, 0x18, 0x18, 0x18, 0x18, 0x28, 0x24, 0xe7, 0xa6, 0x30, 0x1c, 0xbb, 0x91,
	0x19, 0x1c, 0xad, 0xbc, 0xc2, 0xd2, 0xf1, 0x56, 0xae, 0xee, 0xdb, 0x3f, 0x99, 0x9a, 0xd3, 0x7e,
	0x7e, 0xf2, 0x83, 0x1a, 0x91, 0x9e, 0xaf, 0x97, 0x9b, 0x9f, 0xc2, 0x60, 0xfa, 0x75, 0x72, 0xc7,
	0xdf, 0xd4, 0x33, 0x5e, 0x93, 0xeb, 0xb9, 0x6f, 0x4d, 0xe5, 0xd4, 0xbb, 0x63, 0x13, 0xe9, 0x94,
	0x0c, 0xd1, 0x9f, 0x99, 0x57, 0x92, 0x5a, 0x94, 0x97, 0x98, 0x53, 0xcc, 0x50, 0x73, 0xed, 0x69,
	0xc4, 0xfa, 0x9d, 0xdb, 0xfc, 0x99, 0x54, 0xfe, 0x3f, 0x37, 0x79, 0x27, 0xb1, 0xe6, 0xee, 0x55,
	0x2b, 0x07, 0xa8, 0x29, 0xb9, 0x89, 0x99, 0x79, 0x7a, 0xe9, 0xf9, 0x0c, 0x3d, 0x5d, 0x6b, 0xde,
	0x55, 0x1c, 0xeb, 0x7e, 0xff, 0xee, 0xa2, 0xf1, 0xfb, 0xed, 0x93, 0x1e, 0xe8, 0x49, 0xc7, 0xfc,
	0x6d, 0x86, 0x18, 0x53, 0x92, 0x5a, 0x5c, 0x52, 0xcc, 0xa0, 0xb3, 0x33, 0x7d, 0x83, 0x43, 0xe9,
	0xe6, 0x7b, 0x73, 0xba, 0x75, 0x0b, 0x0f, 0xf8, 0x4d, 0x3d, 0xf6, 0xfe, 0xc8, 0x31, 0x4e, 0x40,
	0x00, 0x00, 0x00, 0xff, 0xff, 0x52, 0x4a, 0x75, 0x35}

var TREE_RAW_OUTPUT = []byte{0x31, 0x30, 0x30, 0x36, 0x34, 0x34, 0x20, 0x2e, 0x67, 0x69, 0x74, 0x69, 0x67, 0x6e, 0x6f, 0x72,
	0x65, 0x00, 0x66, 0x30, 0x5f, 0x50, 0x65, 0x30, 0x40, 0x6d, 0xcb, 0x4b, 0xd5, 0xbf, 0x55, 0x34,
	0xc0, 0x0b, 0xbd, 0xb3, 0xcb, 0x35, 0x31, 0x30, 0x30, 0x36, 0x34, 0x34, 0x20, 0x4d, 0x61, 0x6b,
	0x65, 0x66, 0x69, 0x6c, 0x65, 0x00, 0xbf, 0x8d, 0x41, 0x72, 0xcc, 0x25, 0x09, 0x66, 0xd7, 0x53,
	0x37, 0xd7, 0x07, 0x53, 0x3e, 0x1f, 0x8a, 0x3b, 0x9d, 0xd2, 0x31, 0x30, 0x30, 0x36, 0x34, 0x34,
	0x20, 0x52, 0x65, 0x61, 0x64, 0x6d, 0x65, 0x2e, 0x6d, 0x64, 0x00, 0xf6, 0xab, 0x07, 0x25, 0x4c,
	0xbf, 0xd8, 0x20, 0xed, 0x21, 0xf4, 0x5b, 0xf6, 0x46, 0xc2, 0x28, 0x07, 0xd0, 0xd6, 0x0f, 0x34,
	0x30, 0x30, 0x30, 0x30, 0x20, 0x63, 0x6d, 0x64, 0x00, 0xc6, 0xd8, 0x69, 0x53, 0x5b, 0x23, 0xa8,
	0x39, 0x41, 0xda, 0x6d, 0x2d, 0xed, 0xfc, 0x69, 0x29, 0x96, 0xf9, 0xf2, 0x4e, 0x31, 0x30, 0x30,
	0x36, 0x34, 0x34, 0x20, 0x67, 0x6f, 0x2e, 0x6d, 0x6f, 0x64, 0x00, 0x35, 0xf5, 0x93, 0x88, 0xfd,
	0x65, 0xcc, 0x4a, 0x93, 0x7f, 0x0b, 0xda, 0x95, 0x09, 0x2e, 0xdc, 0x3c, 0x59, 0x42, 0x63, 0x34,
	0x30, 0x30, 0x30, 0x30, 0x20, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x73, 0x00, 0x7c,
	0xd6, 0xe5, 0x58, 0xaf, 0xb9, 0xb6, 0x4f, 0x02, 0x24, 0xff, 0xe7, 0x34, 0xee, 0x18, 0xac, 0xdd,
	0xd5, 0x3a, 0x40, 0x31, 0x30, 0x30, 0x36, 0x34, 0x34, 0x20, 0x6d, 0x61, 0x69, 0x6e, 0x2e, 0x67,
	0x6f, 0x00, 0x8c, 0x8a, 0xac, 0xee, 0x78, 0xc6, 0x8b, 0xef, 0xee, 0xd1, 0x33, 0xef, 0xb7, 0x92,
	0xe0, 0x2e, 0x1b, 0x5c, 0xfd, 0x83, 0x34, 0x30, 0x30, 0x30, 0x30, 0x20, 0x74, 0x65, 0x73, 0x74,
	0x73, 0x00, 0x2c, 0xb9, 0x67, 0xb0, 0x40, 0x75, 0xb3, 0xde, 0x9c, 0x8b, 0x2d, 0x71, 0xc0, 0x4e,
	0x95, 0xc6, 0xef, 0xc4, 0xc6, 0x09}

func TestTree(t *testing.T) {
	t.Run("parses the tree correctly from SHA", func(t *testing.T) {
		t.Setenv("GIT_DIR", ".git")

		objSha, _ := sha.FromString(DATA_SHA_STR)

		mapFs := fstest.MapFS{}

		objPath, _ := objSha.GetObjPath()

		mapFs[objPath] = &fstest.MapFile{Data: TREE_COMPRESSED_BYTES}

		fs := fstest.MapFS(mapFs)

		tree, err := tree.FromSHA(objSha, fs)

		if err != nil {
			t.Error(err)
		}

		if !tree.GetSHA().Eq(objSha) {
			t.Errorf("expected the sha to be %s but got %s", DATA_SHA_STR, tree.GetSHA())
		}

		testutils.AssertString(t, "string", TEST_TREE_STR, tree.String())
		testutils.AssertString(t, "raw", string(TREE_RAW_OUTPUT), tree.Raw())

		if tree.GetObjType() != object.TreeObj {
			t.Errorf("expected object type to be %s but got %s", object.CommitObj, tree.GetObjType())
		}

		var buffer bytes.Buffer

		tree.Write(&buffer)

		compressedBytes := buffer.Bytes()

		testutils.AssertBytes(t, "raw compressed", TREE_COMPRESSED_BYTES, compressedBytes)
	})
}
