package internal

import (
	"encoding/base64"
	"fmt"
	"io/fs"
	"os"

	"github.com/authgear/authgear-server/pkg/util/filepathutil"
)

func pack(inputDirectoryPath string) (map[string]string, error) {
	root := os.DirFS(inputDirectoryPath)
	data := make(map[string]string)
	err := fs.WalkDir(root, ".", func(path string, d fs.DirEntry, err error) error {
		// Skip directory
		if d.IsDir() {
			return nil
		}

		b, err := fs.ReadFile(root, path)
		if err != nil {
			return err
		}

		base64Str := base64.StdEncoding.EncodeToString(b)
		key := filepathutil.EscapePath(path)
		data[key] = base64Str

		return nil
	})
	if err != nil {
		err = fmt.Errorf("failed to prepare JSON: %w", err)
		return nil, err
	}

	return data, nil
}
