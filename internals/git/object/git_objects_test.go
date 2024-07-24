package object_test

import (
	"fmt"
	"testing"

	"github.com/uragirii/got/internals/git/object"
)

func TestGetContents(t *testing.T) {
	DATA := "data data data data"
	DATA_LEN := len(DATA)

	t.Run("parses blob obj correctly", func(t *testing.T) {
		blobHeader := fmt.Sprintf(object.BlobHeader, DATA_LEN)
		uncompressedObj := []byte(fmt.Sprintf("%s%s", blobHeader, DATA))

		contents, err := object.GetContents(&uncompressedObj)

		if err != nil {
			t.Errorf("expected err to be nil but found %v", err)
		}

		if contents.ObjType != object.BlobObj {
			t.Errorf("expected BlobObj but found %s", contents.ObjType)
		}

		if string(*contents.Contents) != DATA {
			t.Errorf("expected data to be %s but got %s", DATA, *contents.Contents)
		}

	})

	t.Run("parses commit obj correctly", func(t *testing.T) {
		commitHeader := fmt.Sprintf(object.CommitHeader, DATA_LEN)
		uncompressedObj := []byte(fmt.Sprintf("%s%s", commitHeader, DATA))

		contents, err := object.GetContents(&uncompressedObj)

		if err != nil {
			t.Errorf("expected err to be nil but found %v", err)
		}

		if contents.ObjType != object.CommitObj {
			t.Errorf("expected CommitObj but found %s", contents.ObjType)
		}

		if string(*contents.Contents) != DATA {
			t.Errorf("expected data to be %s but got %s", DATA, *contents.Contents)
		}

	})

	t.Run("parses tree obj correctly", func(t *testing.T) {
		treeHeader := fmt.Sprintf(object.TreeHeader, DATA_LEN)
		uncompressedObj := []byte(fmt.Sprintf("%s%s", treeHeader, DATA))

		contents, err := object.GetContents(&uncompressedObj)

		if err != nil {
			t.Errorf("expected err to be nil but found %v", err)
		}

		if contents.ObjType != object.TreeObj {
			t.Errorf("expected TreeObj but found %s", contents.ObjType)
		}

		if string(*contents.Contents) != DATA {
			t.Errorf("expected data to be %s but got %s", DATA, *contents.Contents)
		}

		t.Run("throws error when content len doesnt match", func(t *testing.T) {
			blobHeader := fmt.Sprintf(object.BlobHeader, DATA_LEN+5)
			uncompressedObj := []byte(fmt.Sprintf("%s%s", blobHeader, DATA))

			_, err := object.GetContents(&uncompressedObj)

			if err != nil {
				return
			} else {
				t.Errorf("expected error to be thrown")
			}

		})

	})
}
