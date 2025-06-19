package internal

import (
	"encoding/json"
	"fmt"
	"os"
)

type PackOptions struct {
	InputDirectoryPath string
}

func Pack(opts *PackOptions) (err error) {
	data, err := pack(opts.InputDirectoryPath)
	if err != nil {
		return
	}

	err = json.NewEncoder(os.Stdout).Encode(data)
	if err != nil {
		err = fmt.Errorf("failed to write to stdout: %w", err)
		return
	}

	return
}
