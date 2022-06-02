package web

import (
	"github.com/authgear/authgear-server/pkg/lib/webparcel"
)

func GetHashedName(file string) string {
	n := webparcel.ParcelAssetMap[file]
	if n == "" {
		return file
	}
	return n
}
