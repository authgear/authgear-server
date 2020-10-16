package resource

import "os"

type LayerFile struct {
	Path string
	Data []byte
}

type MergedFile struct {
	Data []byte
}

type Descriptor interface {
	ReadResource(fs Fs) ([]LayerFile, error)
	MatchResource(path string) bool
	Merge(layers []LayerFile, args map[string]interface{}) (*MergedFile, error)
	Parse(merged *MergedFile) (interface{}, error)
}

// SimpleFile merges files from different layers, by using the top-most
// layer available.
type SimpleFile struct {
	Name    string
	ParseFn func(data []byte) (interface{}, error)
}

func (f SimpleFile) ReadResource(fs Fs) ([]LayerFile, error) {
	data, err := ReadFile(fs, f.Name)
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return []LayerFile{{Path: f.Name, Data: data}}, nil
}

func (f SimpleFile) MatchResource(path string) bool {
	return path == f.Name
}

func (f SimpleFile) Merge(layers []LayerFile, args map[string]interface{}) (*MergedFile, error) {
	file := layers[len(layers)-1]
	return &MergedFile{Data: file.Data}, nil
}

func (f SimpleFile) Parse(merged *MergedFile) (interface{}, error) {
	if f.ParseFn == nil {
		return merged.Data, nil
	}
	return f.ParseFn(merged.Data)
}
