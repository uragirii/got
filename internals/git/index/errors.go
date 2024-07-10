package index

import "fmt"

var ErrInvalidIndex = fmt.Errorf("invalid index file")
var ErrVersionNotSupported = fmt.Errorf("index file version not supported")
var ErrCorruptedIndex = fmt.Errorf("index file corrupted")

var ErrInvalidEntryMode = fmt.Errorf("entry mode is invalid")
