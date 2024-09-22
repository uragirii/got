package testutils

import (
	"bytes"
	"strings"
	"testing"
)

func AssertString(t *testing.T, key, expected, got string) {
	t.Helper()

	comp := strings.Compare(got, expected)

	if comp != 0 {
		t.Errorf("String assertion failed for %s\n%s", key, Diff(expected, got))
	}

}

func AssertBytes(t *testing.T, key string, got, expected []byte) {
	t.Helper()

	comp := bytes.Compare(got, expected)

	if comp != 0 {
		t.Errorf("Bytes assertion failed for %s\n%s", key, DiffBytes(&expected, &got))
	}
}
