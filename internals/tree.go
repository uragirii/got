package internals

import (
	"fmt"
	"path"
	"path/filepath"
	"sync"
)

type status string

const (
	Modified status = "modified"
	Created  status = "created"
	Added    status = "added"
)

type FileStatus struct {
	Status status
	Path   string
}

func CompareTree(gitTree *GitTree, currTree *dirTree, indexFile *GitIndex, statusChan chan<- *FileStatus) {
	if gitTree.SHA == currTree.SHA {
		return
	}

	gitTree.LoadChildren()

	gitDir, err := GetGitDir()

	if err != nil {
		panic(err)
	}

	rootDir := path.Join(gitDir, "..")

	for absPath := range currTree.childFiles {

		relPath, _ := filepath.Rel(rootDir, absPath)

		currSha := currTree.childFiles[absPath].SHA
		gitFile, ok := gitTree.Files[absPath]

		// TOD0: consider staged but modified after staged
		// TOD0: deleted files/ folders

		// New file
		if !ok {
			// Added in index, meaning tracked
			if indexFile.Has(relPath) {
				statusChan <- &FileStatus{
					Status: Added,
					Path:   absPath,
				}
			} else {
				statusChan <- &FileStatus{
					Status: Created,
					Path:   absPath,
				}
			}
		} else {
			gitSha := gitFile.SHA

			indexSha := fmt.Sprintf("%x", indexFile.Get(relPath).SHA1)

			// modified
			if currSha != gitSha {
				if currSha == indexSha {
					// added
					statusChan <- &FileStatus{
						Status: Added,
						Path:   absPath,
					}
				} else {
					// modefied not added
					statusChan <- &FileStatus{
						Status: Modified,
						Path:   absPath,
					}
				}
			}

		}

	}

	var wg sync.WaitGroup

	for absPath, subDir := range currTree.childDirs {
		subGitTree, ok := gitTree.SubTrees[absPath]

		if !ok {
			// new folder
			statusChan <- &FileStatus{
				Status: Created,
				Path:   absPath,
			}
		} else {
			wg.Add(1)
			go func(subGitTree *GitTree, subTree *dirTree) {
				defer wg.Done()
				CompareTree(subGitTree, subTree, indexFile, statusChan)
			}(subGitTree, subDir)
		}
	}

	wg.Wait()

}
