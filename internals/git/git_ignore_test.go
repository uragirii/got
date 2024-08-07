package git_test

import (
	"fmt"
	"testing"
	"testing/fstest"

	"github.com/uragirii/got/internals/git"
)

var SIMPLE_GITGNORE = []byte(`/test-folder
main
got
#comment
build`)

func TestNewIgnore(t *testing.T) {
	fsys := fstest.MapFS{
		"Users/username/Codes/golang/got/.gitignore": {Data: SIMPLE_GITGNORE},
	}

	ignore, err := git.NewIgnore("Users/username/Codes/golang/got/.gitignore", fsys)

	if err != nil {
		t.Errorf("expected to error to be nil but got %s", err)
	}

	TEST_DATA := []struct {
		Name     string
		Filepath string
		Expected bool
	}{
		{
			Name:     "Matches folder correctly",
			Filepath: "Users/username/Codes/golang/got/test-folder",
			Expected: true,
		},
		{
			Name:     "Matches folder correctly",
			Filepath: "Users/username/Codes/golang/got/test-folder/file1.txt",
			Expected: true,
		},
		{
			Name:     "Matches file and folder correctly",
			Filepath: "Users/username/Codes/golang/got/got/",
			Expected: true,
		},
		{
			Name:     "Comments are ignored",
			Filepath: "Users/username/Codes/golang/got/comment",
			Expected: false,
		},
		{
			Name:     "Doesn't match other files",
			Filepath: "Users/username/Codes/golang/got/some-file",
			Expected: false,
		},
		{
			Name:     "Matches file and folder correctly",
			Filepath: "Users/username/Codes/golang/got/got",
			Expected: true,
		}}

	for _, test := range TEST_DATA {
		name := fmt.Sprintf("%s (%s)", test.Name, test.Filepath)

		t.Run(name, func(t *testing.T) {
			got := ignore.Match(test.Filepath)
			if got != test.Expected {
				t.Errorf("expected %v but got %v", test.Expected, got)
			}
		})
	}

}
