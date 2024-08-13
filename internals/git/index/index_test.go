package index_test

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/uragirii/got/internals/git/index"
	testutils "github.com/uragirii/got/internals/test_utils"
)

const TEST_DIR = "../../../testdata/index"

func TestIndex(t *testing.T) {
	entries, err := os.ReadDir(TEST_DIR)

	if err != nil {
		t.Fatalf("%v", err)
	}

	if len(entries) == 0 {
		t.Fatalf("Found no test files in testdata/index")
	}

	testCases := make([]struct {
		Name      string
		DebugFile string
	}, 0, len(entries)/2)

	for _, e := range entries {
		name := e.Name()

		if !strings.HasSuffix(name, ".debug") {
			testCases = append(testCases, struct {
				Name      string
				DebugFile string
			}{
				Name:      name,
				DebugFile: fmt.Sprintf("%s.debug", name),
			})
		}

	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			debugOutput, err := os.ReadFile(path.Join(TEST_DIR, testCase.DebugFile))

			if err != nil {
				t.Fatalf("File %s failed with %v", testCase.Name, err)
			}

			indexFile, err := os.Open(path.Join(TEST_DIR, testCase.Name))

			if err != nil {
				t.Fatalf("File %s failed with %v", testCase.Name, err)
			}

			i, err := index.New(indexFile)

			if err != nil {
				t.Fatalf("File %s failed with %v", testCase.Name, err)
			}

			var b bytes.Buffer

			i.Debug(&b)

			isSame := bytes.Compare(debugOutput, b.Bytes())

			if isSame != 0 {
				t.Log(testutils.Diff(string(debugOutput), b.String()))
				t.Fatal("Debug output doesnt match")
			}
		})

		t.Run(fmt.Sprintf("Checks write for %s", testCase.Name), func(t *testing.T) {
			indexFile, err := os.Open(path.Join(TEST_DIR, testCase.Name))

			if err != nil {
				t.Fatalf("File %s failed with %v", testCase.Name, err)
			}

			i, err := index.New(indexFile)

			if err != nil {
				t.Fatalf("File %s failed with %v", testCase.Name, err)
			}

			var b bytes.Buffer

			i.Write(&b)

			indexFile, err = os.Open(path.Join(TEST_DIR, testCase.Name))

			if err != nil {
				t.Fatalf("File %s failed with %v", testCase.Name, err)
			}

			var indexFileBuffer bytes.Buffer

			indexFileBuffer.ReadFrom(indexFile)

			actualBytes := indexFileBuffer.Bytes()
			gotBytes := b.Bytes()

			isSame := bytes.Compare(actualBytes, gotBytes)

			if isSame != 0 {
				t.Logf(testutils.DiffBytes(&actualBytes, &gotBytes))
				t.Fatal("Write output doesnt match")
			}

		})

	}

}
