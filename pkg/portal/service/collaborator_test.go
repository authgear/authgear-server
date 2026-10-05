package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
	"github.com/authgear/authgear-server/pkg/portal/model"
)

type fakeCollaboratorAdminAPI struct {
	serverURL *url.URL
}

func (f *fakeCollaboratorAdminAPI) SelfDirector(ctx context.Context, actorUserID string, usage Usage) (func(*http.Request), error) {
	return func(r *http.Request) {
		r.URL.Scheme = f.serverURL.Scheme
		r.URL.Host = f.serverURL.Host
	}, nil
}

type fakeCollaboratorAppConfigs struct {
	emailConfig *config.LoginIDEmailConfig
}

func (f *fakeCollaboratorAppConfigs) ResolveContext(ctx context.Context, appID string, fn func(context.Context, *config.AppContext) error) error {
	return fn(ctx, &config.AppContext{
		Config: &config.Config{
			AppConfig: &config.AppConfig{
				Identity: &config.IdentityConfig{
					LoginID: &config.LoginIDConfig{
						Types: &config.LoginIDTypesConfig{
							Email: f.emailConfig,
						},
					},
				},
			},
		},
	})
}

func TestCollaboratorServiceCheckInviteeEmail(t *testing.T) {
	Convey("CheckInviteeEmail", t, func() {
		ctx := context.Background()

		var actorEmail string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			standardAttributes := map[string]any{}
			if actorEmail != "" {
				standardAttributes["email"] = actorEmail
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"node": map[string]any{
						"id":                 "user-id",
						"standardAttributes": standardAttributes,
					},
				},
			})
		}))
		defer server.Close()
		serverURL, err := url.Parse(server.URL)
		So(err, ShouldBeNil)

		newService := func(caseSensitive bool) *CollaboratorService {
			ignoreDotSign := false
			return &CollaboratorService{
				HTTPClient:     HTTPClient{Client: server.Client()},
				AdminAPI:       &fakeCollaboratorAdminAPI{serverURL: serverURL},
				AuthgearConfig: &portalconfig.AuthgearConfig{AppID: "accounts"},
				AppConfigs: &fakeCollaboratorAppConfigs{
					emailConfig: &config.LoginIDEmailConfig{
						CaseSensitive: &caseSensitive,
						IgnoreDotSign: &ignoreDotSign,
					},
				},
			}
		}
		invitation := &model.CollaboratorInvitation{InviteeEmail: "John.Doe@Example.com"}

		Convey("accepts email differing in case when case insensitive", func() {
			actorEmail = "john.doe@example.com"
			So(newService(false).CheckInviteeEmail(ctx, invitation, "user-id"), ShouldBeNil)
		})

		Convey("rejects email differing in case when case sensitive", func() {
			actorEmail = "john.doe@example.com"
			So(newService(true).CheckInviteeEmail(ctx, invitation, "user-id"), ShouldBeError, ErrCollaboratorInvitationInvalidEmail)
		})

		Convey("accepts domain differing in case when case sensitive", func() {
			actorEmail = "John.Doe@example.com"
			So(newService(true).CheckInviteeEmail(ctx, invitation, "user-id"), ShouldBeNil)
		})

		Convey("rejects different email", func() {
			actorEmail = "jane@example.com"
			So(newService(false).CheckInviteeEmail(ctx, invitation, "user-id"), ShouldBeError, ErrCollaboratorInvitationInvalidEmail)
		})

		Convey("rejects actor without email", func() {
			actorEmail = ""
			So(newService(false).CheckInviteeEmail(ctx, invitation, "user-id"), ShouldBeError, ErrCollaboratorInvitationInvalidEmail)
		})
	})
}
