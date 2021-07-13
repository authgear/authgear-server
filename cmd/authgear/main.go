package main

import (
	"errors"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	"github.com/authgear/authgear-server/pkg/util/debug"
	"github.com/authgear/authgear-server/pkg/version"
)

func main() {
	debug.TrapSIGQUIT()

	err := godotenv.Load()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Printf("failed to load .env file: %s", err)
	}

	_ = cmdRoot.Execute()
}

var cmdRoot = &cobra.Command{
	Use:     "authgear",
	Version: version.Version,
}

func init() {
	cmdRoot.AddCommand(cmdStart)
	cmdRoot.AddCommand(cmdInit)
	cmdRoot.AddCommand(cmdDatabase)
	cmdRoot.AddCommand(cmdInternal)
	cmdRoot.AddCommand(cmdAudit)
}
