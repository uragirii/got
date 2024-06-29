package internals

import (
	"fmt"
	"os"
	"path"
	"sync"
)

type fileHash struct {
	Path string
	SHA  string
}

type dirTree struct {
	Name       string
	Path       string
	SHA        string
	Parent     *dirTree
	childFiles map[string]*fileHash
	childDirs  map[string]*dirTree
}

func getDirHash(dirPath string, gitIgnore *GitIgnore) (*dirTree, error) {

	items, err := os.ReadDir(dirPath)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	dirs := make([]string, 0, len(items))
	files := make([]string, 0, len(items))

	dirHash := &dirTree{
		Name:       path.Base(dirPath),
		Path:       dirPath,
		childFiles: make(map[string]*fileHash),
		childDirs:  make(map[string]*dirTree),
	}

	for _, dirItem := range items {
		if dirItem.IsDir() {
			if dirItem.Name() != ".git" {
				dirs = append(dirs, path.Join(dirPath, dirItem.Name()))
			}

		} else {

			filePath := path.Join(dirPath, dirItem.Name())

			files = append(files, filePath)

			if dirItem.Name() == ".gitignore" {
				gitIgnore, err = gitIgnore.WithFile(filePath, dirPath)

				if err != nil {
					fmt.Println(err)
					return nil, err
				}
			}
		}
	}

	var wg sync.WaitGroup

	for _, file := range files {
		if !gitIgnore.Match(file) {
			// wg.Add(1)
			// go func(file string) {
			// defer wg.Done()
			fileSHA, _, err := HashBlob(file, false)

			if err != nil {
				fmt.Printf("error while hashing %s", file)
				panic(err)
			}

			dirHash.childFiles[file] = &fileHash{
				Path: file,
				SHA:  fmt.Sprintf("%x", *fileSHA),
			}
			// }(file)
		}
	}

	for _, dir := range dirs {
		if !gitIgnore.Match(dir) {
			// wg.Add(1)
			// go func(dirpath string) {
			// defer wg.Done()
			subDirHash, err := getDirHash(dir, gitIgnore)

			subDirHash.Parent = dirHash

			if err != nil {
				fmt.Printf("error while hashing %s", dir)
				panic(err)
			}

			dirHash.childDirs[dir] = subDirHash

			// }(dir)
		}
	}

	wg.Wait()

	treeSha, err := HashTree(dirHash)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	dirHash.SHA = fmt.Sprintf("%x", *treeSha)

	return dirHash, nil
}

// Deprecated: use Tree instead
func GetTreeHash(rootDir string) *dirTree {

	var gitIgnore GitIgnore

	dirHash, err := getDirHash(rootDir, &gitIgnore)

	if err != nil {
		panic(err)
	}
	return dirHash

}
