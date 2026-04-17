package authflowv2

import (
	"context"
	"testing"

	"github.com/spf13/afero"

	runtimeresource "github.com/authgear/authgear-server"
	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/authgear/authgear-server/pkg/util/template"

	. "github.com/smartystreets/goconvey/convey"
)

func newFatalErrorTemplateEngine(t *testing.T) *template.Engine {
	t.Helper()

	manager := resource.NewManagerWithDir(resource.NewManagerWithDirOptions{
		Registry:              resource.DefaultRegistry,
		BuiltinResourceFS:     runtimeresource.EmbedFS_resources_authgear,
		BuiltinResourceFSRoot: runtimeresource.RelativePath_resources_authgear,
	})

	fs := afero.NewMemMapFs()
	err := fs.MkdirAll("templates/en/web/authflowv2", 0o777)
	if err != nil {
		t.Fatal(err)
	}
	err = afero.WriteFile(fs, "templates/en/web/authflowv2/__page_frame.html", []byte(`{{ define "authflowv2/__page_frame.html" }}{{ template "page-content" . }}{{ end }}`), 0o666)
	if err != nil {
		t.Fatal(err)
	}

	manager = manager.Overlay(resource.LeveledAferoFs{
		Fs:      fs,
		FsLevel: resource.FsLevelCustom,
	})

	return &template.Engine{
		Resolver: &template.Resolver{
			Resources:             manager,
			DefaultLanguageTag:    template.DefaultLanguageTag("en"),
			SupportedLanguageTags: template.SupportedLanguageTags{"en"},
		},
	}
}

func renderFatalErrorPage(t *testing.T, reason string) string {
	t.Helper()

	engine := newFatalErrorTemplateEngine(t)
	result, err := engine.Render(context.Background(), TemplateWebFatalErrorHTML, []string{"en"}, map[string]interface{}{
		"Platform": "",
		"Error": map[string]interface{}{
			"reason": reason,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	return result.String
}

func TestFatalErrorTemplateTreatsFlowNotFoundLikeInvalidSession(t *testing.T) {
	Convey("fatal_error.html", t, func() {
		invalidSession := renderFatalErrorPage(t, "WebUIInvalidSession")
		flowNotFound := renderFatalErrorPage(t, "AuthenticationFlowNotFound")

		So(flowNotFound, ShouldEqual, invalidSession)
	})
}
