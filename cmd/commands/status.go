package commands

import (
	"fmt"
	"path"

	"github.com/uragirii/got/cmd/internals"
)

var STATUS *internals.Command = &internals.Command{
	Name:  "status",
	Desc:  "Show the working tree status",
	Flags: []*internals.Flag{},
	Run:   Status,
}

// TODO: extract to internals as might be needed
// Crawls recursively and returns all the files that should be under git's supervision
// func crawlDir(loc string, filesChan chan<- string, wg *sync.WaitGroup, gitIgnore *internals.GitIgnore) {

// 	items, err := os.ReadDir(loc)

// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}

// 	dirs := make([]string, 0, len(items))
// 	files := make([]string, 0, len(items))

// 	for _, dirItem := range items {
// 		if dirItem.IsDir() {
// 			if dirItem.Name() != ".git" {
// 				dirs = append(dirs, path.Join(loc, dirItem.Name()))
// 			}

// 		} else {

// 			filePath := path.Join(loc, dirItem.Name())

// 			files = append(files, filePath)

// 			if dirItem.Name() == ".gitignore" {
// 				gitIgnore, err = gitIgnore.WithFile(filePath, loc)

// 				if err != nil {
// 					fmt.Println(err)
// 					return
// 				}
// 			}
// 		}
// 	}

// 	for _, file := range files {
// 		if !gitIgnore.Match(file) {
// 			filesChan <- file
// 		}
// 	}

// 	for _, dir := range dirs {
// 		if !gitIgnore.Match(dir) {
// 			wg.Add(1)
// 			go func(dirpath string) {
// 				defer wg.Done()
// 				crawlDir(dirpath, filesChan, wg, gitIgnore)
// 			}(dir)
// 		}
// 	}
// }

func Status(c *internals.Command, gitPath string) {

	if gitPath == "" {
		fmt.Println("fatal: not a git repository (or any of the parent directories): .git")
		return
	}

	fmt.Println(internals.GetTreeHash(path.Join(gitPath, "..")))

	// This 10 would also force to only run 10 goroutines at a time (Hopefully)
	// filesChan := make(chan string, 10)
	// rootDir := path.Join(gitPath, "..")

	// var gitIndex internals.GitIndex

	// err := gitIndex.New(gitPath)

	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// possibleNumFiles := len(gitIndex.GetTrackedFiles())

	// files := make([]string, 0, possibleNumFiles)

	// var wg sync.WaitGroup

	// var gitIgnore internals.GitIgnore

	// crawlDir(rootDir, filesChan, &wg, &gitIgnore)

	// go func() {
	// 	wg.Wait()
	// 	close(filesChan)
	// }()

	// for file := range filesChan {
	// 	relPath, err := filepath.Rel(rootDir, file)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	files = append(files, relPath)
	// }

	// created := make([]string, 0, 10)
	// changeChan := make(chan string, 10)

	// var fwg sync.WaitGroup

	// for _, file := range files {

	// 	if gitIndex.Has(file) {
	// 		fwg.Add(1)
	// 		go func(file string) {
	// 			defer fwg.Done()
	// 			sha, _, err := internals.HashBlob(file, false)
	// 			if err != nil {
	// 				fmt.Println(err)
	// 			}

	// 			if string(sha[:]) != string((*gitIndex.Get(file)).SHA1[:]) {
	// 				changeChan <- file
	// 			}

	// 		}(file)
	// 	} else {
	// 		created = append(created, file)
	// 	}
	// }
	// go func() {
	// 	fwg.Wait()
	// 	close(changeChan)
	// }()

	// for changedF := range changeChan {
	// 	fmt.Println("Changed", changedF)
	// }

	// for _, newF := range created {
	// 	fmt.Println("Created", newF)
	// }

}
