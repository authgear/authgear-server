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
	// Readdir does not follow symlinks, use Readdirnames instead to avoid surprises.
	Readdir(count int) ([]os.FileInfo, error)
	Readdirnames(n int) ([]string, error)
}

type FsLevel int

const (
	FsLevelBuiltin FsLevel = iota + 1
	FsLevelCustom
	FsLevelApp
)

type Fs interface {
	Open(name string) (File, error)
	Stat(name string) (os.FileInfo, error)
	GetFsLevel() FsLevel
}

type LeveledAferoFs struct {
	Fs      afero.Fs
	FsLevel FsLevel
}

func (f LeveledAferoFs) Open(name string) (File, error) {
	return f.Fs.Open(name)
}

func (f LeveledAferoFs) Stat(name string) (os.FileInfo, error) {
	return f.Fs.Stat(name)
}

func (f LeveledAferoFs) GetFsLevel() FsLevel {
	return f.FsLevel
}

func ReadLocation(location Location) ([]byte, error) {
	file, err := location.Fs.Open(location.Path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return ioutil.ReadAll(file)
}

func readDirNames(fs Fs, dir string) ([]string, error) {
	f, err := fs.Open(dir)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return f.Readdirnames(0)
}

func EnumerateAllLocations(fs Fs) ([]Location, error) {
	var locations []Location
	var list func(dir string) error
	list = func(dir string) error {
		files, err := readDirNames(fs, dir)
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
			locations = append(locations, Location{
				Fs:   fs,
				Path: p,
			})
		}
		return nil
	}

	if err := list(""); err != nil {
		return nil, err
	}

	return locations, nil
}
