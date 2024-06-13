package internals

import "fmt"

func ReadGitObject(hash string) (string, *[]byte, error) {
	decoded, err := DecodeHash(hash)

	if err != nil {
		return "", nil, err
	}

	objType, content := GetObj(decoded)

	if content == nil {
		return "", nil, fmt.Errorf("content is nil")
	}

	return objType, content, nil
}
