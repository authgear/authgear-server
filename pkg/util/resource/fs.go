package resource

import (
	"io"
	"io/ioutil"
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
	Readdir(n int) ([]os.FileInfo, error)
}

type Fs interface {
	Open(name string) (File, error)
}

type AferoFs struct {
	Fs afero.Fs
}

func (f AferoFs) Open(name string) (File, error) {
	return f.Fs.Open(name)
}

func ReadFile(fs Fs, path string) ([]byte, error) {
	file, err := fs.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return ioutil.ReadAll(file)
}
