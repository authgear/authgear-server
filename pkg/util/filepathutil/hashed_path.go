package filepathutil

import (
	"fmt"
	"path"
	"strings"
)

func ParseHashedPath(hashedPath string) (filePath string, hash string, ok bool) {
	extension := path.Ext(hashedPath)
	if extension == "" {
		return
	}

	if IsSourceMapPath(hashedPath) {
		extension = fmt.Sprintf("%s%s", path.Ext(strings.TrimSuffix(hashedPath, extension)), extension)
	}

	nameWithHash := strings.TrimSuffix(hashedPath, extension)
	dotIdx := strings.LastIndex(nameWithHash, ".")
	if dotIdx == -1 {
		// hashedPath doesn't have extension, e.g. filename.hash
		// so the extension is the hashed
		filePath = nameWithHash
		hash = strings.TrimPrefix(extension, ".")
		return
	}

	nameOnly := nameWithHash[:dotIdx]

	hash = nameWithHash[dotIdx+1:]
	filePath = fmt.Sprintf("%s%s", nameOnly, extension)

	return
}

func IsSourceMapPath(filePath string) bool {
	return path.Ext(filePath) == ".map"
}
