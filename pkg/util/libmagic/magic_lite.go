//go:build authgearlite
// +build authgearlite

package libmagic

import (
	"fmt"
)

func MimeFromBytes(data []byte) string {
	panic(fmt.Errorf("libmagic is not available in lite build"))
}
