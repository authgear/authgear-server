package config_test

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

func TestOAuthClientConfigXFramework(t *testing.T) {
	cases := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid react", "react", false},
		{"valid vue", "vue", false},
		{"valid angular", "angular", false},
		{"valid nextjs", "nextjs", false},
		{"valid express", "express", false},
		{"valid other-spa", "other-spa", false},
		{"valid flask", "flask", false},
		{"valid laravel", "laravel", false},
		{"valid java", "java", false},
		{"valid aspnet", "aspnet", false},
		{"valid other-oidc", "other-oidc", false},
		{"valid react-native", "react-native", false},
		{"valid ios", "ios", false},
		{"valid android", "android", false},
		{"valid flutter", "flutter", false},
		{"valid ionic", "ionic", false},
		{"empty allowed", "", false},
		{"unknown value rejected", "not-a-framework", true},
	}

	Convey("TestOAuthClientConfigXFramework", t, func() {
		for _, tc := range cases {
			tc := tc
			Convey(tc.name, func() {
				doc := map[string]any{
					"client_id":          "test",
					"name":               "Test",
					"x_application_type": "spa",
					"redirect_uris":      []string{"https://example.com/cb"},
				}
				if tc.value != "" {
					doc["x_framework"] = tc.value
				}
				data, err := json.Marshal(doc)
				So(err, ShouldBeNil)

				ctx := context.Background()
				err = config.Schema.PartValidator("OAuthClientConfig").Validate(ctx, bytes.NewReader(data))
				if tc.wantErr {
					So(err, ShouldBeError)
				} else {
					So(err, ShouldBeNil)
				}
			})
		}
	})
}
