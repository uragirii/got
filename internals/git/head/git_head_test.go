package head_test

import (
	"testing"
	"testing/fstest"

	"github.com/uragirii/got/internals/git/head"
	"github.com/uragirii/got/internals/git/sha"
)

func TestNew(t *testing.T) {

	t.Setenv("GIT_DIR", ".git")

	t.Run("it works for detached head", func(t *testing.T) {
		sha, err := sha.FromString("14201e266991676173cbd041257cf1a0d8ff3a3a")

		if err != nil {
			t.Errorf("SHA creation failed with error: %v", err)
		}

		fs := fstest.MapFS(fstest.MapFS{
			".git/HEAD": {Data: []byte("14201e266991676173cbd041257cf1a0d8ff3a3a\n")},
		})

		gitHead, err := head.New(fs)

		if err != nil {
			t.Errorf("expected error to be nil got: %v", err)
		}

		if gitHead.Mode != head.Detached {
			t.Errorf("expected mode to be detached but got: %d", gitHead.Mode)
		}

		if !gitHead.SHA.Eq(sha) {
			t.Errorf("expected SHA to be %s but got %s", gitHead.SHA.MarshallToStr(), sha.MarshallToStr())
		}
	})

	t.Run("it works for branched HEAD", func(t *testing.T) {
		sha, err := sha.FromString("14201e266991676173cbd041257cf1a0d8ff3a3a")

		if err != nil {
			t.Errorf("SHA creation failed with error: %v", err)
		}

		fs := fstest.MapFS(fstest.MapFS{
			".git/HEAD":                   {Data: []byte("ref: refs/heads/branch-name\n")},
			".git/refs/heads/branch-name": {Data: []byte("14201e266991676173cbd041257cf1a0d8ff3a3a\n")},
		})

		gitHead, err := head.New(fs)

		if err != nil {
			t.Errorf("expected error to be nil got: %v", err)
		}

		if gitHead.Mode != head.Branch {
			t.Errorf("expected mode to be branch but got: %d", gitHead.Mode)
		}

		if !gitHead.SHA.Eq(sha) {
			t.Errorf("expected SHA to be %s but got %s", gitHead.SHA.MarshallToStr(), sha.MarshallToStr())
		}
	})

	// TODO: check for Tagged head
}
