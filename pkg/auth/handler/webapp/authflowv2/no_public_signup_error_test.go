package authflowv2

import (
	"context"
	"testing"

	"github.com/spf13/afero"

	runtimeresource "github.com/authgear/authgear-server"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	v2viewmodels "github.com/authgear/authgear-server/pkg/auth/handler/webapp/authflowv2/viewmodels"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/authgear/authgear-server/pkg/util/template"

	. "github.com/smartystreets/goconvey/convey"
)

func renderLoginPageWithError(t *testing.T, rawError *apierrors.APIError) string {
	t.Helper()

	manager := resource.NewManagerWithDir(resource.NewManagerWithDirOptions{
		Registry:              resource.DefaultRegistry,
		BuiltinResourceFS:     runtimeresource.EmbedFS_resources_authgear,
		BuiltinResourceFSRoot: runtimeresource.RelativePath_resources_authgear,
	})

	// Render the page content only, without the page frame.
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

	engine := &template.Engine{
		Resolver: &template.Resolver{
			Resources:             manager,
			DefaultLanguageTag:    template.DefaultLanguageTag("en"),
			SupportedLanguageTags: template.SupportedLanguageTags{"en"},
		},
	}

	data := map[string]any{
		"IdentityCandidates": []map[string]any{
			{"type": "login_id", "login_id_key": "email", "login_id_type": "email", "login_id_input_type": "email"},
		},
		"EmailLoginIDEnabled":      true,
		"LoginIDKey":               "email",
		"LoginIDInputType":         "email",
		"NonPhoneLoginIDInputType": "email",
		"NonPhoneLoginIDType":      "email",
		"FormsByIdentification":    map[string]FormViewModel{},
		"MakeURL":                  func(path string, pairs ...string) string { return path },
		"Error": map[string]any{
			"reason": rawError.Reason,
		},
		"OAuthError": v2viewmodels.NewOAuthErrorViewModel(rawError).OAuthError,
	}

	result, err := engine.Render(context.Background(), TemplateWebAuthflowLoginHTML, []string{"en"}, data)
	if err != nil {
		t.Fatal(err)
	}
	return result.String
}

func TestNoPublicSignupErrorOnLoginPage(t *testing.T) {
	Convey("AuthenticationFlowNoPublicSignup on the login page", t, func() {
		// The login page reads the error back from the session, so the kind carries no logging flag.
		rawError := &apierrors.APIError{
			Kind: apierrors.Kind{
				Name:   apierrors.Forbidden,
				Reason: declarative.NoPublicSignup.Reason,
			},
			Message: "public signup is disabled",
		}

		Convey("is shown as an OAuth error", func() {
			So(v2viewmodels.NewOAuthErrorViewModel(rawError).OAuthError, ShouldNotBeNil)
		})

		Convey("renders the signup disabled message instead of the unknown error", func() {
			html := renderLoginPageWithError(t, rawError)
			So(html, ShouldContainSubstring, "Sign up is disabled.")
			So(html, ShouldNotContainSubstring, "An unknown error occurred")
		})
	})
}
