package config

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/cobra"

	libconfig "github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/cliutil"
)

func TestReadOAuthClientConfigsFromConsole(t *testing.T) {
	Convey("ReadOAuthClientConfigsFromConsole", t, func() {
		Convey("returns both clients when configured", func() {
			cmd := newOAuthClientConfigCommand()
			mustSetFlag(t, cmd, "portal-origin", "http://portal.localhost:8000")
			mustSetFlag(t, cmd, "portal-client-id", "portal")
			mustSetFlag(t, cmd, "siteadmin-client-id", "siteadmin")
			mustSetFlag(t, cmd, "siteadmin-redirect-uri", "http://localhost:3005/oauth-redirect")
			mustSetFlag(t, cmd, "siteadmin-post-logout-redirect-uri", "http://localhost:3005/")

			opts, err := ReadOAuthClientConfigsFromConsole(context.Background(), cmd)
			So(err, ShouldBeNil)
			So(opts, ShouldHaveLength, 2)
			So(opts[0].ClientID, ShouldEqual, "portal")
			So(opts[0].RedirectURI, ShouldEqual, "http://portal.localhost:8000/oauth-redirect")
			So(opts[0].PostLogoutRedirectURI, ShouldEqual, "http://portal.localhost:8000/")
			So(opts[0].ApplicationType, ShouldEqual, libconfig.OAuthClientApplicationTypeTraditionalWeb)
			So(opts[1].ClientID, ShouldEqual, "siteadmin")
			So(opts[1].RedirectURI, ShouldEqual, "http://localhost:3005/oauth-redirect")
			So(opts[1].PostLogoutRedirectURI, ShouldEqual, "http://localhost:3005/")
			So(opts[1].ApplicationType, ShouldEqual, libconfig.OAuthClientApplicationTypeSPA)
		})

		Convey("skips siteadmin when client id is empty", func() {
			cmd := newOAuthClientConfigCommand()
			mustSetFlag(t, cmd, "portal-origin", "http://portal.localhost:8000")
			mustSetFlag(t, cmd, "portal-client-id", "portal")

			opts, err := ReadOAuthClientConfigsFromConsole(context.Background(), cmd)
			So(err, ShouldBeNil)
			So(opts, ShouldHaveLength, 1)
			So(opts[0].ClientID, ShouldEqual, "portal")
		})

		Convey("generates portal client id randomly when empty", func() {
			cmd := newOAuthClientConfigCommand()
			mustSetFlag(t, cmd, "portal-origin", "http://portal.localhost:8000")

			opts, err := ReadOAuthClientConfigsFromConsole(context.Background(), cmd)
			So(err, ShouldBeNil)
			So(opts, ShouldHaveLength, 1)
			So(opts[0].ClientID, ShouldEqual, "")

			generated, err := libconfig.GenerateOAuthConfigFromOptions(&opts[0])
			So(err, ShouldBeNil)
			So(generated.ClientID, ShouldNotEqual, "")
			So(generated.ClientID, ShouldNotEqual, "portal")
		})

		Convey("requires siteadmin redirect uri when client id is provided", func() {
			cmd := newOAuthClientConfigCommand()
			mustSetFlag(t, cmd, "siteadmin-client-id", "siteadmin")

			_, err := ReadOAuthClientConfigsFromConsole(context.Background(), cmd)
			So(err, ShouldBeError, "siteadmin redirect uri is required when siteadmin-client-id is provided")
		})
	})
}

func newOAuthClientConfigCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cliutil.DefineFlagInteractive(cmd)
	if err := cmd.Flags().Set("interactive", "false"); err != nil {
		panic(err)
	}
	Prompt_PortalOrigin.DefineFlag(cmd)
	Prompt_PortalClientID.DefineFlag(cmd)
	Prompt_SiteadminClientID.DefineFlag(cmd)
	Prompt_SiteadminRedirectURI.DefineFlag(cmd)
	Prompt_SiteadminPostLogoutRedirectURI.DefineFlag(cmd)
	return cmd
}

func mustSetFlag(t *testing.T, cmd *cobra.Command, name string, value string) {
	if err := cmd.Flags().Set(name, value); err != nil {
		if t == nil {
			panic(err)
		}
		t.Fatalf("failed to set flag %s: %v", name, err)
	}
}
