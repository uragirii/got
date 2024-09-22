package index_test

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"strings"
	"syscall"
	"testing"
	"testing/fstest"

	"github.com/uragirii/got/internals/git/index"
	testutils "github.com/uragirii/got/internals/test_utils"
)

type testCaseType struct {
	Name      string
	DebugFile string
}

const TEST_DIR = "../../../testdata/index"

func getIndexTestFiles() ([]testCaseType, error) {
	entries, err := os.ReadDir(TEST_DIR)

	if err != nil {
		return []testCaseType{}, err
	}

	if len(entries) == 0 {
		return []testCaseType{}, fmt.Errorf("Found no test files in testdata/index")
	}

	testCases := make([]testCaseType, 0, len(entries)/2)

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
	return testCases, nil
}

func TestIndexDebug(t *testing.T) {

	testCases, err := getIndexTestFiles()

	if err != nil {
		t.Fatalf("%v", err)
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

			testutils.AssertBytes(t, "debug", debugOutput, b.Bytes())
		})
	}

}

func TestIndexWrite(t *testing.T) {
	testCases, err := getIndexTestFiles()

	if err != nil {
		t.Fatalf("%v", err)
	}

	for _, testCase := range testCases {
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

			testutils.AssertBytes(t, "write", actualBytes, gotBytes)

		})

	}
}

func TestIndexAdd(t *testing.T) {
	index.SysStat = func(path string, stat *syscall.Stat_t) (err error) {
		if path == "testfile.txt" {
			t.Logf("Overiding syscall.Stat")
			stat.Ctimespec = syscall.Timespec{
				Sec:  1723625479,
				Nsec: 560332952,
			}
			stat.Mtimespec = syscall.Timespec{
				Sec:  1723625479,
				Nsec: 560332952,
			}

			stat.Dev = 16777232
			stat.Ino = 36578468
			stat.Uid = 502
			stat.Gid = 20
			stat.Size = 19

			return nil
		}

		return fmt.Errorf("invalid file name %s", path)
	}

	indexFile, err := os.Open(path.Join(TEST_DIR, "normal"))

	if err != nil {
		t.Fatalf("%v", err)
	}

	addedFile, err := os.Open(path.Join(TEST_DIR, "added.debug"))

	if err != nil {
		t.Fatalf("%v", err)
	}

	i, err := index.New(indexFile)

	if err != nil {
		t.Fatalf("%v", err)
	}

	mapFs := fstest.MapFS(fstest.MapFS{
		"testfile.txt": {Data: []byte("this is a test file")},
	})

	err = i.Add([]string{"testfile.txt"}, mapFs)

	if err != nil {
		t.Fatalf("%v", err)
	}

	var b bytes.Buffer
	var addedFilesBuffer bytes.Buffer

	addedFilesBuffer.ReadFrom(addedFile)
	i.Debug(&b)

	addedFileBytes := addedFilesBuffer.String()
	gotBytes := b.String()

	testutils.AssertString(t, "debug", addedFileBytes, gotBytes)

}
