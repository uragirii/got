package tree

import (
	"io/fs"
	"path"
	"sync"

	"github.com/uragirii/got/internals"
	"github.com/uragirii/got/internals/git"
	"github.com/uragirii/got/internals/git/blob"
)

func getModeFromAbsPath(fsys fs.FS, absPath string) Mode {
	fileInfo, _ := fs.Stat(fsys, absPath)

	if !fileInfo.Mode().IsRegular() {
		return ModeExecutable
	}

	return ModeNormal
}

func getTreeForDir(fsys fs.FS, dirPath string, ignore *git.Ignore) (*Tree, error) {
	items, err := fs.ReadDir(fsys, dirPath)

	if err != nil {
		return nil, err
	}

	entries := make([]TreeEntry, 0, len(items))

	ignoreFile, err := ignore.WithFile(path.Join(dirPath, git.GIT_IGNORE), fsys)

	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup

	for idx, entry := range items {

		absPath := path.Join(dirPath, entry.Name())

		if ignoreFile.Match(absPath) {
			continue
		}

		if entry.IsDir() {

			if entry.Name() != ".git" {
				wg.Add(1)
				go func(entry fs.DirEntry, idx int) {
					defer wg.Done()

					subTree, err := getTreeForDir(fsys, absPath, ignoreFile)

					if err != nil {
						panic(err)
					}

					entries = append(entries, TreeEntry{
						Mode: ModeDir,
						Name: entry.Name(),
						SHA:  subTree.SHA,
						tree: subTree,
					})

				}(entry, idx)

			}
		} else {
			wg.Add(1)
			// is a file
			go func(entry fs.DirEntry, idx int) {
				defer wg.Done()

				blobFile, err := fsys.Open(absPath)

				if err != nil {
					panic(err)
				}

				defer blobFile.Close()

				obj, err := blob.FromFile(blobFile)

				if err != nil {
					panic(err)
				}

				entries = append(entries, TreeEntry{
					Name: entry.Name(),
					Mode: getModeFromAbsPath(fsys, absPath),
					SHA:  obj.GetSHA(),
				})

			}(entry, idx)
		}
	}

	wg.Wait()

	return FromEnteries(entries)
}

func FromDir(fsys fs.FS) (*Tree, error) {

	gitDir, err := internals.GetGitDir()

	if err != nil {
		return nil, err
	}

	rootDir := path.Join(gitDir, "..")

	ignore, err := git.NewIgnore(path.Join(rootDir, git.GIT_IGNORE), fsys)

	if err != nil {
		return nil, err
	}

	return getTreeForDir(fsys, rootDir, ignore)
}
