package main

import (
	"errors"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

func main() {
	err := godotenv.Load()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Printf("failed to load .env file: %s", err)
	}

	cmdRoot.Execute()
}

var cmdRoot = &cobra.Command{
	Use: "authgear",
}

func init() {
	cmdRoot.AddCommand(cmdStart)
}
