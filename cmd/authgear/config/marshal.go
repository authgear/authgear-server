package config

import (
	"fmt"
	"os"

	"sigs.k8s.io/yaml"
)

func MarshalConfigYAML(cfg interface{}, outputPath string) error {
	yaml, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	if outputPath == "-" {
		_, err = os.Stdout.Write(yaml)
		return err
	}

	file, err := os.OpenFile(outputPath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
	if os.IsExist(err) {
		overwrite := promptBool{
			Title:        fmt.Sprintf("%s already exists, overwrite?", outputPath),
			DefaultValue: false,
		}.Prompt()
		if !overwrite {
			fmt.Println("cancelled")
			return nil
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
