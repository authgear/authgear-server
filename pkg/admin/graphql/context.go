package graphql

import (
	"context"
	"time"

	"github.com/authgear/authgear-server/pkg/admin/model"
	"github.com/authgear/authgear-server/pkg/api/event"
	apimodel "github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/audit"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	libuser "github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/feature/forgotpassword"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/lib/rolesgroups"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/sessionlisting"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type UserLoader interface {
	graphqlutil.DataLoaderInterface
}

type IdentityLoader interface {
	graphqlutil.DataLoaderInterface
}

type AuthenticatorLoader interface {
	graphqlutil.DataLoaderInterface
}

type RoleLoader interface {
	graphqlutil.DataLoaderInterface
}

type GroupLoader interface {
	graphqlutil.DataLoaderInterface
}

type AuditLogLoader interface {
	graphqlutil.DataLoaderInterface
}

type AuditLogFacade interface {
	QueryPage(ctx context.Context, opts audit.QueryPageOptions, pageArgs graphqlutil.PageArgs) ([]apimodel.PageItemRef, *graphqlutil.PageResult, error)
}

type UserFacade interface {
	ListPage(ctx context.Context, listOption libuser.ListOptions, args graphqlutil.PageArgs) ([]apimodel.PageItemRef, *graphqlutil.PageResult, error)
	SearchPage(
		ctx context.Context,
		searchKeyword string,
		filterOptions libuser.FilterOptions,
		sortOption libuser.SortOption,
		args graphqlutil.PageArgs) ([]apimodel.PageItemRef, *graphqlutil.PageResult, error)
	Create(ctx context.Context, identityDef model.IdentityDef, password string, generatePassword bool, sendPassword bool, setPasswordExpired bool) (string, error)
	ResetPassword(ctx context.Context, id string, password string, generatePassword bool, sendPassword bool, changeOnLogin bool) error
	SetPasswordExpired(ctx context.Context, id string, isExpired bool) error
	SetDisabled(ctx context.Context, id string, isDisabled bool, reason *string) error
	ScheduleDeletion(ctx context.Context, id string) error
	UnscheduleDeletion(ctx context.Context, id string) error
	Delete(ctx context.Context, id string) error
	ScheduleAnonymization(ctx context.Context, id string) error
	UnscheduleAnonymization(ctx context.Context, id string) error
	Anonymize(ctx context.Context, id string) error
	SetMFAGracePeriod(ctx context.Context, id string, endAt *time.Time) error
	GetUsersByStandardAttribute(ctx context.Context, attributeKey string, attributeValue string) ([]string, error)
	GetUserByLoginID(ctx context.Context, loginIDKey string, loginIDValue string) (string, error)
	GetUserByOAuth(ctx context.Context, oauthProviderAlias string, oauthProviderUserID string) (string, error)
}

type RolesGroupsFacade interface {
	CreateRole(ctx context.Context, options *rolesgroups.NewRoleOptions) (string, error)
	UpdateRole(ctx context.Context, options *rolesgroups.UpdateRoleOptions) error
	DeleteRole(ctx context.Context, id string) error
	ListGroupsByRoleID(ctx context.Context, roleID string) ([]*apimodel.Group, error)
	ListRoles(ctx context.Context, options *rolesgroups.ListRolesOptions, pageArgs graphqlutil.PageArgs) ([]apimodel.PageItemRef, *graphqlutil.PageResult, error)

	CreateGroup(ctx context.Context, options *rolesgroups.NewGroupOptions) (string, error)
	UpdateGroup(ctx context.Context, options *rolesgroups.UpdateGroupOptions) error
	DeleteGroup(ctx context.Context, id string) error
	ListRolesByGroupID(ctx context.Context, groupID string) ([]*apimodel.Role, error)
	ListGroups(ctx context.Context, options *rolesgroups.ListGroupsOptions, pageArgs graphqlutil.PageArgs) ([]apimodel.PageItemRef, *graphqlutil.PageResult, error)

	AddRoleToGroups(ctx context.Context, options *rolesgroups.AddRoleToGroupsOptions) (string, error)
	RemoveRoleFromGroups(ctx context.Context, options *rolesgroups.RemoveRoleFromGroupsOptions) (string, error)

	AddRoleToUsers(ctx context.Context, options *rolesgroups.AddRoleToUsersOptions) (string, error)
	RemoveRoleFromUsers(ctx context.Context, options *rolesgroups.RemoveRoleFromUsersOptions) (string, error)

	AddGroupToUsers(ctx context.Context, options *rolesgroups.AddGroupToUsersOptions) (groupID string, err error)
	RemoveGroupFromUsers(ctx context.Context, options *rolesgroups.RemoveGroupFromUsersOptions) (groupID string, err error)

	AddGroupToRoles(ctx context.Context, options *rolesgroups.AddGroupToRolesOptions) (groupID string, err error)
	RemoveGroupFromRoles(ctx context.Context, options *rolesgroups.RemoveGroupFromRolesOptions) (groupID string, err error)

	AddUserToRoles(ctx context.Context, options *rolesgroups.AddUserToRolesOptions) (err error)
	RemoveUserFromRoles(ctx context.Context, options *rolesgroups.RemoveUserFromRolesOptions) (err error)

	AddUserToGroups(ctx context.Context, options *rolesgroups.AddUserToGroupsOptions) (err error)
	RemoveUserFromGroups(ctx context.Context, options *rolesgroups.RemoveUserFromGroupsOptions) (err error)

	ListRolesByUserID(ctx context.Context, userID string) ([]*apimodel.Role, error)
	ListGroupsByUserID(ctx context.Context, userID string) ([]*apimodel.Group, error)
	ListUserIDsByRoleID(ctx context.Context, roleID string, pageArgs graphqlutil.PageArgs) ([]apimodel.PageItemRef, *graphqlutil.PageResult, error)
	ListUserIDsByGroupID(ctx context.Context, groupID string, pageArgs graphqlutil.PageArgs) ([]apimodel.PageItemRef, *graphqlutil.PageResult, error)
	ListEffectiveRolesByUserID(ctx context.Context, userID string) ([]*apimodel.Role, error)
	ListAllUserIDsByGroupIDs(ctx context.Context, groupIDs []string) ([]string, error)
	ListAllUserIDsByGroupKeys(ctx context.Context, groupKeys []string) ([]string, error)
	ListAllUserIDsByRoleIDs(ctx context.Context, roleIDs []string) ([]string, error)
	ListAllUserIDsByEffectiveRoleIDs(ctx context.Context, roleIDs []string) ([]string, error)
	ListAllRolesByKeys(ctx context.Context, keys []string) ([]*apimodel.Role, error)
	ListAllGroupsByKeys(ctx context.Context, keys []string) ([]*apimodel.Group, error)

	GetRole(ctx context.Context, roleID string) (*apimodel.Role, error)
	GetGroup(ctx context.Context, groupID string) (*apimodel.Group, error)
}

type IdentityFacade interface {
	Get(ctx context.Context, id string) (*identity.Info, error)
	List(ctx context.Context, userID string, identityType *apimodel.IdentityType) ([]*apimodel.IdentityRef, error)
	Remove(ctx context.Context, identityInfo *identity.Info) error
	Create(ctx context.Context, userID string, identityDef model.IdentityDef, password string) (*apimodel.IdentityRef, error)
	Update(ctx context.Context, identityID string, userID string, identityDef model.IdentityDef) (*apimodel.IdentityRef, error)
}

type AuthenticatorFacade interface {
	Get(ctx context.Context, id string) (*authenticator.Info, error)
	List(ctx context.Context, userID string, authenticatorType *apimodel.AuthenticatorType, authenticatorKind *authenticator.Kind) ([]*authenticator.Ref, error)
	Remove(ctx context.Context, authenticatorInfo *authenticator.Info) error
	CreateBySpec(ctx context.Context, spec *authenticator.Spec) (*authenticator.Info, error)
}

type VerificationFacade interface {
	Get(ctx context.Context, userID string) ([]model.Claim, error)
	SetVerified(ctx context.Context, userID string, claimName string, claimValue string, isVerified bool) error
}

type UserProfileFacade interface {
	DeriveStandardAttributes(ctx context.Context, role accesscontrol.Role, userID string, updatedAt time.Time, attrs map[string]interface{}) (map[string]interface{}, error)
	ReadCustomAttributesInStorageForm(ctx context.Context, role accesscontrol.Role, userID string, storageForm map[string]interface{}) (map[string]interface{}, error)
	UpdateUserProfile(
		ctx context.Context,
		role accesscontrol.Role,
		userID string,
		stdAttrs map[string]interface{},
		customAttrs map[string]interface{},
	) error
}

type SessionFacade interface {
	List(ctx context.Context, userID string) ([]session.ListableSession, error)
	Get(ctx context.Context, id string) (session.ListableSession, error)
	Revoke(ctx context.Context, id string) error
	RevokeAll(ctx context.Context, userID string) error
}

type AuthorizationFacade interface {
	Get(ctx context.Context, id string) (*oauth.Authorization, error)
	List(ctx context.Context, userID string, filters ...oauth.AuthorizationFilter) ([]*oauth.Authorization, error)
	Delete(ctx context.Context, a *oauth.Authorization) error
}

type OAuthFacade interface {
	CreateSession(ctx context.Context, clientID string, userID string, deviceInfo map[string]interface{}) (session.ListableSession, protocol.TokenResponse, error)
}

type SessionListingService interface {
	FilterForDisplay(ctx context.Context, sessions []session.ListableSession, currentSession session.ResolvedSession) ([]*sessionlisting.Session, error)
}

type OTPCodeService interface {
	GenerateOTP(ctx context.Context, kind otp.Kind, target string, form otp.Form, opts *otp.GenerateOptions) (string, error)
}

type ForgotPasswordService interface {
	SendCode(ctx context.Context, loginID string, options *forgotpassword.CodeOptions) error
}

type EventService interface {
	DispatchEventOnCommit(ctx context.Context, payload event.Payload) error
}

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger { return Logger{lf.New("admin-graphql")} }

type Context struct {
	GQLLogger Logger

	Config                *config.AppConfig
	OAuthConfig           *config.OAuthConfig
	AdminAPIFeatureConfig *config.AdminAPIFeatureConfig

	Users          UserLoader
	Identities     IdentityLoader
	Authenticators AuthenticatorLoader
	Roles          RoleLoader
	Groups         GroupLoader
	AuditLogs      AuditLogLoader

	UserFacade          UserFacade
	RolesGroupsFacade   RolesGroupsFacade
	AuditLogFacade      AuditLogFacade
	IdentityFacade      IdentityFacade
	AuthenticatorFacade AuthenticatorFacade
	VerificationFacade  VerificationFacade
	SessionFacade       SessionFacade
	UserProfileFacade   UserProfileFacade
	AuthorizationFacade AuthorizationFacade
	OAuthFacade         OAuthFacade
	SessionListing      SessionListingService
	OTPCode             OTPCodeService
	ForgotPassword      ForgotPasswordService
	Events              EventService
}

func (c *Context) Logger() *log.Logger {
	return c.GQLLogger.Logger
}

func WithContext(ctx context.Context, gqlContext *Context) context.Context {
	return graphqlutil.WithContext(ctx, gqlContext)
}

func GQLContext(ctx context.Context) *Context {
	return graphqlutil.GQLContext(ctx).(*Context)
}
