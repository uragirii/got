package pack

import (
	"fmt"
	"strconv"
)

func fourBytesToInt(b []byte) (int64, error) {
	return strconv.ParseInt(fmt.Sprintf("%x", b), 16, 32)
}
