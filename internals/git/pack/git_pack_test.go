package pack_test

import (
	"os"
	"testing"

	"github.com/uragirii/got/internals/git/pack"
	"github.com/uragirii/got/internals/git/sha"
)

func TestFromIdxFile(t *testing.T) {
	fsys := os.DirFS("/Users/apoorv.kansal/Codes/golang/got/testdata/pack")

	idx, _ := pack.FromIdxFile(fsys, "pack-9fd2cca459eacd57246d2ba2349866deea5ed542.idx")

	file, _ := os.Open("/Users/apoorv.kansal/Codes/golang/got/testdata/pack/pack-9fd2cca459eacd57246d2ba2349866deea5ed542.pack")

	p := pack.ParsePackFile(file, idx)

	s, _ := sha.FromString("153d60f34ff8a983fd0870b7d1a49bf434fb06f9")

	p.GetObj(s)

}
