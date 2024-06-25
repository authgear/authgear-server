package configsource

import (
	"context"
	"testing"

	gomock "github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	apimodel "github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	configtest "github.com/authgear/authgear-server/pkg/lib/config/test"
	_ "github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/facebook"
	_ "github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/google"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

func TestAuthgearYAML(t *testing.T) {
	Convey("AuthgearYAML custom attributes", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		domainService := NewMockDomainService(ctrl)
		domainService.EXPECT().ListDomains("test").Return([]*apimodel.Domain{}, nil).AnyTimes()

		path := "authgear.yaml"
		featureConfig := config.NewEffectiveDefaultFeatureConfig()
		ctx := context.Background()
		ctx = context.WithValue(ctx, ContextKeyFeatureConfig, featureConfig)
		ctx = context.WithValue(ctx, ContextKeyAppHostSuffixes, &config.AppHostSuffixes{})
		ctx = context.WithValue(ctx, ContextKeyDomainService, domainService)
		app := resource.LeveledAferoFs{FsLevel: resource.FsLevelApp}
		descriptor := &AuthgearYAMLDescriptor{}

		Convey("Custom attributes cannot be removed", func() {
			_, err := descriptor.UpdateResource(
				ctx,
				nil,
				&resource.ResourceFile{
					Location: resource.Location{
						Fs:   app,
						Path: path,
					},
					Data: []byte(`id: test
http:
  public_origin: http://test
user_profile:
  custom_attributes:
    attributes:
    - id: "0000"
      pointer: /a
      type: string
`),
				},
				[]byte(`id: test
http:
  public_origin: http://test
`),
			)
			So(err, ShouldBeError, `invalid authgear.yaml:
/user_profile/custom_attributes/attributes: custom attribute of id '0000' was deleted`)
		})

		Convey("Custom attribute ID cannot be changed", func() {
			_, err := descriptor.UpdateResource(
				ctx,
				nil,
				&resource.ResourceFile{
					Location: resource.Location{
						Fs:   app,
						Path: path,
					},
					Data: []byte(`id: test
http:
  public_origin: http://test
user_profile:
  custom_attributes:
    attributes:
    - id: "0000"
      pointer: /a
      type: string
`),
				},
				[]byte(`id: test
http:
  public_origin: http://test
user_profile:
  custom_attributes:
    attributes:
    - id: "0001"
      pointer: /a
      type: string
`),
			)
			So(err, ShouldBeError, `invalid authgear.yaml:
/user_profile/custom_attributes/attributes: custom attribute of id '0000' was deleted`)
		})

		Convey("Custom attribute type cannot be changed", func() {
			_, err := descriptor.UpdateResource(
				ctx,
				nil,
				&resource.ResourceFile{
					Location: resource.Location{
						Fs:   app,
						Path: path,
					},
					Data: []byte(`id: test
http:
  public_origin: http://test
user_profile:
  custom_attributes:
    attributes:
    - id: "0000"
      pointer: /a
      type: string
`),
				},
				[]byte(`id: test
http:
  public_origin: http://test
user_profile:
  custom_attributes:
    attributes:
    - id: "0000"
      pointer: /a
      type: integer
`),
			)
			So(err, ShouldBeError, `invalid authgear.yaml:
/user_profile/custom_attributes/attributes/0: custom attribute of id '0000' has type changed; original: string, incoming: integer`)
		})

		Convey("Custom attribute can be added", func() {
			_, err := descriptor.UpdateResource(
				ctx,
				nil,
				&resource.ResourceFile{
					Location: resource.Location{
						Fs:   app,
						Path: path,
					},
					Data: []byte(`id: test
http:
  public_origin: http://test
user_profile:
  custom_attributes:
    attributes:
    - id: "0000"
      pointer: /a
      type: string
`),
				},
				[]byte(`id: test
http:
  public_origin: http://test
user_profile:
  custom_attributes:
    attributes:
    - id: "0000"
      pointer: /a
      type: string
    - id: "0001"
      pointer: /b
      type: string
`),
			)
			So(err, ShouldBeNil)
		})

		Convey("Custom attribute can be reordered", func() {
			_, err := descriptor.UpdateResource(
				ctx,
				nil,
				&resource.ResourceFile{
					Location: resource.Location{
						Fs:   app,
						Path: path,
					},
					Data: []byte(`id: test
http:
  public_origin: http://test
user_profile:
  custom_attributes:
    attributes:
    - id: "0000"
      pointer: /a
      type: string
    - id: "0001"
      pointer: /b
      type: string
`),
				},
				[]byte(`id: test
http:
  public_origin: http://test
user_profile:
  custom_attributes:
    attributes:
    - id: "0001"
      pointer: /b
      type: string
    - id: "0000"
      pointer: /a
      type: string
`),
			)
			So(err, ShouldBeNil)
		})
	})

	Convey("AuthgearYAML public origin", t, func() {
		path := "authgear.yaml"
		app := resource.LeveledAferoFs{FsLevel: resource.FsLevelApp}
		descriptor := &AuthgearYAMLDescriptor{}

		Convey("Public origin can be changed to builtin appHostSuffix", func() {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			domainService := NewMockDomainService(ctrl)
			domainService.EXPECT().ListDomains("test").Return([]*apimodel.Domain{}, nil).AnyTimes()

			featureConfig := config.NewEffectiveDefaultFeatureConfig()

			ctx := context.Background()
			ctx = context.WithValue(ctx, ContextKeyFeatureConfig, featureConfig)
			ctx = context.WithValue(ctx, ContextKeyAppHostSuffixes, &config.AppHostSuffixes{".builtin.com"})
			ctx = context.WithValue(ctx, ContextKeyDomainService, domainService)

			_, err := descriptor.UpdateResource(
				ctx,
				nil,
				&resource.ResourceFile{
					Location: resource.Location{
						Fs:   app,
						Path: path,
					},
					Data: []byte(`id: test
http:
  public_origin: http://test
`),
				},
				[]byte(`id: test
http:
  public_origin: http://test.builtin.com
`),
			)

			So(err, ShouldBeNil)
		})

		Convey("Public origin can be changed to custom domain", func() {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			domainService := NewMockDomainService(ctrl)
			domainService.EXPECT().ListDomains("test").Return([]*apimodel.Domain{
				{
					ID:     "domain-id",
					AppID:  "test",
					Domain: "customdomain.com",
				},
			}, nil).AnyTimes()

			featureConfig := config.NewEffectiveDefaultFeatureConfig()

			ctx := context.Background()
			ctx = context.WithValue(ctx, ContextKeyFeatureConfig, featureConfig)
			ctx = context.WithValue(ctx, ContextKeyAppHostSuffixes, &config.AppHostSuffixes{})
			ctx = context.WithValue(ctx, ContextKeyDomainService, domainService)

			_, err := descriptor.UpdateResource(
				ctx,
				nil,
				&resource.ResourceFile{
					Location: resource.Location{
						Fs:   app,
						Path: path,
					},
					Data: []byte(`id: test
http:
  public_origin: http://test
`),
				},
				[]byte(`id: test
http:
  public_origin: http://customdomain.com
`),
			)

			So(err, ShouldBeNil)
		})

		Convey("Public origin cannot be changed to unknown domain", func() {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			domainService := NewMockDomainService(ctrl)
			domainService.EXPECT().ListDomains("test").Return([]*apimodel.Domain{
				{
					ID:     "domain-id",
					AppID:  "test",
					Domain: "customdomain.com",
				},
			}, nil).AnyTimes()

			featureConfig := config.NewEffectiveDefaultFeatureConfig()

			ctx := context.Background()
			ctx = context.WithValue(ctx, ContextKeyFeatureConfig, featureConfig)
			ctx = context.WithValue(ctx, ContextKeyAppHostSuffixes, &config.AppHostSuffixes{})
			ctx = context.WithValue(ctx, ContextKeyDomainService, domainService)

			_, err := descriptor.UpdateResource(
				ctx,
				nil,
				&resource.ResourceFile{
					Location: resource.Location{
						Fs:   app,
						Path: path,
					},
					Data: []byte(`id: test
http:
  public_origin: http://test
`),
				},
				[]byte(`id: test
http:
  public_origin: http://incorrectdomain.com
`),
			)

			So(err, ShouldBeError, "invalid authgear.yaml:\n/http/public_origin: public origin is not allowed")
		})
	})

	Convey("AuthgearYAML feature config", t, func() {
		path := "authgear.yaml"
		app := resource.LeveledAferoFs{FsLevel: resource.FsLevelApp}
		descriptor := &AuthgearYAMLDescriptor{}

		Convey("test unlimited plan", func() {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			domainService := NewMockDomainService(ctrl)
			domainService.EXPECT().ListDomains("test").Return([]*apimodel.Domain{}, nil).AnyTimes()

			featureConfig := configtest.FixtureFeatureConfig(configtest.FixtureUnlimitedPlanName)
			ctx := context.Background()
			ctx = context.WithValue(ctx, ContextKeyFeatureConfig, featureConfig)
			ctx = context.WithValue(ctx, ContextKeyAppHostSuffixes, &config.AppHostSuffixes{})
			ctx = context.WithValue(ctx, ContextKeyDomainService, domainService)

			Convey("should not return feature config error", func() {
				_, err := descriptor.UpdateResource(
					ctx,
					nil,
					&resource.ResourceFile{
						Location: resource.Location{
							Fs:   app,
							Path: path,
						},
						Data: []byte(`
id: app-id
http:
  public_origin: http://test
`),
					},
					[]byte(`
id: app-id
http:
  public_origin: http://test
authentication:
  identities:
  - biometric
  - anonymous
identity:
  oauth:
    providers:
    - alias: facebook
      type: facebook
      client_id: client_a
    - alias: google
      type: google
      client_id: client_a
authenticator:
  password:
    policy:
      min_length: 8
      uppercase_required: true
      lowercase_required: true
      digit_required: true
      symbol_required: true
      minimum_guessable_level: 4
      excluded_keywords:
      - \"text\"
      history_size: 3
      history_days: 30
hook:
  non_blocking_handlers:
  - events:
    - '*'
    url: http://example.com
  - events:
    - '*'
    url: http://example.com
  blocking_handlers:
  - event: user.pre_create
    url: http://example.com
  - event: user.pre_create
    url: http://example.com
oauth:
  clients:
    - name: Test Client
      client_id: test-client
      x_custom_ui_uri: https://custom-auth-webapp.example.com
      redirect_uris:
      - "https://example.com"
    - name: Test Client2
      client_id: test-client2
      redirect_uris:
      - "https://example2.com"
`),
				)
				So(err, ShouldBeNil)
			})

		})

		Convey("test limited plan", func() {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			domainService := NewMockDomainService(ctrl)
			domainService.EXPECT().ListDomains("test").Return([]*apimodel.Domain{}, nil).AnyTimes()

			featureConfig := configtest.FixtureFeatureConfig(configtest.FixtureLimitedPlanName)
			ctx := context.Background()
			ctx = context.WithValue(ctx, ContextKeyFeatureConfig, featureConfig)
			ctx = context.WithValue(ctx, ContextKeyAppHostSuffixes, &config.AppHostSuffixes{})
			ctx = context.WithValue(ctx, ContextKeyDomainService, domainService)

			Convey("should return feature config error", func() {
				_, err := descriptor.UpdateResource(
					ctx,
					nil,
					&resource.ResourceFile{
						Location: resource.Location{
							Fs:   app,
							Path: path,
						},
						Data: []byte(`
id: app-id
http:
  public_origin: http://test
`),
					},
					[]byte(`
id: app-id
http:
  public_origin: http://test
authentication:
  identities:
  - biometric
  - anonymous
identity:
  oauth:
    providers:
    - alias: facebook
      type: facebook
      client_id: client_a
    - alias: google
      type: google
      client_id: client_a
authenticator:
  password:
    policy:
      min_length: 8
      uppercase_required: true
      lowercase_required: true
      digit_required: true
      symbol_required: true
      minimum_guessable_level: 4
      excluded_keywords:
      - \"text\"
      history_size: 3
      history_days: 30
hook:
  non_blocking_handlers:
  - events:
    - '*'
    url: http://example.com
  - events:
    - '*'
    url: http://example.com
  blocking_handlers:
  - event: user.pre_create
    url: http://example.com
  - event: user.pre_create
    url: http://example.com
oauth:
  clients:
    - name: Test Client
      client_id: test-client
      redirect_uris:
      - "https://example.com"
      x_custom_ui_uri: https://custom-auth-webapp.example.com
    - name: Test Client2
      client_id: test-client2
      redirect_uris:
      - "https://example2.com"
`),
				)
				So(err, ShouldBeError, `invalid authgear.yaml:
/oauth/clients: exceed the maximum number of oauth clients, actual: 2, expected: 1
/identity/oauth/providers: exceed the maximum number of sso providers, actual: 2, expected: 1
/hook/blocking_handlers: exceed the maximum number of blocking handlers, actual: 2, expected: 1
/hook/non_blocking_handlers: exceed the maximum number of non blocking handlers, actual: 2, expected: 1
/authentication/identities: enabling biometric authentication is not supported
/authenticator/password/policy/minimum_guessable_level: minimum_guessable_level is disallowed
/authenticator/password/policy/excluded_keywords: excluded_keywords is disallowed
/authenticator/password/policy: password history is disallowed
/oauth/clients/0: custom ui is disallowed`)
			})

			Convey("should allow saving with the same feature config error", func() {
				_, err := descriptor.UpdateResource(
					ctx,
					nil,
					&resource.ResourceFile{
						Location: resource.Location{
							Fs:   app,
							Path: path,
						},
						Data: []byte(`
id: app-id
http:
  public_origin: http://test
authentication:
  identities:
  - biometric
  - anonymous
identity:
  oauth:
    providers:
    - alias: facebook
      type: facebook
      client_id: client_a
    - alias: google
      type: google
      client_id: client_a
authenticator:
  password:
    policy:
      min_length: 8
      uppercase_required: true
      lowercase_required: true
      digit_required: true
      symbol_required: true
      minimum_guessable_level: 4
      excluded_keywords:
      - \"text\"
      history_size: 3
      history_days: 30
hook:
  non_blocking_handlers:
  - events:
    - '*'
    url: http://example.com
  - events:
    - '*'
    url: http://example.com
  blocking_handlers:
  - event: user.pre_create
    url: http://example.com
  - event: user.pre_create
    url: http://example.com
oauth:
  clients:
    - name: Test Client
      client_id: test-client
      redirect_uris:
      - "https://example.com"
    - name: Test Client2
      client_id: test-client2
      redirect_uris:
      - "https://example2.com"
`),
					},
					[]byte(`
id: app-id
http:
  public_origin: http://test
authentication:
  identities:
  - biometric
  - anonymous
identity:
  oauth:
    providers:
    - alias: facebook
      type: facebook
      client_id: client_a
    - alias: google
      type: google
      client_id: client_a
authenticator:
  password:
    policy:
      min_length: 8
      uppercase_required: true
      lowercase_required: true
      digit_required: true
      symbol_required: true
      minimum_guessable_level: 4
      excluded_keywords:
      - \"text\"
      history_size: 3
      history_days: 30
hook:
  non_blocking_handlers:
  - events:
    - '*'
    url: http://example.com
  - events:
    - '*'
    url: http://example.com
  blocking_handlers:
  - event: user.pre_create
    url: http://example.com
  - event: user.pre_create
    url: http://example.com
oauth:
  clients:
    - name: Test Client
      client_id: test-client
      redirect_uris:
      - "https://example.com"
    - name: Test Client2
      client_id: test-client2
      redirect_uris:
      - "https://example2.com"
`),
				)
				So(err, ShouldBeNil)
			})

			Convey("should return new feature config error", func() {
				_, err := descriptor.UpdateResource(
					ctx,
					nil,
					&resource.ResourceFile{
						Location: resource.Location{
							Fs:   app,
							Path: path,
						},
						Data: []byte(`
id: app-id
http:
  public_origin: http://test
authentication:
  identities:
  - oauth
identity:
  oauth:
    providers:
    - alias: facebook
      type: facebook
      client_id: client_a
    - alias: google
      type: google
      client_id: client_a
oauth:
  clients:
    - name: Test Client
      client_id: test-client
      redirect_uris:
      - "https://example.com"
    - name: Test Client2
      client_id: test-client2
      redirect_uris:
      - "https://example2.com"
authenticator:
  password:
    policy:
        minimum_guessable_level: 4
`),
					},
					[]byte(`
id: app-id
http:
  public_origin: http://test
authentication:
  identities:
  - oauth
  - biometric
identity:
  oauth:
    providers:
    - alias: facebook
      type: facebook
      client_id: client_a
    - alias: google
      type: google
      client_id: client_a
oauth:
  clients:
    - name: Test Client
      client_id: test-client
      redirect_uris:
      - "https://example.com"
    - name: Test Client2
      client_id: test-client2
      redirect_uris:
      - "https://example2.com"
    - name: Test Client3
      client_id: test-client3
      redirect_uris:
      - "https://example3.com"
authenticator:
  password:
    policy:
      minimum_guessable_level: 4
      excluded_keywords:
      - \"text\"
      history_size: 3
      history_days: 30
`),
				)
				So(err, ShouldBeError, `invalid authgear.yaml:
/oauth/clients: exceed the maximum number of oauth clients, actual: 3, expected: 1
/authentication/identities: enabling biometric authentication is not supported
/authenticator/password/policy/excluded_keywords: excluded_keywords is disallowed
/authenticator/password/policy: password history is disallowed`)
			})

			Convey("should return both validation error and feature config error", func() {
				_, err := descriptor.UpdateResource(
					ctx,
					nil,
					&resource.ResourceFile{
						Location: resource.Location{
							Fs:   app,
							Path: path,
						},
						Data: []byte(`id: test
http:
  public_origin: http://test
user_profile:
  custom_attributes:
    attributes:
    - id: "0000"
      pointer: /a
      type: string
`),
					},
					[]byte(`id: test
http:
  public_origin: http://test
authentication:
  identities:
  - biometric
`),
				)
				So(err, ShouldBeError, `invalid authgear.yaml:
/user_profile/custom_attributes/attributes: custom attribute of id '0000' was deleted
/authentication/identities: enabling biometric authentication is not supported`)
			})

		})
	})
	Convey("AuthgearYAML oauth client", t, func() {
		path := "authgear.yaml"
		app := resource.LeveledAferoFs{FsLevel: resource.FsLevelApp}
		descriptor := &AuthgearYAMLDescriptor{}

		Convey("test disallow client id change", func() {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			domainService := NewMockDomainService(ctrl)
			domainService.EXPECT().ListDomains("test").Return([]*apimodel.Domain{}, nil).AnyTimes()

			featureConfig := configtest.FixtureFeatureConfig(configtest.FixtureUnlimitedPlanName)
			ctx := context.Background()
			ctx = context.WithValue(ctx, ContextKeyFeatureConfig, featureConfig)
			ctx = context.WithValue(ctx, ContextKeyAppHostSuffixes, &config.AppHostSuffixes{})
			ctx = context.WithValue(ctx, ContextKeyDomainService, domainService)

			Convey("should not allow changing client id", func() {
				_, err := descriptor.UpdateResource(
					ctx,
					nil,
					&resource.ResourceFile{
						Location: resource.Location{
							Fs:   app,
							Path: path,
						},
						Data: []byte(`
id: app-id
http:
  public_origin: http://test
oauth:
  clients:
    - name: Test Client
      client_id: foo-client
      x_custom_ui_uri: https://custom-auth-webapp.example.com
      redirect_uris:
      - "https://example.com"
`),
					},
					[]byte(`
id: app-id
http:
  public_origin: http://test
oauth:
  clients:
    - name: Test Client
      client_id: bar-client
      x_custom_ui_uri: https://custom-auth-webapp.example.com
      redirect_uris:
      - "https://example.com"
`),
				)
				So(err, ShouldBeError, `invalid authgear.yaml:
/oauth/clients: client ids cannot be changed`)
			})

			Convey("should allow adding client id", func() {
				_, err := descriptor.UpdateResource(
					ctx,
					nil,
					&resource.ResourceFile{
						Location: resource.Location{
							Fs:   app,
							Path: path,
						},
						Data: []byte(`
id: app-id
http:
  public_origin: http://test
oauth:
  clients:
    - name: Test Client
      client_id: foo-client
      x_custom_ui_uri: https://custom-auth-webapp.example.com
      redirect_uris:
      - "https://example.com"
`),
					},
					[]byte(`
id: app-id
http:
  public_origin: http://test
oauth:
  clients:
    - name: Test Client
      client_id: foo-client
      x_custom_ui_uri: https://custom-auth-webapp.example.com
      redirect_uris:
      - "https://example.com"
    - name: Test Client 2
      client_id: bar-client
      x_custom_ui_uri: https://custom-auth-webapp.example2.com
      redirect_uris:
      - "https://example2.com"
`),
				)
				So(err, ShouldBeNil)
			})

			Convey("should allow removing client id", func() {
				_, err := descriptor.UpdateResource(
					ctx,
					nil,
					&resource.ResourceFile{
						Location: resource.Location{
							Fs:   app,
							Path: path,
						},
						Data: []byte(`
id: app-id
http:
  public_origin: http://test
oauth:
  clients:
    - name: Test Client
      client_id: foo-client
      x_custom_ui_uri: https://custom-auth-webapp.example.com
      redirect_uris:
      - "https://example.com"
    - name: Test Client 2
      client_id: bar-client
      x_custom_ui_uri: https://custom-auth-webapp.example2.com
      redirect_uris:
      - "https://example2.com"
`),
					},
					[]byte(`
id: app-id
http:
  public_origin: http://test
oauth:
  clients:
    - name: Test Client 2
      client_id: bar-client
      x_custom_ui_uri: https://custom-auth-webapp.example2.com
      redirect_uris:
      - "https://example2.com"
`),
				)
				So(err, ShouldBeNil)
			})
		})
	})
}
