package sha_test

import (
	"fmt"
	"testing"

	"github.com/uragirii/got/internals/git/sha"
)

var TEST_BYTE_SLICE = []byte{0x23, 0x22, 0x24, 0xfe, 0xef, 0xab, 0xbc, 0xcd, 0xde, 0xef, 0xff, 0x23, 0x22, 0x24, 0xfe, 0xef, 0xab, 0xbc, 0xcd, 0xde, 0xef, 0xff}

const TEST_SHA_STR = "232224feefabbccddeefff232224feefabbccddeefff"

func TestFromByteSlice(t *testing.T) {
	sha, err := sha.FromByteSlice(&TEST_BYTE_SLICE)

	if err != nil {
		t.Fatalf("FromByteSlice returned err %v", err)
	}

	for i, b := range *sha.GetBytes() {
		if TEST_BYTE_SLICE[i] != b {
			t.Fatalf("GetBytes returned invalid byte at index %d expected %x got %x", i, TEST_BYTE_SLICE[i], b)
		}
	}

	if sha.String() != TEST_SHA_STR {
		t.Fatalf("MarshallToStr returned incorrect string, expected %s got %s", TEST_SHA_STR, sha)
	}
}

func TestFromString(t *testing.T) {
	sha, err := sha.FromString(TEST_SHA_STR)

	if err != nil {
		t.Fatalf("FromString returned err %v", err)
	}

	for i, b := range *sha.GetBytes() {
		if TEST_BYTE_SLICE[i] != b {
			t.Fatalf("GetBytes returned invalid byte at index %d expected %x got %x", i, TEST_BYTE_SLICE[i], b)
		}
	}

	if sha.String() != TEST_SHA_STR {
		t.Fatalf("MarshallToStr returned incorrect string, expected %s got %s", TEST_SHA_STR, sha)
	}
}

func TestEq(t *testing.T) {
	sha1, err := sha.FromString(TEST_SHA_STR)

	if err != nil {
		t.Fatalf("FromString returned err %v", err)
	}

	sha2, err := sha.FromByteSlice(&TEST_BYTE_SLICE)

	if err != nil {
		t.Fatalf("FromByteSlice returned err %v", err)
	}

	if !sha1.Eq(sha2) {
		t.Fatalf("Expected two SHAs to be equal")
	}

}

func TestGetObjPath(t *testing.T) {
	t.Setenv("GIT_DIR", ".git")

	expected := fmt.Sprintf("objects/%s/%s", TEST_SHA_STR[:2], TEST_SHA_STR[2:])

	sha, err := sha.FromByteSlice(&TEST_BYTE_SLICE)

	if err != nil {
		t.Fatalf("FromByteSlice returned err %v", err)
	}

	objPath, err := sha.GetObjPath()

	if err != nil {
		t.Fatalf("GetObjPath returned err %v", err)
	}

	if objPath != expected {
		t.Fatalf("expected objPath to be %s but got %s", expected, objPath)
	}

}

func TestFromData(t *testing.T) {
	DATA := []byte("this is some random data")
	EXPECTED_SHA_STR := "11f5ec9b6ee9c64b614106db0f07bb7092db1602"

	sha, err := sha.FromData(&DATA)

	if err != nil {
		t.Fatalf("FromData returned err %v", err)
	}

	if sha.String() != EXPECTED_SHA_STR {
		t.Fatalf("expected sha to be %s but got %s", EXPECTED_SHA_STR, sha)
	}

}
