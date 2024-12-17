package pack_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/uragirii/got/internals/git/pack"
	"github.com/uragirii/got/internals/git/sha"
	"github.com/uragirii/got/testdata"
)

const _PACK_FILE_PATH = "pack/pack-9fd2cca459eacd57246d2ba2349866deea5ed542.pack"

func getPackFileReader(t *testing.T) (*bytes.Reader, error) {
	t.Helper()

	file, _ := testdata.TestData.Open(_PACK_FILE_PATH)

	b, err := io.ReadAll(file)

	if err != nil {
		return nil, err
	}

	r := bytes.NewReader(b)

	return r, nil

}

func TestFromIdxFile(t *testing.T) {

	idx, err := pack.FromIdxFile(testdata.TestData, _IDX_FILE_PATH)

	if err != nil {
		t.Fatalf("error while parsing index file %v", err)
	}

	packReader, err := getPackFileReader(t)

	if err != nil {
		t.Fatalf("error while reading pack file %v", err)
	}

	output, err := loadVerboseOutput(t)

	if err != nil {
		t.Fatalf("error while reading output file %v", err)
	}

	p := pack.ParsePackFile(*packReader, idx)

	for _, item := range output[0:200:200] {
		t.Run(fmt.Sprintf("Testing for %s", item.SHA), func(t *testing.T) {
			sha, err := sha.FromString(item.SHA)

			if err != nil {
				t.Errorf("error while parsing sha %v", err)
			}

			obj, err := p.GetObj(sha)

			// TODO: check for OFS Delta
			if err != nil {
				if errors.Is(err, pack.ErrOFSDeltaNotImplemented) {
					t.Skip()
				} else {
					t.Errorf("expected not an error but got %v", err)
				}
			}

			if obj.ObjType != item.Type {
				t.Errorf("expected type to be %s but got %s", item.Type, obj.ObjType)
			}
		})
	}
}
