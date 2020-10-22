package resource

import (
	"io"
	"io/ioutil"
	"os"
	"path"

	"github.com/spf13/afero"
)

type File interface {
	io.Closer
	io.Reader
	io.ReaderAt
	io.Seeker

	Name() string
	Stat() (os.FileInfo, error)
	// Readdir does not follow symlinks, instead use Readdirnames to avoid surprises.
	Readdirnames(n int) ([]string, error)
}

type Fs interface {
	Open(name string) (File, error)
	Stat(name string) (os.FileInfo, error)
}

type AferoFs struct {
	Fs afero.Fs
}

func (f AferoFs) Open(name string) (File, error) {
	return f.Fs.Open(name)
}

func (f AferoFs) Stat(name string) (os.FileInfo, error) {
	return f.Fs.Stat(name)
}

func ReadFile(fs Fs, path string) ([]byte, error) {
	file, err := fs.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return ioutil.ReadAll(file)
}

func ReadDirNames(fs Fs, dir string) ([]string, error) {
	f, err := fs.Open(dir)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return f.Readdirnames(0)
}

func ListFiles(fs Fs) ([]string, error) {
	var paths []string

	var list func(dir string) error
	list = func(dir string) error {
		files, err := ReadDirNames(fs, dir)
		if err != nil {
			return err
		}

		for _, f := range files {
			p := path.Join(dir, f)
			f, err := fs.Stat(p)
			if err != nil {
				return err
			}

			if f.IsDir() {
				if err := list(p); err != nil {
					return err
				}
				continue
			}
			paths = append(paths, p)
		}
		return nil
	}

	if err := list(""); err != nil {
		return nil, err
	}

	return paths, nil
}
