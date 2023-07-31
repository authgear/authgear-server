package main

import (
	"errors"
	"log"
	"os"

	"github.com/joho/godotenv"

	"github.com/authgear/authgear-server/cmd/authgear/cmd"
	_ "github.com/authgear/authgear-server/cmd/authgear/cmd/cmdaudit"
	_ "github.com/authgear/authgear-server/cmd/authgear/cmd/cmdbackground"
	_ "github.com/authgear/authgear-server/cmd/authgear/cmd/cmddatabase"
	_ "github.com/authgear/authgear-server/cmd/authgear/cmd/cmdimages"
	_ "github.com/authgear/authgear-server/cmd/authgear/cmd/cmdimages/cmddatabase"
	_ "github.com/authgear/authgear-server/cmd/authgear/cmd/cmdimages/cmdstart"
	_ "github.com/authgear/authgear-server/cmd/authgear/cmd/cmdimport"
	_ "github.com/authgear/authgear-server/cmd/authgear/cmd/cmdinit"
	_ "github.com/authgear/authgear-server/cmd/authgear/cmd/cmdinternal"
	_ "github.com/authgear/authgear-server/cmd/authgear/cmd/cmdstart"
	_ "github.com/authgear/authgear-server/pkg/latte"
	_ "github.com/authgear/authgear-server/pkg/lib/workflow/workflowconfig"
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
