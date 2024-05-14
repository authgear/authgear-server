package main

import (
	"errors"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "go.uber.org/automaxprocs"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	"github.com/authgear/authgear-server/cmd/portal/cmd"
	_ "github.com/authgear/authgear-server/cmd/portal/cmd/cmdanalytic"
	_ "github.com/authgear/authgear-server/cmd/portal/cmd/cmddatabase"
	_ "github.com/authgear/authgear-server/cmd/portal/cmd/cmdinternal"
	_ "github.com/authgear/authgear-server/cmd/portal/cmd/cmdpricing"
	_ "github.com/authgear/authgear-server/cmd/portal/cmd/cmdstart"
	_ "github.com/authgear/authgear-server/cmd/portal/cmd/cmdusage"
	_ "github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/adfs"
	_ "github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/apple"
	_ "github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/azureadb2c"
	_ "github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/azureadv2"
	_ "github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/facebook"
	_ "github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/github"
	_ "github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/google"
	_ "github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/linkedin"
	_ "github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/wechat"
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
