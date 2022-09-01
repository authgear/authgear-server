package filepathutil

import (
	"fmt"
	"path"
	"strings"
)

func IsSourceMapPath(filePath string) bool {
	return path.Ext(filePath) == ".map"
}

func Ext(filePath string) string {
	extension := path.Ext(filePath)
	if extension == "" {
		return ""
	}

	if IsSourceMapPath(filePath) {
		extension = fmt.Sprintf("%s%s", path.Ext(strings.TrimSuffix(filePath, extension)), extension)
	}

	return extension
}

func ParseHashedPath(hashedPath string) (filePath string, hash string, ok bool) {
	extension := Ext(hashedPath)
	if extension == "" {
		return
	}

	nameWithHash := strings.TrimSuffix(hashedPath, extension)
	dotIdx := strings.LastIndex(nameWithHash, ".")
	if dotIdx == -1 {
		// hashedPath doesn't have extension, e.g. filename.hash
		// so the extension is the hashed
		filePath = nameWithHash
		hash = strings.TrimPrefix(extension, ".")
		ok = true
		return
	}

	nameOnly := nameWithHash[:dotIdx]

	hash = nameWithHash[dotIdx+1:]
	filePath = fmt.Sprintf("%s%s", nameOnly, extension)
	ok = true

	return
}

func MakeHashedPath(filePath string, hash string) string {
	if hash == "" {
		return filePath
	}

	extension := Ext(filePath)
	if extension == "" {
		return fmt.Sprintf("%s.%s", filePath, hash)
	}

	filename := strings.TrimSuffix(filePath, extension)
	return fmt.Sprintf("%s.%s%s", filename, hash, extension)
}
