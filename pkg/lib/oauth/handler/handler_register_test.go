package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	gomock "github.com/golang/mock/gomock"
	_ "github.com/mattn/go-sqlite3"
	. "github.com/smartystreets/goconvey/convey"

	apievent "github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/dcr"
	dbinfra "github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/oauth/handler"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/lib/usage"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/otelutil/oteldatabasesql"
)

// registerTestPool and newTestAppDBHandle back RegistrationHandler.Database
// (a concrete *appdb.Handle, not an interface) with a real sqlite
// connection, exactly like pkg/lib/usage/limit_test.go's newTestAppDBHandle
// -- dispatchImmediately's IsInTx/WithTx/ReadOnly branch needs a working
// transaction, not a mock, and HookHandle issues only plain database/sql
// calls, so sqlite is a faithful stand-in.
type registerTestPool struct {
	connPool oteldatabasesql.ConnPool_
}

func (p *registerTestPool) Open(info dbinfra.ConnectionInfo, opts dbinfra.ConnectionOptions) (oteldatabasesql.ConnPool_, error) {
	return p.connPool, nil
}

func (p *registerTestPool) Close() error { return nil }

func newRegisterTestAppDBHandle(t *testing.T) *appdb.Handle {
	t.Helper()

	connPool, err := oteldatabasesql.Open(oteldatabasesql.OpenOptions{
		DriverName: "sqlite3",
		DSN:        ":memory:",
	})
	So(err, ShouldBeNil)
	connPool.SetMaxOpenConns(1)
	connPool.SetMaxIdleConns(0)
	connPool.SetConnMaxLifetime(0)
	connPool.SetConnMaxIdleTime(0)

	return &appdb.Handle{
		HookHandle: dbinfra.NewHookHandle(&registerTestPool{connPool: connPool}, dbinfra.ConnectionInfo{}, dbinfra.ConnectionOptions{}),
	}
}

