//go:build !authgearlite
// +build !authgearlite

package libmagic

import (
	"github.com/vimeo/go-magic/magic"
)

func MimeFromBytes(data []byte) string {
	return magic.MimeFromBytes(data)
}
