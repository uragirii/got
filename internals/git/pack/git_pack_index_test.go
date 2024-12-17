package pack_test

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/uragirii/got/internals/git/object"
	"github.com/uragirii/got/internals/git/pack"
	"github.com/uragirii/got/internals/git/sha"
	"github.com/uragirii/got/testdata"
)

type VerifyPackOutput struct {
	SHA            string            `json:sha`
	Type           object.ObjectType `json:type`
	Size           int64             `json:size`
	CompressedSize uint32            `json:compressedSize`
	Offset         uint32            `json:offset`
}

const _VERBOSE_OUTPUT_PATH = "pack/verify-pack-verbose-output.json"
const _IDX_FILE_PATH = "pack/pack-9fd2cca459eacd57246d2ba2349866deea5ed542.idx"

func loadVerboseOutput(t *testing.T) ([]VerifyPackOutput, error) {
	t.Helper()

	_VerifyPackFormattedOutput, err := testdata.TestData.ReadFile(_VERBOSE_OUTPUT_PATH)

	if err != nil {
		return []VerifyPackOutput{}, err
	}

	var output []VerifyPackOutput

	err = json.Unmarshal(_VerifyPackFormattedOutput, &output)

	if err != nil {
		return []VerifyPackOutput{}, err
	}

	return output, nil
}

func TestPackIndex(t *testing.T) {

	output, err := loadVerboseOutput(t)

	if err != nil {
		t.Errorf("error while loading verbose output %v", err)
	}

	idxFile, err := pack.FromIdxFile(testdata.TestData, _IDX_FILE_PATH)

	if err != nil {
		t.Errorf("error while parsing index file %v", err)
	}

	for _, item := range output {
		t.Run(fmt.Sprintf("Checking for output %s", item.SHA), func(t *testing.T) {

			sha, err := sha.FromString(item.SHA)

			if err != nil {
				t.Errorf("error while parsing sha %v", err)
			}

			idxItem, has := idxFile.GetObjOffset(sha)

			if !has {
				t.Errorf("expected index file to contain SHA %s but not found", item.SHA)
				t.Fail()
			}

			if idxItem.Offset != item.Offset {
				t.Errorf("Expected offset to be %d but got %d", item.Offset, idxItem.Offset)
				t.Fail()
			}

			// For now, we don't know if we should fetch compressed size from index
			if idxItem.CompressedSize != 0 {
				t.Errorf("Expected compressed size to be %d but got %d", item.CompressedSize, 0)
				t.Fail()
			}
		})
	}
}
