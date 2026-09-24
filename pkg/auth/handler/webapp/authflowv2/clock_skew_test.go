package authflowv2

import (
	"context"
	"testing"
	"time"

	"github.com/spf13/afero"

	runtimeresource "github.com/authgear/authgear-server"
	v2viewmodels "github.com/authgear/authgear-server/pkg/auth/handler/webapp/authflowv2/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/authgear/authgear-server/pkg/util/template"

	. "github.com/smartystreets/goconvey/convey"
)

func renderClockSkewPage(t *testing.T, dpopEnabled bool) string {
	t.Helper()

	manager := resource.NewManagerWithDir(resource.NewManagerWithDirOptions{
		Registry:              resource.DefaultRegistry,
		BuiltinResourceFS:     runtimeresource.EmbedFS_resources_authgear,
		BuiltinResourceFSRoot: runtimeresource.RelativePath_resources_authgear,
	})

	// Keep the real clock skew check but skip the rest of the page frame.
	fs := afero.NewMemMapFs()
	err := fs.MkdirAll("templates/en/web/authflowv2", 0o777)
	if err != nil {
		t.Fatal(err)
	}
	err = afero.WriteFile(fs, "templates/en/web/authflowv2/__page_frame.html", []byte(`{{ define "authflowv2/__page_frame.html" }}{{ template "authflowv2/__clock_skew_check.html" . }}{{ template "page-content" . }}{{ end }}`), 0o666)
	if err != nil {
		t.Fatal(err)
	}
	manager = manager.Overlay(resource.LeveledAferoFs{
		Fs:      fs,
		FsLevel: resource.FsLevelCustom,
	})

	engine := &template.Engine{
		Resolver: &template.Resolver{
			Resources:             manager,
			DefaultLanguageTag:    template.DefaultLanguageTag("en"),
			SupportedLanguageTags: template.SupportedLanguageTags{"en"},
		},
	}

	data := map[string]any{}
	now := time.Date(2026, 9, 24, 0, 0, 0, 0, time.UTC)
	viewmodels.Embed(data, v2viewmodels.NewClockSkewViewModel(dpopEnabled, now))

	result, err := engine.Render(context.Background(), TemplateWebAuthflowClockSkewHTML, []string{"en"}, data)
	if err != nil {
		t.Fatal(err)
	}
	return result.String
}

func TestClockSkewTemplates(t *testing.T) {
	Convey("clock skew page", t, func() {
		html := renderClockSkewPage(t, false)
		So(html, ShouldContainSubstring, "Incorrect Device Time")
		So(html, ShouldContainSubstring, `href="/login"`)
		So(html, ShouldContainSubstring, "Back to Login")
	})

	Convey("clock skew check is rendered only when DPoP is enabled", t, func() {
		So(renderClockSkewPage(t, false), ShouldNotContainSubstring, `data-controller="clock-skew"`)

		html := renderClockSkewPage(t, true)
		So(html, ShouldContainSubstring, `data-controller="clock-skew"`)
		So(html, ShouldContainSubstring, `data-clock-skew-server-time-value="1790208000000"`)
		So(html, ShouldContainSubstring, `data-clock-skew-max-ahead-value="300000"`)
		So(html, ShouldContainSubstring, `data-clock-skew-max-behind-value="300000"`)
		So(html, ShouldContainSubstring, `data-clock-skew-redirect-url-value="/authflow/v2/clock_skew"`)
	})
}
