package fs

import (
	"io"
	"os"

	"github.com/spf13/afero"
)

type File interface {
	io.Closer
	io.Reader
	io.ReaderAt
	io.Seeker

	Name() string
	Stat() (os.FileInfo, error)
}

type Fs interface {
	Open(name string) (File, error)
}

type AferoFs struct {
	Fs afero.Fs
}

func (f *AferoFs) Open(name string) (File, error) {
	return f.Fs.Open(name)
}
