package pack_test

import (
	"errors"
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

	s, _ := sha.FromString("d48bc2b17ec2a27d73345d8c5a36dcc08f3667e7")

	_, err := p.GetObj(s)

	if !errors.Is(err, pack.ErrOFSDeltaNotImplemented) {
		t.Fatalf("Expected OFS delta to be not implemented")
	}

	// if err != nil {
	// 	t.Fatalf("failed with error %v", err)
	// }

	// if obj.ObjType != object.CommitObj {
	// 	t.Errorf("Expected object to be commit but got %s", obj.ObjType)
	// }

}
