package internals_test

import (
	"testing"

	"github.com/uragirii/got/internals"
)

func TestGetGitDir(t *testing.T) {
	tempDir := t.TempDir()
	// Should recognise the env var set
	t.Setenv("GIT_DIR", tempDir)

	gitDir, err := internals.GetGitDir()

	if err != nil {
		t.Fatalf("GetGitDir failed with err %v", err)
	}

	if gitDir != tempDir {
		t.Fatalf("GetGitDir expected %s, found %s", tempDir, gitDir)
	}

}
