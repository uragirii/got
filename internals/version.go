package internals

import (
	"fmt"
	"runtime/debug"
)

var Version = ""

func SetupVersion(v string) {
	info, ok := debug.ReadBuildInfo()

	if !ok {
		panic("fatal: cannot read buildinfo")
	}

	revision := ""

	for _, kv := range info.Settings {
		if kv.Key == "vcs.revision" {
			revision = kv.Value
		}
	}

	Version = fmt.Sprintf("%s-%s", v, revision[0:7])
}

func Agent() string {
	return fmt.Sprintf("got/%s", Version[1:])
}
