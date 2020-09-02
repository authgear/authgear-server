package httputil

import (
	"errors"
	"net/http"
	"os"
)

type TryFileSystem struct {
	Fallback string
	FS       http.FileSystem
}

func (fs *TryFileSystem) Open(name string) (file http.File, err error) {
	file, err = fs.FS.Open(name)

	if errors.Is(err, os.ErrNotExist) {
		file, err = fs.FS.Open(fs.Fallback)
	} else if err != nil {
		return
	}

	return
}
