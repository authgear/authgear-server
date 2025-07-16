package oauth

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/access"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

func newBool(b bool) *bool {
	return &b
}

// staticClientResolver implements OAuthClientResolver for testing
type staticClientResolver struct {
	Config *config.OAuthClientConfig
}

func (r *staticClientResolver) ResolveClient(clientID string) *config.OAuthClientConfig {
	if clientID == "testclient" {
		return r.Config
	}
	return nil
}

func TestOfflineGrantService(t *testing.T) {

	Convey("CreateNewRefreshToken", t, func() {
		Convey("Expired token will be removed", func() {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockIDPSessionProvider := NewMockServiceIDPSessionProvider(ctrl)
			mockAccessEventProvider := NewMockOfflineGrantServiceAccessEventProvider(ctrl)
			mockMeterService := NewMockOfflineGrantServiceMeterService(ctrl)
			mockOfflineGrantStore := NewMockOfflineGrantStore(ctrl)
			mockClock := clock.NewMockClockAt("2020-01-01T00:00:00Z")

			// Stub OAuthClientResolver
			testClientCfg := &config.OAuthClientConfig{
				ClientID:                       "testclient",
				RefreshTokenLifetime:           3600,
				RefreshTokenIdleTimeoutEnabled: newBool(true),
				RefreshTokenIdleTimeout:        300,
			}
			testResolver := &staticClientResolver{Config: testClientCfg}

			svc := &OfflineGrantService{
				IDPSessions:    mockIDPSessionProvider,
				AccessEvents:   mockAccessEventProvider,
				MeterService:   mockMeterService,
				OfflineGrants:  mockOfflineGrantStore,
				Clock:          mockClock,
				ClientResolver: testResolver,
			}

			ctx := context.Background()
			twoHoursAgo := mockClock.NowUTC().Add(-2 * 3600 * time.Second) // 2 hours ago, expired
			now := mockClock.NowUTC()                                      // valid

			rootToken := OfflineGrantRefreshToken{
				TokenHash: "root",
				ClientID:  "testclient",
				CreatedAt: twoHoursAgo,
				AccessInfo: &access.Info{
					InitialAccess: access.Event{Timestamp: twoHoursAgo},
					LastAccess:    access.Event{Timestamp: twoHoursAgo},
				},
			}
			expiredToken := OfflineGrantRefreshToken{
				TokenHash: "expired",
				ClientID:  "testclient",
				CreatedAt: twoHoursAgo,
				AccessInfo: &access.Info{
					InitialAccess: access.Event{Timestamp: twoHoursAgo},
					LastAccess:    access.Event{Timestamp: twoHoursAgo},
				},
			}
			validToken := OfflineGrantRefreshToken{
				TokenHash: "valid",
				ClientID:  "testclient",
				CreatedAt: now,
				AccessInfo: &access.Info{
					InitialAccess: access.Event{Timestamp: now},
					LastAccess:    access.Event{Timestamp: now},
				},
			}
			grant := &OfflineGrant{
				ID:              "grant-id",
				InitialClientID: "testclient",
				CreatedAt:       now,
				AccessInfo:      access.Info{InitialAccess: access.Event{Timestamp: now}, LastAccess: access.Event{Timestamp: now}},
				RefreshTokens:   []OfflineGrantRefreshToken{rootToken, expiredToken, validToken},
			}

			// AddOfflineGrantRefreshToken should add a new token
			mockOfflineGrantStore.EXPECT().
				AddOfflineGrantRefreshToken(gomock.Any(), "grant-id", gomock.Any(), gomock.Any(), gomock.Any(), "testclient", gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(ctx context.Context, grantID string, accessInfo access.Info, expireAt time.Time, tokenHash, clientID string, scopes []string, authorizationID, dpopJKT string) (*OfflineGrant, error) {
					newToken := OfflineGrantRefreshToken{
						TokenHash:  "newtoken",
						ClientID:   "testclient",
						CreatedAt:  mockClock.NowUTC(),
						Scopes:     []string{"openid"},
						AccessInfo: &accessInfo,
					}
					grant.RefreshTokens = append(grant.RefreshTokens, newToken)
					return grant, nil
				})

			// RemoveOfflineGrantRefreshTokens should be called with the expired token's hash
			// Only expired should be removed
			// Root token are always kept
			// Valid token are not removed
			expectedHashes := []string{"expired"}
			mockOfflineGrantStore.EXPECT().
				RemoveOfflineGrantRefreshTokens(gomock.Any(), "grant-id", gomock.Eq(expectedHashes), gomock.Any()).
				Return(grant, nil)

			_, _, err := svc.CreateNewRefreshToken(ctx, grant, "testclient", []string{"openid"}, "authz-id", "")

			So(err, ShouldBeNil)

		})

		Convey("Expired token with nil AccessInfo will be removed (backward compatibility)", func() {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockIDPSessionProvider := NewMockServiceIDPSessionProvider(ctrl)
			mockAccessEventProvider := NewMockOfflineGrantServiceAccessEventProvider(ctrl)
			mockMeterService := NewMockOfflineGrantServiceMeterService(ctrl)
			mockOfflineGrantStore := NewMockOfflineGrantStore(ctrl)
			mockClock := clock.NewMockClockAt("2020-01-01T00:00:00Z")

			testClientCfg := &config.OAuthClientConfig{
				ClientID:                       "testclient",
				RefreshTokenLifetime:           3600,
				RefreshTokenIdleTimeoutEnabled: newBool(true),
				RefreshTokenIdleTimeout:        300,
			}
			testResolver := &staticClientResolver{Config: testClientCfg}

			svc := &OfflineGrantService{
				IDPSessions:    mockIDPSessionProvider,
				AccessEvents:   mockAccessEventProvider,
				MeterService:   mockMeterService,
				OfflineGrants:  mockOfflineGrantStore,
				Clock:          mockClock,
				ClientResolver: testResolver,
			}

			ctx := context.Background()
			twoHoursAgo := mockClock.NowUTC().Add(-2 * 3600 * time.Second)
			now := mockClock.NowUTC()

			rootToken := OfflineGrantRefreshToken{
				TokenHash:  "root",
				ClientID:   "testclient",
				CreatedAt:  twoHoursAgo,
				AccessInfo: nil,
			}
			expiredTokenNilAccessInfo := OfflineGrantRefreshToken{
				TokenHash:  "expired-nil",
				ClientID:   "testclient",
				CreatedAt:  twoHoursAgo,
				AccessInfo: nil, // Simulate legacy token
			}
			validToken := OfflineGrantRefreshToken{
				TokenHash: "valid",
				ClientID:  "testclient",
				CreatedAt: now,
				AccessInfo: &access.Info{
					InitialAccess: access.Event{Timestamp: now},
					LastAccess:    access.Event{Timestamp: now},
				},
			}
			grant := &OfflineGrant{
				ID:              "grant-id",
				InitialClientID: "testclient",
				CreatedAt:       now,
				AccessInfo:      access.Info{InitialAccess: access.Event{Timestamp: now}, LastAccess: access.Event{Timestamp: now}},
				RefreshTokens:   []OfflineGrantRefreshToken{rootToken, expiredTokenNilAccessInfo, validToken},
			}

			mockOfflineGrantStore.EXPECT().
				AddOfflineGrantRefreshToken(gomock.Any(), "grant-id", gomock.Any(), gomock.Any(), gomock.Any(), "testclient", gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(ctx context.Context, grantID string, accessInfo access.Info, expireAt time.Time, tokenHash, clientID string, scopes []string, authorizationID, dpopJKT string) (*OfflineGrant, error) {
					newToken := OfflineGrantRefreshToken{
						TokenHash:  "newtoken",
						ClientID:   "testclient",
						CreatedAt:  mockClock.NowUTC(),
						Scopes:     []string{"openid"},
						AccessInfo: &accessInfo,
					}
					grant.RefreshTokens = append(grant.RefreshTokens, newToken)
					return grant, nil
				})

			expectedHashes := []string{"expired-nil"}
			mockOfflineGrantStore.EXPECT().
				RemoveOfflineGrantRefreshTokens(gomock.Any(), "grant-id", gomock.Eq(expectedHashes), gomock.Any()).
				Return(grant, nil)

			_, _, err := svc.CreateNewRefreshToken(ctx, grant, "testclient", []string{"openid"}, "authz-id", "")

			So(err, ShouldBeNil)
		})

		Convey("Token with nil AccessInfo is not removed if not expired (backward compatibility)", func() {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockIDPSessionProvider := NewMockServiceIDPSessionProvider(ctrl)
			mockAccessEventProvider := NewMockOfflineGrantServiceAccessEventProvider(ctrl)
			mockMeterService := NewMockOfflineGrantServiceMeterService(ctrl)
			mockOfflineGrantStore := NewMockOfflineGrantStore(ctrl)
			mockClock := clock.NewMockClockAt("2020-01-01T00:00:00Z")

			testClientCfg := &config.OAuthClientConfig{
				ClientID:                       "testclient",
				RefreshTokenLifetime:           3600,
				RefreshTokenIdleTimeoutEnabled: newBool(true),
				RefreshTokenIdleTimeout:        601, // 10 minutes + 1 sec
			}
			testResolver := &staticClientResolver{Config: testClientCfg}

			svc := &OfflineGrantService{
				IDPSessions:    mockIDPSessionProvider,
				AccessEvents:   mockAccessEventProvider,
				MeterService:   mockMeterService,
				OfflineGrants:  mockOfflineGrantStore,
				Clock:          mockClock,
				ClientResolver: testResolver,
			}

			ctx := context.Background()
			oneDayAgo := mockClock.NowUTC().Add(-1 * time.Hour * 24)
			tenMinsAgo := mockClock.NowUTC().Add(-10 * time.Minute)
			now := mockClock.NowUTC()

			rootToken := OfflineGrantRefreshToken{
				TokenHash:  "root",
				ClientID:   "testclient",
				CreatedAt:  tenMinsAgo,
				AccessInfo: nil,
			}
			// This token should not be removed because it was created 10 minutes ago,
			// and the idle timeout is 10minutes + 1 seconds.
			// CreatedAt should be used as the last access time so it is not expired yet.
			validToken := OfflineGrantRefreshToken{
				TokenHash:  "valid",
				ClientID:   "testclient",
				CreatedAt:  tenMinsAgo,
				AccessInfo: nil,
			}
			grant := &OfflineGrant{
				ID:              "grant-id",
				InitialClientID: "testclient",
				CreatedAt:       now,
				AccessInfo:      access.Info{InitialAccess: access.Event{Timestamp: oneDayAgo}, LastAccess: access.Event{Timestamp: tenMinsAgo}},
				RefreshTokens:   []OfflineGrantRefreshToken{rootToken, validToken},
			}

			mockOfflineGrantStore.EXPECT().
				AddOfflineGrantRefreshToken(gomock.Any(), "grant-id", gomock.Any(), gomock.Any(), gomock.Any(), "testclient", gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(ctx context.Context, grantID string, accessInfo access.Info, expireAt time.Time, tokenHash, clientID string, scopes []string, authorizationID, dpopJKT string) (*OfflineGrant, error) {
					newToken := OfflineGrantRefreshToken{
						TokenHash:  "newtoken",
						ClientID:   "testclient",
						CreatedAt:  mockClock.NowUTC(),
						Scopes:     []string{"openid"},
						AccessInfo: &accessInfo,
					}
					grant.RefreshTokens = append(grant.RefreshTokens, newToken)
					return grant, nil
				})

			// No tokens should be removed
			mockOfflineGrantStore.EXPECT().
				RemoveOfflineGrantRefreshTokens(gomock.Any(), "grant-id", gomock.Eq([]string{}), gomock.Any()).
				Return(grant, nil)

			_, _, err := svc.CreateNewRefreshToken(ctx, grant, "testclient", []string{"openid"}, "authz-id", "")

			So(err, ShouldBeNil)
		})
	})

	Convey("AccessOfflineGrant", t, func() {
		createMockedUpdateOfflineGrantWithMutator := func(grant *OfflineGrant) func(ctx context.Context, grantID string, expireAt time.Time, mutator func(*OfflineGrant) *OfflineGrant) (*OfflineGrant, error) {
			return func(ctx context.Context, grantID string, expireAt time.Time, mutator func(*OfflineGrant) *OfflineGrant) (*OfflineGrant, error) {
				// run mutator on a copy
				grantCopy := *grant
				grantCopy.RefreshTokens = make([]OfflineGrantRefreshToken, len(grant.RefreshTokens))
				copy(grantCopy.RefreshTokens, grant.RefreshTokens)
				result := mutator(&grantCopy)
				return result, nil
			}
		}
		Convey("updates last access and calls side effects", func() {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockIDPSessionProvider := NewMockServiceIDPSessionProvider(ctrl)
			mockAccessEventProvider := NewMockOfflineGrantServiceAccessEventProvider(ctrl)
			mockMeterService := NewMockOfflineGrantServiceMeterService(ctrl)
			mockOfflineGrantStore := NewMockOfflineGrantStore(ctrl)
			mockClock := clock.NewMockClockAt("2020-01-01T00:00:00Z")

			svc := &OfflineGrantService{
				IDPSessions:   mockIDPSessionProvider,
				AccessEvents:  mockAccessEventProvider,
				MeterService:  mockMeterService,
				OfflineGrants: mockOfflineGrantStore,
				Clock:         mockClock,
			}

			ctx := context.Background()
			now := mockClock.NowUTC()
			userID := "user-id"
			previousAccessEvent := access.NewEvent(now.Add(-1*time.Hour), "1.2.3.4", "UA")
			accessEvent := access.NewEvent(now, "1.2.3.4", "UA")
			tokenHash := "token-hash"

			grant := &OfflineGrant{
				ID:              "grant-id",
				InitialClientID: "testclient",
				CreatedAt:       now,
				Attrs:           session.Attrs{UserID: userID},
				AccessInfo:      access.Info{InitialAccess: previousAccessEvent, LastAccess: previousAccessEvent},
				RefreshTokens: []OfflineGrantRefreshToken{
					{
						TokenHash:  tokenHash,
						ClientID:   "testclient",
						CreatedAt:  now,
						AccessInfo: &access.Info{InitialAccess: previousAccessEvent, LastAccess: previousAccessEvent},
					},
					{
						TokenHash:  "another-token-hash",
						ClientID:   "testclient",
						CreatedAt:  now,
						AccessInfo: &access.Info{InitialAccess: previousAccessEvent, LastAccess: previousAccessEvent},
					},
				},
				ExpireAtForResolvedSession: now.Add(1 * time.Hour),
			}

			// Mock UpdateOfflineGrantWithMutator to run the mutator and return the result
			mockOfflineGrantStore.EXPECT().
				UpdateOfflineGrantWithMutator(gomock.Any(), "grant-id", grant.ExpireAtForResolvedSession, gomock.Any()).
				DoAndReturn(createMockedUpdateOfflineGrantWithMutator(grant))

			mockAccessEventProvider.EXPECT().
				RecordAccess(gomock.Any(), "grant-id", grant.ExpireAtForResolvedSession, &accessEvent).
				Return(nil)

			mockMeterService.EXPECT().
				TrackActiveUser(gomock.Any(), userID).
				Return(nil)

			updatedGrant, err := svc.AccessOfflineGrant(ctx, "grant-id", tokenHash, &accessEvent, grant.ExpireAtForResolvedSession)

			So(err, ShouldBeNil)
			// AccessInfo of the OfflineGrant should be updated
			So(updatedGrant.AccessInfo.LastAccess, ShouldResemble, accessEvent)
			// AccessInfo of the used refresh token should be updated
			So(updatedGrant.RefreshTokens[0].AccessInfo.LastAccess, ShouldResemble, accessEvent)
			// AccessInfo of the unused refresh token should not be updated
			So(updatedGrant.RefreshTokens[1].AccessInfo.LastAccess, ShouldResemble, previousAccessEvent)
		})

		Convey("updates legacy token with nil AccessInfo (backward compatibility)", func() {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockIDPSessionProvider := NewMockServiceIDPSessionProvider(ctrl)
			mockAccessEventProvider := NewMockOfflineGrantServiceAccessEventProvider(ctrl)
			mockMeterService := NewMockOfflineGrantServiceMeterService(ctrl)
			mockOfflineGrantStore := NewMockOfflineGrantStore(ctrl)
			mockClock := clock.NewMockClockAt("2020-01-01T00:00:00Z")

			svc := &OfflineGrantService{
				IDPSessions:   mockIDPSessionProvider,
				AccessEvents:  mockAccessEventProvider,
				MeterService:  mockMeterService,
				OfflineGrants: mockOfflineGrantStore,
				Clock:         mockClock,
			}

			ctx := context.Background()
			now := mockClock.NowUTC()
			userID := "user-id"
			previousAccessEvent := access.NewEvent(now.Add(-1*time.Hour), "1.2.3.4", "UA")
			accessEvent := access.NewEvent(now, "1.2.3.4", "UA")
			legacyTokenHash := "legacy-token-hash"

			grant := &OfflineGrant{
				ID:              "grant-id",
				InitialClientID: "testclient",
				CreatedAt:       now,
				Attrs:           session.Attrs{UserID: userID},
				AccessInfo:      access.Info{InitialAccess: previousAccessEvent, LastAccess: previousAccessEvent},
				RefreshTokens: []OfflineGrantRefreshToken{
					{
						TokenHash:  legacyTokenHash,
						ClientID:   "testclient",
						CreatedAt:  now,
						AccessInfo: nil, // legacy/old token
					},
					{
						TokenHash:  "other-token-hash",
						ClientID:   "testclient",
						CreatedAt:  now,
						AccessInfo: &access.Info{InitialAccess: previousAccessEvent, LastAccess: previousAccessEvent},
					},
				},
				ExpireAtForResolvedSession: now.Add(1 * time.Hour),
			}

			mockOfflineGrantStore.EXPECT().
				UpdateOfflineGrantWithMutator(gomock.Any(), "grant-id", grant.ExpireAtForResolvedSession, gomock.Any()).
				DoAndReturn(createMockedUpdateOfflineGrantWithMutator(grant))

			mockAccessEventProvider.EXPECT().
				RecordAccess(gomock.Any(), "grant-id", grant.ExpireAtForResolvedSession, &accessEvent).
				Return(nil)

			mockMeterService.EXPECT().
				TrackActiveUser(gomock.Any(), userID).
				Return(nil)

			updatedGrant, err := svc.AccessOfflineGrant(ctx, "grant-id", legacyTokenHash, &accessEvent, grant.ExpireAtForResolvedSession)

			So(err, ShouldBeNil)
			// The legacy token's AccessInfo should now be non-nil and LastAccess should be updated
			So(updatedGrant.RefreshTokens[0].AccessInfo, ShouldNotBeNil)
			So(updatedGrant.RefreshTokens[0].AccessInfo.LastAccess, ShouldResemble, accessEvent)
			// The other token should remain unchanged
			So(updatedGrant.RefreshTokens[1].AccessInfo.LastAccess, ShouldResemble, previousAccessEvent)
		})
	})
}
