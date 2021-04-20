package main

import (
	"errors"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	"github.com/authgear/authgear-server/pkg/version"
)

var cmdRoot = &cobra.Command{
	Use:     "authgear-portal",
	Version: version.Version,
}

func init() {
	cmdRoot.AddCommand(cmdStart)
	cmdRoot.AddCommand(cmdMigrate)
	cmdRoot.AddCommand(cmdInternal)
}

func main() {
	err := godotenv.Load()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Printf("failed to load .env file: %s", err)
	}

	_ = cmdRoot.Execute()
}
