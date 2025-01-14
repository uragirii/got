package cmd

import (
	"fmt"

	"github.com/uragirii/got/internals"
	"github.com/uragirii/got/internals/git/transport/http"
)

var CLONE *internals.Command = &internals.Command{
	Name:  "clone",
	Desc:  "Clones the repository into a folder",
	Flags: []*internals.Flag{},
	Run:   Clone,
}

func Clone(c *internals.Command, _ string) {
	if len(c.Args) == 0 {
		panic("expected more arguments")
	}

	cap, err := http.CapabilityAdvertisement(c.Args[0])

	if err != nil {
		panic(err)
	}

	fmt.Println(cap)

}
