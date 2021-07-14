package main

import (
	"errors"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	"github.com/authgear/authgear-server/pkg/util/debug"
	"github.com/authgear/authgear-server/pkg/version"
)

var cmdRoot = &cobra.Command{
	Use:     "authgear-portal",
	Version: version.Version,
}

func init() {
	cmdRoot.AddCommand(cmdStart)
	cmdRoot.AddCommand(cmdDatabase)
	cmdRoot.AddCommand(cmdInternal)
	cmdRoot.AddCommand(cmdPricing)
}

func main() {
	debug.TrapSIGQUIT()

	err := godotenv.Load()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Printf("failed to load .env file: %s", err)
	}

	_ = cmdRoot.Execute()
}
