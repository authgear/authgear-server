package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/pkg/util/cliutil"
)

func MarshalConfigYAML(ctx context.Context, cmd *cobra.Command, cfg interface{}, outputFolderPath string, fileName string) error {
	yaml, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	outputPath := filepath.Join(outputFolderPath, fileName)

	file, err := os.OpenFile(outputPath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
	if os.IsExist(err) {
		var overwrite bool
		overwrite, err = cliutil.PromptBool{
			Title:                       fmt.Sprintf("%v already exists, overwrite?", outputPath),
			InteractiveDefaultUserInput: false,
			NonInteractiveFlagName:      "overwrite",
		}.Prompt(ctx, cmd)
		if err != nil {
			return err
		}

		if !overwrite {
			fmt.Fprintf(os.Stderr, "canceled\n")
			return ErrUserCancel
		}

		file, err = os.OpenFile(outputPath, os.O_RDWR|os.O_TRUNC, 0666)
	}
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(yaml)
	fmt.Printf("config written to %v\n", outputPath)
	return err
}
