package pktline_test

import (
	"strings"
	"testing"

	"github.com/uragirii/got/internals/git/pktline"
	testutils "github.com/uragirii/got/internals/test_utils"
)

func TestEncode(t *testing.T) {
	t.Run("returns flush packet in case of empty string", func(t *testing.T) {
		encoded := pktline.Encode("")

		testutils.AssertString(t, "encoded", pktline.FlushPacket, encoded)
	})

	t.Run("binary data doesnt include LF", func(t *testing.T) {
		encoded := pktline.EncodeBinary([]byte{0x00, 0x30, 0x5, 0x67, 0x63})

		if strings.HasSuffix(encoded, "\n") {
			t.Errorf("binary data shouldn't include LF at the end")
		}

	})

	t.Run("ascii data includes LF at the end", func(t *testing.T) {
		encoded := pktline.Encode("some random string")

		if !strings.HasSuffix(encoded, "\n") {
			t.Errorf("ascii data should include LF at the end")
		}
	})
}

func TestDecode(t *testing.T) {
	t.Run("0006a\n", func(t *testing.T) {
		decoded, err := pktline.Decode("0006a\n")

		if err != nil {
			t.Errorf("failed with error %v", err)
		}

		testutils.AssertString(t, "decoded", "a\n", string(*decoded))
	})
	t.Run("0005a", func(t *testing.T) {
		decoded, err := pktline.Decode("0005a")

		if err != nil {
			t.Errorf("failed with error %v", err)
		}

		testutils.AssertBytes(t, "decoded", []byte{byte('a')}, *decoded)
	})
	t.Run("000bfoobar\n", func(t *testing.T) {
		decoded, err := pktline.Decode("000bfoobar\n")

		if err != nil {
			t.Errorf("failed with error %v", err)
		}

		testutils.AssertString(t, "decoded", "foobar\n", string(*decoded))
	})

	t.Run(pktline.FlushPacket, func(t *testing.T) {
		decoded, err := pktline.Decode(pktline.FlushPacket)

		if err != nil {
			t.Errorf("failed with error %v", err)
		}

		testutils.AssertString(t, "decoded", "", string(*decoded))
	})
}
