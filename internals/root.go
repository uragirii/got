package internals

import (
	"fmt"
	"os"
	"path"
)

var ErrNoRepo = fmt.Errorf("no repo found")

func FindRoot(cwd string) (string, error) {
	wd := cwd
	for {
		if wd == "/" {
			return "", ErrNoRepo
		}
		items, err := os.ReadDir(wd)

		if err != nil {
			return "", err
		}

		for _, item := range items {
			if item.IsDir() && item.Name() == ".git" {
				return path.Join(wd, ".git"), nil
			}
		}

		wd = path.Join(wd, "..")

	}
}
