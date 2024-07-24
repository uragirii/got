package object_test

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/uragirii/got/internals/git/object"
	"github.com/uragirii/got/internals/git/sha"
)

func TestFromSHA(t *testing.T) {
	DATA := "data data data data"
	DATA_LEN := len(DATA)

	setupMapFs := func(uncompressedData []byte) (*sha.SHA, fs.FS) {
		var buffer bytes.Buffer

		writer := zlib.NewWriter(&buffer)

		writer.Write(uncompressedData)

		writer.Close()

		sha, _ := sha.FromData(&uncompressedData)

		mapFs := fstest.MapFS{}

		objPath, _ := sha.GetObjPath()

		mapFs[objPath] = &fstest.MapFile{Data: buffer.Bytes()}

		return sha, fstest.MapFS(mapFs)
	}

	t.Run("parses blob obj correctly", func(t *testing.T) {
		blobHeader := fmt.Sprintf(object.BlobHeader, DATA_LEN)
		uncompressedObj := []byte(fmt.Sprintf("%s%s", blobHeader, DATA))

		t.Setenv("GIT_DIR", ".git")

		sha, fsys := setupMapFs(uncompressedObj)

		contents, err := object.FromSHA(sha, fsys)

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

		t.Setenv("GIT_DIR", ".git")

		sha, fsys := setupMapFs(uncompressedObj)

		contents, err := object.FromSHA(sha, fsys)

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

		t.Setenv("GIT_DIR", ".git")

		sha, fsys := setupMapFs(uncompressedObj)

		contents, err := object.FromSHA(sha, fsys)

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
			t.Setenv("GIT_DIR", ".git")

			sha, fsys := setupMapFs(uncompressedObj)

			_, err := object.FromSHA(sha, fsys)

			if err != nil {
				return
			} else {
				t.Errorf("expected error to be thrown")
			}

		})

	})
}
