package config

import (
	"fmt"
	"os"
	"path/filepath"

	"sigs.k8s.io/yaml"
)

func MarshalConfigYAML(cfg interface{}, outputFolderPath string, fileName string) error {
	yaml, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	if outputFolderPath == "-" {
		_, err = os.Stdout.Write(yaml)
		return err
	}

	outputPath := filepath.Join(outputFolderPath, fileName)

	file, err := os.OpenFile(outputPath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
	if os.IsExist(err) {
		overwrite := promptBool{
			Title:        fmt.Sprintf("%s already exists, overwrite?", outputPath),
			DefaultValue: false,
		}.Prompt()
		if !overwrite {
			fmt.Println("cancelled")
			return ErrUserCancel
		}
		file, err = os.Create(outputPath)
	}
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(yaml)
	fmt.Printf("config written to %s\n", outputPath)
	return err
}
