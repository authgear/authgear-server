package internal

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/authgear/authgear-server/pkg/util/filepathutil"
)

const dirFileMode fs.FileMode = 0700
const fileFileMode fs.FileMode = 0600

type UnpackOptions struct {
	DataJSONPath        string
	OutputDirectoryPath string
}

func Unpack(opts *UnpackOptions) (err error) {
	f, err := os.Open(opts.DataJSONPath)
	if err != nil {
		err = fmt.Errorf("failed to open data JSON file: %w", err)
		return
	}
	defer f.Close()

	var data map[string]interface{}
	err = json.NewDecoder(f).Decode(&data)
	if err != nil {
		err = fmt.Errorf("failed to decode data JSON file: %w", err)
		return
	}

	_, statErr := os.Stat(opts.OutputDirectoryPath)
	if statErr == nil {
		err = fmt.Errorf("expected `%v` to not exist", opts.OutputDirectoryPath)
		return
	}
	if !errors.Is(statErr, os.ErrNotExist) {
		err = fmt.Errorf("failed to check output directory: %w", statErr)
		return
	}

	err = os.MkdirAll(opts.OutputDirectoryPath, dirFileMode)
	if err != nil {
		err = fmt.Errorf("failed to create output directory: %w", err)
		return
	}
	defer func() {
		if err != nil {
			_ = os.RemoveAll(opts.OutputDirectoryPath)
		}
	}()

	for key, value := range data {
		var path string
		path, err = filepathutil.UnescapePath(key)
		if err != nil {
			err = fmt.Errorf("failed to unescape key `%v`: %w", key, err)
			return
		}
		base64Str, ok := value.(string)
		if !ok {
			err = fmt.Errorf("expected `%v` to be a string, but found %T", key, value)
			return
		}
		var b []byte
		b, err = base64.StdEncoding.DecodeString(base64Str)
		if err != nil {
			err = fmt.Errorf("failed to base64 decode `%v` `%v`: %w", key, base64Str, err)
			return
		}

		outputPath := filepath.Join(opts.OutputDirectoryPath, path)
		outputParentPath := filepath.Dir(outputPath)
		err = os.MkdirAll(outputParentPath, dirFileMode)
		if err != nil {
			err = fmt.Errorf("failed to create directory at `%v`: %w", outputParentPath, err)
			return
		}

		err = os.WriteFile(outputPath, b, fileFileMode)
		if err != nil {
			err = fmt.Errorf("failed to write `%v`: %w", outputPath, err)
			return
		}
	}

	return
}
