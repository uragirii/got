package blob_test

import (
	"bytes"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/uragirii/got/internals/git/blob"
	"github.com/uragirii/got/internals/git/object"
	"github.com/uragirii/got/internals/git/sha"
)

func TestObject(t *testing.T) {

	// Example is taken from https://git-scm.com/book/en/v2/Git-Internals-Git-Objects
	TEST_STR := "what is up, doc?"
	DATA := strings.NewReader(TEST_STR)
	DATA_SHA_STR := "bd9dbf5aae1a3862dd1526723246b20206e5fc37"
	DATA_COMPRESSED_BYTES := []byte{0x78, 0x9c, 0x4a, 0xca, 0xc9, 0x4f, 0x52, 0x30, 0x34, 0x63, 0x28, 0xcf, 0x48, 0x2c, 0x51, 0xc8,
		0x2c, 0x56, 0x28, 0x2d, 0xd0, 0x51, 0x48, 0xc9, 0x4f, 0xb6, 0x07, 0x04, 0x00, 0x00, 0xff, 0xff,
		0x5f, 0x1c, 0x07, 0x9d}

	t.Run("hashes and reads the file correctly", func(t *testing.T) {

		obj, err := blob.FromFile(DATA)

		if err != nil {
			t.Error(err)
		}

		if obj.GetSHA().MarshallToStr() != DATA_SHA_STR {
			t.Errorf("expected the sha to be %s but got %s", DATA_SHA_STR, obj.GetSHA().MarshallToStr())
		}

		var objFile bytes.Buffer

		obj.Write(&objFile)

		objBytes := objFile.Bytes()

		if !bytes.Equal(objBytes, DATA_COMPRESSED_BYTES) {
			t.Errorf("compressed data not equal to expected data")
			t.Errorf("expected % x", DATA_COMPRESSED_BYTES)
			t.Errorf("got % x", objBytes)
		}

		if obj.String() != TEST_STR {
			t.Errorf("expected raw string to be %s but got %s", TEST_STR, obj.String())
		}

		if obj.GetObjType() != object.BlobObj {
			t.Errorf("expected object type to be %s but got %s", object.BlobObj, obj.GetObjType())
		}

	})

	t.Run("reads data correctly from hashed object", func(t *testing.T) {

		t.Setenv("GIT_DIR", ".git")

		objSha, _ := sha.FromString(DATA_SHA_STR)

		mapFs := fstest.MapFS{}

		objPath, _ := objSha.GetObjPath()

		mapFs[objPath] = &fstest.MapFile{Data: DATA_COMPRESSED_BYTES}

		fs := fstest.MapFS(mapFs)

		obj, err := blob.FromSHA(objSha, fs)

		if err != nil {
			t.Error(err)
		}

		if !obj.GetSHA().Eq(objSha) {
			t.Errorf("expected the sha to be %s but got %s", DATA_SHA_STR, obj.GetSHA().MarshallToStr())
		}

		if obj.String() != TEST_STR {
			t.Errorf("expected raw string to be %s but got %s", TEST_STR, obj.String())
		}

		if obj.GetObjType() != object.BlobObj {
			t.Errorf("expected object type to be %s but got %s", object.BlobObj, obj.GetObjType())
		}

	})

}
