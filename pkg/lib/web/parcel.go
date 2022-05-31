package web

import (
	"github.com/authgear/authgear-server/pkg/lib/web_parcel"
)

func GetHashedName(file string) string {
	n := web_parcel.ParcelAssetMap[file]
	if n == "" {
		return file
	}
	return n
}
