package main

import (
	"errors"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "go.uber.org/automaxprocs"

	cmd "github.com/authgear/authgear-server/e2e/cmd/e2e/cmd"
	"github.com/authgear/authgear-server/pkg/util/debug"
)

func main() {
	debug.TrapSIGQUIT()

	err := godotenv.Load()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Printf("failed to load .env file: %s", err)
	}

	err = cmd.Root.Execute()
	if err != nil {
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}
