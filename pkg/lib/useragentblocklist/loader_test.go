package useragentblocklist

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/afero"

	runtimeresource "github.com/authgear/authgear-server"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

func TestLoad(t *testing.T) {
	Convey("Load", t, func() {
		Convey("reads the embedded blocklist and matches representative bots", func() {
			manager := resource.NewManagerWithDir(resource.NewManagerWithDirOptions{
				Registry:              resource.DefaultRegistry,
				BuiltinResourceFS:     runtimeresource.EmbedFS_resources_authgear,
				BuiltinResourceFSRoot: runtimeresource.RelativePath_resources_authgear,
				CustomResourceDir:     "",
			})

			list, err := Load(manager)
			So(err, ShouldBeNil)
			So(list.IsBlocked("Googlebot"), ShouldBeTrue)
			So(list.IsBlocked("bingbot"), ShouldBeTrue)
			So(list.IsBlocked("Bytespider"), ShouldBeTrue)
			So(list.IsBlocked("CCBot"), ShouldBeTrue)
			So(list.IsBlocked("ChatGPT-User"), ShouldBeTrue)
			So(list.IsBlocked("Claude-User"), ShouldBeTrue)
			So(list.IsBlocked("Perplexity-User"), ShouldBeTrue)
			So(list.IsBlocked("anthropic-ai"), ShouldBeTrue)
			So(list.IsBlocked("cohere-ai"), ShouldBeTrue)
			So(list.IsBlocked("omgili"), ShouldBeTrue)
			So(list.IsBlocked("omgilibot"), ShouldBeTrue)
			So(list.IsBlocked("AhrefsBot"), ShouldBeTrue)
			So(list.IsBlocked("SemrushBot"), ShouldBeTrue)
			So(list.IsBlocked("MJ12bot"), ShouldBeTrue)
			So(list.IsBlocked("DotBot"), ShouldBeTrue)
			So(list.IsBlocked("rogerbot"), ShouldBeTrue)
			So(list.IsBlocked("BLEXBot"), ShouldBeTrue)
			So(list.IsBlocked("Barkrowler"), ShouldBeTrue)
			So(list.IsBlocked("PetalBot"), ShouldBeTrue)
			So(list.IsBlocked("YandexBot"), ShouldBeTrue)
			So(list.IsBlocked("Baiduspider-mobile"), ShouldBeTrue)
			So(list.IsBlocked("PerplexityBot"), ShouldBeTrue)
			So(list.IsBlocked("GPTBot"), ShouldBeTrue)
			So(list.IsBlocked("ClaudeBot"), ShouldBeTrue)
			So(list.IsBlocked("unknown-bot"), ShouldBeFalse)
		})

		Convey("includes deployment overlays", func() {
			builtinFS := afero.NewMemMapFs()
			customFS := afero.NewMemMapFs()
			err := afero.WriteFile(builtinFS, "user_agent_blocklist.txt", []byte(`/builtinbot/`), 0o644)
			So(err, ShouldBeNil)
			err = afero.WriteFile(customFS, "user_agent_blocklist.txt", []byte(`/custombot/`), 0o644)
			So(err, ShouldBeNil)

			manager := resource.NewManager(resource.DefaultRegistry, []resource.Fs{
				resource.LeveledAferoFs{Fs: builtinFS, FsLevel: resource.FsLevelBuiltin},
				resource.LeveledAferoFs{Fs: customFS, FsLevel: resource.FsLevelCustom},
			})

			list, err := Load(manager)
			So(err, ShouldBeNil)
			So(list.IsBlocked("builtinbot"), ShouldBeTrue)
			So(list.IsBlocked("custombot"), ShouldBeTrue)
			So(list.IsBlocked("otherbot"), ShouldBeFalse)
		})

		Convey("fails fast on invalid regex entries", func() {
			builtinFS := afero.NewMemMapFs()
			err := afero.WriteFile(builtinFS, "user_agent_blocklist.txt", []byte(`/\c/`), 0o644)
			So(err, ShouldBeNil)

			manager := resource.NewManager(resource.DefaultRegistry, []resource.Fs{
				resource.LeveledAferoFs{Fs: builtinFS, FsLevel: resource.FsLevelBuiltin},
			})

			_, err = Load(manager)
			So(err, ShouldBeError)
			So(err.Error(), ShouldContainSubstring, "invalid blocklist entry")
		})
	})
}