func TestRegistrationHandler(t *testing.T) {
	Convey("RegistrationHandler", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		dcrService := NewMockRegistrationHandlerDCRService(ctrl)
		iatService := NewMockRegistrationHandlerIATService(ctrl)
		rateLimiter := NewMockRegistrationHandlerRateLimiter(ctrl)
		rateLimiter.EXPECT().Allow(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
		usageLimiter := NewMockRegistrationHandlerUsageLimiter(ctrl)
		events := NewMockEventService(ctrl)

		appDB := newRegisterTestAppDBHandle(t)
		fixedTime := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
		mockClock := clock.NewMockClockAtTime(fixedTime)

		oauthConfig := &config.OAuthConfig{
			DynamicClientRegistration: &config.OAuthDynamicClientRegistrationConfig{
				Enabled: true,
			},
		}
		config.SetFieldDefaults(oauthConfig.DynamicClientRegistration)

		h := &handler.RegistrationHandler{
			Database:     appDB,
			OAuthConfig:  oauthConfig,
			DCR:          dcrService,
			IAT:          iatService,
			Clock:        mockClock,
			Events:       events,
			RemoteIP:     "127.0.0.1",
			RateLimiter:  rateLimiter,
			UsageLimiter: usageLimiter,
		}

		const validBody = `{"redirect_uris":["https://app.example.com/cb"]}`

		newRequest := func(body string, bearer string) *http.Request {
			req := httptest.NewRequest(http.MethodPost, "/oauth2/register", strings.NewReader(body))
			if bearer != "" {
				req.Header.Set("Authorization", bearer)
			}
			return req
		}

		expectNoDispatch := func() {
			events.EXPECT().DispatchEventImmediately(gomock.Any(), gomock.Any()).Times(0)
		}

		fixtureIAT := &model.OAuthInitialAccessToken{
			Meta: model.Meta{
				ID:        "iat-id",
				CreatedAt: fixedTime.Add(-time.Hour),
			},
			Type:      model.OAuthInitialAccessTokenTypeThirdParty,
			ExpiresAt: fixedTime.Add(-time.Minute), // already expired
		}

		Convey("DCR not enabled: no event, no rate limit check", func() {
			oauthConfig.DynamicClientRegistration.Enabled = false
			expectNoDispatch()
			_, err := h.Handle(context.Background(), newRequest(validBody, ""))
			protoErr, ok := err.(*protocol.OAuthProtocolError)
			So(ok, ShouldBeTrue)
			So(protoErr.StatusCode, ShouldEqual, http.StatusForbidden)
		})

		Convey("Authorization header present but not Bearer", func() {
			var captured *nonblocking.OAuthClientRegistrationFailedEventPayload
			events.EXPECT().DispatchEventImmediately(gomock.Any(), gomock.Any()).DoAndReturn(
				func(ctx context.Context, payload apievent.NonBlockingPayload) error {
					captured = payload.(*nonblocking.OAuthClientRegistrationFailedEventPayload)
					return nil
				},
			)
			resp, err := h.Handle(context.Background(), newRequest(validBody, "Basic abc"))
			So(resp, ShouldBeNil)
			protoErr, ok := err.(*protocol.OAuthProtocolError)
			So(ok, ShouldBeTrue)
			So(protoErr.StatusCode, ShouldEqual, http.StatusUnauthorized)
			So(protoErr.Response["error"], ShouldEqual, "invalid_initial_access_token")

			So(captured.Reason, ShouldEqual, nonblocking.OAuthClientRegistrationReasonInvalidInitialAccessToken)
			So(captured.Message, ShouldEqual, "malformed_header")
			So(captured.InitialAccessToken, ShouldBeNil)
		})

		Convey("token required and absent", func() {
			b := true
			oauthConfig.DynamicClientRegistration.InitialAccessTokenRequired = &b
			var captured *nonblocking.OAuthClientRegistrationFailedEventPayload
			events.EXPECT().DispatchEventImmediately(gomock.Any(), gomock.Any()).DoAndReturn(
				func(ctx context.Context, payload apievent.NonBlockingPayload) error {
					captured = payload.(*nonblocking.OAuthClientRegistrationFailedEventPayload)
					return nil
				},
			)
			resp, err := h.Handle(context.Background(), newRequest(validBody, ""))
			So(resp, ShouldBeNil)
			protoErr, ok := err.(*protocol.OAuthProtocolError)
			So(ok, ShouldBeTrue)
			So(protoErr.StatusCode, ShouldEqual, http.StatusUnauthorized)
			So(protoErr.Response["error"], ShouldEqual, "invalid_initial_access_token")

			So(captured.Message, ShouldEqual, "not_presented")
			So(captured.InitialAccessToken, ShouldBeNil)
		})

		Convey("malformed body", func() {
			b := false
			oauthConfig.DynamicClientRegistration.InitialAccessTokenRequired = &b
			var captured *nonblocking.OAuthClientRegistrationFailedEventPayload
			events.EXPECT().DispatchEventImmediately(gomock.Any(), gomock.Any()).DoAndReturn(
				func(ctx context.Context, payload apievent.NonBlockingPayload) error {
					captured = payload.(*nonblocking.OAuthClientRegistrationFailedEventPayload)
					return nil
				},
			)
			resp, err := h.Handle(context.Background(), newRequest("{not json", ""))
			So(resp, ShouldBeNil)
			protoErr, ok := err.(*protocol.OAuthProtocolError)
			So(ok, ShouldBeTrue)
			So(protoErr.Response["error"], ShouldEqual, "invalid_client_metadata")

			So(captured.Reason, ShouldEqual, nonblocking.OAuthClientRegistrationReasonInvalidClientMetadata)
			So(captured.Message, ShouldEqual, "malformed_json")
		})

		Convey("every dcr.ErrDCR* validation sentinel gets a non-empty message", func() {
			b := false
			oauthConfig.DynamicClientRegistration.InitialAccessTokenRequired = &b
			cases := []struct {
				name    string
				body    string
				message string
			}{
				{
					"redirect_uris missing",
					`{}`,
					"redirect_uris_missing",
				},
				{
					"redirect_uri invalid (http for web)",
					`{"redirect_uris":["http://app.example.com/cb"]}`,
					"redirect_uri_invalid",
				},
				{
					"token_endpoint_auth_method not accepted",
					`{"redirect_uris":["https://app.example.com/cb"],"token_endpoint_auth_method":"client_secret_post"}`,
					"token_endpoint_auth_method_not_accepted",
				},
				{
					"grant_type unsupported",
					`{"redirect_uris":["https://app.example.com/cb"],"grant_types":["client_credentials"]}`,
					"grant_type_unsupported",
				},
				{
					"response_type inconsistent",
					`{"redirect_uris":["https://app.example.com/cb"],"response_types":["token"]}`,
					"response_type_inconsistent",
				},
				{
					"application_type unsupported",
					`{"redirect_uris":["https://app.example.com/cb"],"application_type":"spa"}`,
					"application_type_unsupported",
				},
				{
					"uri field not https",
					`{"redirect_uris":["https://app.example.com/cb"],"logo_uri":"http://app.example.com/logo.png"}`,
					"uri_field_not_https",
				},
			}
			for _, tc := range cases {
				Convey(tc.name, func() {
					var captured *nonblocking.OAuthClientRegistrationFailedEventPayload
					events.EXPECT().DispatchEventImmediately(gomock.Any(), gomock.Any()).DoAndReturn(
						func(ctx context.Context, payload apievent.NonBlockingPayload) error {
							captured = payload.(*nonblocking.OAuthClientRegistrationFailedEventPayload)
							return nil
						},
					)
					resp, err := h.Handle(context.Background(), newRequest(tc.body, ""))
					So(resp, ShouldBeNil)
					So(err, ShouldNotBeNil)
					So(captured.Reason, ShouldEqual, nonblocking.OAuthClientRegistrationReasonInvalidClientMetadata)
					So(captured.Message, ShouldEqual, tc.message)
					So(captured.Message, ShouldNotBeEmpty)
				})
			}
		})

		Convey("unknown initial access token", func() {
			iatService.EXPECT().ValidateAndGetByToken(gomock.Any(), "sometoken").Return(nil, dcr.ErrInitialAccessTokenNotFound)
			var captured *nonblocking.OAuthClientRegistrationFailedEventPayload
			events.EXPECT().DispatchEventImmediately(gomock.Any(), gomock.Any()).DoAndReturn(
				func(ctx context.Context, payload apievent.NonBlockingPayload) error {
					captured = payload.(*nonblocking.OAuthClientRegistrationFailedEventPayload)
					return nil
				},
			)
			resp, err := h.Handle(context.Background(), newRequest(validBody, "Bearer sometoken"))
			So(resp, ShouldBeNil)
			protoErr, ok := err.(*protocol.OAuthProtocolError)
			So(ok, ShouldBeTrue)
			So(protoErr.StatusCode, ShouldEqual, http.StatusUnauthorized)
			So(protoErr.Response["error"], ShouldEqual, "invalid_initial_access_token")

			So(captured.Message, ShouldEqual, "unknown")
			So(captured.InitialAccessToken, ShouldBeNil)
		})

		Convey("expired initial access token: audit record carries the row, HTTP response does not change", func() {
			iatService.EXPECT().ValidateAndGetByToken(gomock.Any(), "expiredtoken").Return(fixtureIAT, dcr.ErrInitialAccessTokenExpired)
			var captured *nonblocking.OAuthClientRegistrationFailedEventPayload
			events.EXPECT().DispatchEventImmediately(gomock.Any(), gomock.Any()).DoAndReturn(
				func(ctx context.Context, payload apievent.NonBlockingPayload) error {
					captured = payload.(*nonblocking.OAuthClientRegistrationFailedEventPayload)
					return nil
				},
			)
			resp, err := h.Handle(context.Background(), newRequest(validBody, "Bearer expiredtoken"))
			So(resp, ShouldBeNil)
			protoErr, ok := err.(*protocol.OAuthProtocolError)
			So(ok, ShouldBeTrue)
			So(protoErr.StatusCode, ShouldEqual, http.StatusUnauthorized)
			So(protoErr.Response["error"], ShouldEqual, "invalid_initial_access_token")

			So(captured.Message, ShouldEqual, "expired")
			So(captured.InitialAccessToken, ShouldNotBeNil)
			So(captured.InitialAccessToken.ID, ShouldEqual, "iat-id")
			So(captured.InitialAccessToken.Type, ShouldEqual, model.OAuthInitialAccessTokenTypeThirdParty)
			So(captured.InitialAccessToken.CreatedAt, ShouldEqual, fixtureIAT.CreatedAt)
			So(captured.InitialAccessToken.ExpiresAt, ShouldEqual, fixtureIAT.ExpiresAt)
		})

		Convey("unknown vs expired: identical status code and error code, only message differs", func() {
			iatService.EXPECT().ValidateAndGetByToken(gomock.Any(), "unknowntoken").Return(nil, dcr.ErrInitialAccessTokenNotFound)
			iatService.EXPECT().ValidateAndGetByToken(gomock.Any(), "expiredtoken").Return(fixtureIAT, dcr.ErrInitialAccessTokenExpired)
			events.EXPECT().DispatchEventImmediately(gomock.Any(), gomock.Any()).Return(nil).Times(2)

			_, unknownErr := h.Handle(context.Background(), newRequest(validBody, "Bearer unknowntoken"))
			_, expiredErr := h.Handle(context.Background(), newRequest(validBody, "Bearer expiredtoken"))

			unknownProtoErr := unknownErr.(*protocol.OAuthProtocolError)
			expiredProtoErr := expiredErr.(*protocol.OAuthProtocolError)
			So(unknownProtoErr.StatusCode, ShouldEqual, expiredProtoErr.StatusCode)
			So(unknownProtoErr.Response["error"], ShouldEqual, expiredProtoErr.Response["error"])
			So(unknownProtoErr.Response["error_description"], ShouldEqual, expiredProtoErr.Response["error_description"])
		})

		Convey("at quota: limit_exceeded, usage_name and quota carried, no registered event", func() {
			b := false
			oauthConfig.DynamicClientRegistration.InitialAccessTokenRequired = &b
			dcrService.EXPECT().LockForClientCount(gomock.Any(), model.OAuthClientSourceDCR).Return(nil)
			dcrService.EXPECT().CountClientsBySource(gomock.Any(), model.OAuthClientSourceDCR).Return(uint64(5), nil)
			usageLimiter.EXPECT().CheckStanding(gomock.Any(), model.UsageNameOAuthClientDCR, 5).
				Return(usage.ErrStandingUsageLimitExceeded(model.UsageNameOAuthClientDCR, 5))

			var captured *nonblocking.OAuthClientRegistrationFailedEventPayload
			events.EXPECT().DispatchEventImmediately(gomock.Any(), gomock.Any()).DoAndReturn(
				func(ctx context.Context, payload apievent.NonBlockingPayload) error {
					captured = payload.(*nonblocking.OAuthClientRegistrationFailedEventPayload)
					return nil
				},
			)

			resp, err := h.Handle(context.Background(), newRequest(validBody, ""))
			So(resp, ShouldBeNil)
			protoErr, ok := err.(*protocol.OAuthProtocolError)
			So(ok, ShouldBeTrue)
			So(protoErr.StatusCode, ShouldEqual, http.StatusForbidden)

			So(captured.Reason, ShouldEqual, nonblocking.OAuthClientRegistrationReasonLimitExceeded)
			So(captured.UsageName, ShouldEqual, model.UsageNameOAuthClientDCR)
			So(captured.Quota, ShouldEqual, 5)
			So(captured.Message, ShouldBeEmpty)
		})

		Convey("successful registration with an IAT: registered event's initial_access_token matches the failure event's shape", func() {
			token := &model.OAuthInitialAccessToken{
				Meta: model.Meta{
					ID:        "iat-live",
					CreatedAt: fixedTime.Add(-time.Hour),
				},
				Type:      model.OAuthInitialAccessTokenTypeThirdParty,
				ExpiresAt: fixedTime.Add(time.Hour),
			}
			iatService.EXPECT().ValidateAndGetByToken(gomock.Any(), "livetoken").Return(token, nil)
			dcrService.EXPECT().LockForClientCount(gomock.Any(), model.OAuthClientSourceDCR).Return(nil)
			dcrService.EXPECT().CountClientsBySource(gomock.Any(), model.OAuthClientSourceDCR).Return(uint64(0), nil)
			usageLimiter.EXPECT().CheckStanding(gomock.Any(), model.UsageNameOAuthClientDCR, 0).Return(nil)
			usageLimiter.EXPECT().ReportStandingCreated(gomock.Any(), model.UsageNameOAuthClientDCR, 0)

			appType := "web"
			client := &model.OAuthClient{
				Meta: model.Meta{
					ID:        "client-record-id",
					CreatedAt: fixedTime,
				},
				ClientID:        "dcrc_abc",
				Source:          model.OAuthClientSourceDCR,
				Kind:            model.OAuthClientKindThirdParty,
				ApplicationType: &appType,
				RedirectURIs:    []string{"https://app.example.com/cb"},
				GrantTypes:      []string{"authorization_code", "refresh_token"},
				ResponseTypes:   []string{"code"},
			}
			dcrService.EXPECT().RegisterClient(gomock.Any(), gomock.Any()).Return(client, nil)

			var registered *nonblocking.OAuthClientRegisteredEventPayload
			events.EXPECT().DispatchEventOnCommit(gomock.Any(), gomock.Any()).DoAndReturn(
				func(ctx context.Context, payload apievent.Payload) error {
					registered = payload.(*nonblocking.OAuthClientRegisteredEventPayload)
					return nil
				},
			)
			events.EXPECT().DispatchEventImmediately(gomock.Any(), gomock.Any()).Times(0)

			resp, err := h.Handle(context.Background(), newRequest(validBody, "Bearer livetoken"))
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.ClientID, ShouldEqual, "dcrc_abc")

			So(registered.InitialAccessToken, ShouldNotBeNil)
			So(registered.InitialAccessToken.ID, ShouldEqual, "iat-live")
			So(registered.InitialAccessToken.Type, ShouldEqual, model.OAuthInitialAccessTokenTypeThirdParty)
			So(registered.InitialAccessToken.CreatedAt, ShouldEqual, token.CreatedAt)
			So(registered.InitialAccessToken.ExpiresAt, ShouldEqual, token.ExpiresAt)

			// Same fixture token: the shape the registered event just carried
			// is byte-identical to what the failure event would have emitted
			// had this same token instead been expired -- proving both
			// events genuinely share nonblocking.EventPayloadInitialAccessToken
			// rather than two independently-built lookalikes.
			expiredVariant := &model.OAuthInitialAccessToken{Meta: token.Meta, Type: token.Type, ExpiresAt: token.ExpiresAt}
			fromFailure := nonblocking.NewEventPayloadInitialAccessToken(expiredVariant)
			So(registered.InitialAccessToken, ShouldResemble, fromFailure)
		})

		Convey("successful registration under open registration: initial_access_token key is absent, not zero-valued", func() {
			b := false
			oauthConfig.DynamicClientRegistration.InitialAccessTokenRequired = &b
			dcrService.EXPECT().LockForClientCount(gomock.Any(), model.OAuthClientSourceDCR).Return(nil)
			dcrService.EXPECT().CountClientsBySource(gomock.Any(), model.OAuthClientSourceDCR).Return(uint64(0), nil)
			usageLimiter.EXPECT().CheckStanding(gomock.Any(), model.UsageNameOAuthClientDCR, 0).Return(nil)
			usageLimiter.EXPECT().ReportStandingCreated(gomock.Any(), model.UsageNameOAuthClientDCR, 0)

			appType := "web"
			client := &model.OAuthClient{
				Meta:            model.Meta{ID: "client-record-id", CreatedAt: fixedTime},
				ClientID:        "dcrc_open",
				Source:          model.OAuthClientSourceDCR,
				Kind:            model.OAuthClientKindThirdParty,
				ApplicationType: &appType,
				RedirectURIs:    []string{"https://app.example.com/cb"},
				GrantTypes:      []string{"authorization_code", "refresh_token"},
				ResponseTypes:   []string{"code"},
			}
			dcrService.EXPECT().RegisterClient(gomock.Any(), gomock.Any()).Return(client, nil)

			var registered *nonblocking.OAuthClientRegisteredEventPayload
			events.EXPECT().DispatchEventOnCommit(gomock.Any(), gomock.Any()).DoAndReturn(
				func(ctx context.Context, payload apievent.Payload) error {
					registered = payload.(*nonblocking.OAuthClientRegisteredEventPayload)
					return nil
				},
			)
			expectNoDispatch()

			resp, err := h.Handle(context.Background(), newRequest(validBody, ""))
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(registered.InitialAccessToken, ShouldBeNil)
		})

		Convey("successful registration emits only oauth.client.registered, never a failure event", func() {
			dcrService.EXPECT().LockForClientCount(gomock.Any(), model.OAuthClientSourceDCR).Return(nil)
			dcrService.EXPECT().CountClientsBySource(gomock.Any(), model.OAuthClientSourceDCR).Return(uint64(0), nil)
			usageLimiter.EXPECT().CheckStanding(gomock.Any(), model.UsageNameOAuthClientDCR, 0).Return(nil)
			usageLimiter.EXPECT().ReportStandingCreated(gomock.Any(), model.UsageNameOAuthClientDCR, 0)
			b := false
			oauthConfig.DynamicClientRegistration.InitialAccessTokenRequired = &b

			appType := "web"
			client := &model.OAuthClient{
				Meta:            model.Meta{ID: "client-record-id", CreatedAt: fixedTime},
				ClientID:        "dcrc_solo",
				Source:          model.OAuthClientSourceDCR,
				Kind:            model.OAuthClientKindThirdParty,
				ApplicationType: &appType,
				RedirectURIs:    []string{"https://app.example.com/cb"},
				GrantTypes:      []string{"authorization_code", "refresh_token"},
				ResponseTypes:   []string{"code"},
			}
			dcrService.EXPECT().RegisterClient(gomock.Any(), gomock.Any()).Return(client, nil)
			events.EXPECT().DispatchEventOnCommit(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			expectNoDispatch()

			_, err := h.Handle(context.Background(), newRequest(validBody, ""))
			So(err, ShouldBeNil)
		})

		Convey("a dispatch error on the Immediately path does not change the HTTP response", func() {
			events.EXPECT().DispatchEventImmediately(gomock.Any(), gomock.Any()).Return(context.DeadlineExceeded)
			resp, err := h.Handle(context.Background(), newRequest(validBody, "Basic abc"))
			So(resp, ShouldBeNil)
			protoErr, ok := err.(*protocol.OAuthProtocolError)
			So(ok, ShouldBeTrue)
			So(protoErr.StatusCode, ShouldEqual, http.StatusUnauthorized)
			So(protoErr.Response["error"], ShouldEqual, "invalid_initial_access_token")
		})
	})
}
