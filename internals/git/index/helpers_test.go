package index

import (
	"bytes"
	"fmt"
	"math"
	"testing"
)

func TestUint32Parsing(t *testing.T) {
	extra := []int{-1, 0, 1}

	for i := 0; i < 32; i++ {
		val := math.Pow(2, float64(i))

		for add := range extra {
			val := uint32(int(val) + add)

			t.Run(fmt.Sprintf("Checking for %d", val), func(t *testing.T) {
				var buffer bytes.Buffer

				_, err := writeUint32(val, &buffer)

				if err != nil {
					t.Fatalf("Failed for %d with err %v", val, err)
				}

				b := buffer.Bytes()

				parsedValue, err := parse32bit(&b, 0)

				if err != nil {
					t.Fatalf("Failed for %d with err %v", val, err)
				}

				if parsedValue != val {
					t.Fatalf("Expected %d but got %d", val, parsedValue)
				}
			})
		}
	}
}

func TestByteSliceToInt(t *testing.T) {
	TEST_VAL := 46
	b := []byte{0x00, 0x00, 0x00, 0x2e}

	val, err := byteSliceToInt(&b)

	if err != nil {
		t.Fatalf("Failed for %d with err %v", TEST_VAL, err)
	}

	if val != int64(TEST_VAL) {
		t.Fatalf("Expected %d but got %d", TEST_VAL, val)
	}
}
