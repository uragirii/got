package internals

import "os"

var GIT_DIR string

func GetGitDir() (string, error) {
	if GIT_DIR != "" {
		return GIT_DIR, nil
	}

	cwd, err := os.Getwd()

	if err != nil {
		return "", err
	}

	GIT_DIR, err = FindRoot(cwd)

	if err != nil {
		return "", err
	}

	return GIT_DIR, nil

}
