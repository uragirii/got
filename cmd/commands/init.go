package commands

import (
	"fmt"
	"os"
	"path"
	"sync"

	"github.com/uragirii/got/cmd/internals"
)

var INIT *internals.Command = &internals.Command{
	Name:  "init",
	Desc:  "Create an empty Git repository",
	Flags: []*internals.Flag{},
	Run:   Init,
}

var initFoldersList = [6]string{
	"info",
	path.Join("objects", "info"),
	path.Join("objects", "pack"),
	path.Join("refs", "head"),
	path.Join("refs", "tags"),
	"hooks",
}

func createHeadFile(gitPath string) {
	headContents := []byte("ref: refs/heads/main")
	err := os.WriteFile(path.Join(gitPath, "HEAD"), headContents, 0644)

	if err != nil {
		fmt.Println(err)
		fmt.Println("error while creating HEAD file")
	}
}

func createConfigFile(gitPath string) {
	contents := []byte(`[core]
	repositoryformatversion = 0
	filemode = true
	bare = false
	logallrefupdates = true
	ignorecase = true
	precomposeunicode = true`)

	err := os.WriteFile(path.Join(gitPath, "config"), contents, 0644)

	if err != nil {
		fmt.Println(err)
		fmt.Println("error while creating config file")
	}
}

func createInfoExclude(gitPath string) {
	contents := []byte("")

	err := os.WriteFile(path.Join(gitPath, "info", "exclude"), contents, 0644)

	if err != nil {
		fmt.Println(err)
		fmt.Println("error while creating info/exclude file")
	}
}

func initFolder(gitPath string) {
	/*

			Files to create
			HEAD = ref: refs/heads/main
			config =
		[core]
		        repositoryformatversion = 0
		        filemode = true
		        bare = false
		        logallrefupdates = true
		        ignorecase = true
		        precomposeunicode = true
			description=Unnamed repository; edit this file 'description' to name the repository.
			hooks= empty dir
			info => we can skip this one
			objects/info
			objects/pack
			refs/heads
			refs/tags
	*/
	err := os.Mkdir(gitPath, 0750)

	if err != nil {
		fmt.Println(err)
		fmt.Println("error while creating .git folder")
		return
	}

	var wg sync.WaitGroup

	for _, folderLoc := range initFoldersList {
		wg.Add(1)
		go func(folderLoc string) {
			defer wg.Done()
			err := os.MkdirAll(path.Join(gitPath, folderLoc), 0750)

			if err != nil {
				// TODO: handle these with context maybe
				fmt.Println(err)
				fmt.Println("error while creating folder", folderLoc)
			}
		}(folderLoc)
	}

	wg.Wait()

	createHeadFile(gitPath)
	createInfoExclude(gitPath)
	createConfigFile(gitPath)

	fmt.Println("Empty Git repository initialized at ", gitPath)

}

func Init(_ *internals.Command, _ string) {
	cwd, err := os.Getwd()

	if err != nil {
		fmt.Println(err)
		fmt.Println("error getting the cwd")
		return
	}

	gitFolder := path.Join(cwd, ".git")

	_, statErr := os.Stat(gitFolder)

	if statErr == nil {
		fmt.Println("git already initialized")
		return
	}

	if os.IsNotExist(statErr) {
		initFolder(gitFolder)
		return
	} else {
		fmt.Println(err)
		fmt.Println("error checking .git folder")
		return
	}

}
