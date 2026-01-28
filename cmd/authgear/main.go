package main

import (
	"context"
	"errors"
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.uber.org/automaxprocs/maxprocs"

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
	_ "github.com/authgear/authgear-server/cmd/authgear/cmd/cmdsearch"
	_ "github.com/authgear/authgear-server/cmd/authgear/cmd/cmdstart"
	_ "github.com/authgear/authgear-server/pkg/latte"
	_ "github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
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
	"github.com/authgear/authgear-server/pkg/util/otelutil"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

func main() {
	_, _ = maxprocs.Set()

	debug.TrapSIGQUIT()

	err := godotenv.Load()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Printf("failed to load .env file: %s", err)
	}

	ctx := context.Background()
	ctx, shutdown, err := otelutil.SetupOTelSDKGlobally(ctx)
	if err != nil {
		log.Fatalf("failed to setup otel: %v", err)
	}
	defer func() {
		_ = shutdown(ctx)
	}()

	ctx = slogutil.Setup(ctx)

	err = cmd.Root.ExecuteContext(ctx)
	if err != nil {
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}
