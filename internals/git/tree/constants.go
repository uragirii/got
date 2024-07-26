package tree

import "fmt"

type Mode string

const (
	ModeNormal     Mode = "100644"
	ModeExecutable Mode = "100755"
	// Deprecated: Sym link is not supported at the moment
	ModeSymLink Mode = "120000"
	ModeDir     Mode = "40000"
)

var ErrInvalidTree = fmt.Errorf("invalid tree")

func (mode Mode) Pretty() string {
	if mode == ModeDir {
		return fmt.Sprintf("0%s tree", mode)
	}
	return fmt.Sprintf("%s blob", mode)
}
