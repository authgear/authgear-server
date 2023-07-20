package internal

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"

	"github.com/authgear/authgear-server/pkg/util/filepathutil"
)

type PackOptions struct {
	InputDirectoryPath string
}

func Pack(opts *PackOptions) (err error) {
	root := os.DirFS(opts.InputDirectoryPath)
	data := make(map[string]interface{})
	err = fs.WalkDir(root, ".", func(path string, d fs.DirEntry, err error) error {
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
		return
	}

	err = json.NewEncoder(os.Stdout).Encode(data)
	if err != nil {
		err = fmt.Errorf("failed to write to stdout: %w", err)
		return
	}

	return
}
