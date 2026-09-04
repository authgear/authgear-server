package cmddatabase

// pruneTableNames lists every table keyed by app_id in the main database.
//
// Unlike tableNames (used by dump/restore, which lists parents before
// children so restore can satisfy foreign keys on insert), this list is
// ordered children-before-parents so that a prune can delete rows without
// violating foreign key constraints. It is also kept independently because
// tableNames predates several tables (roles/groups, resources, ldap
// identity, bearer token authenticator) and fixing that here would change
// the behavior of the existing dump/restore commands.
var pruneTableNames []string = []string{
	// Identities, keyed off _auth_identity.
	"_auth_identity_anonymous",
	"_auth_identity_biometric",
	"_auth_identity_ldap",
	"_auth_identity_login_id",
	"_auth_identity_oauth",
	"_auth_identity_passkey",
	"_auth_identity_siwe",
	"_auth_identity",

	// Authenticators, keyed off _auth_authenticator.
	// (bearer token and the original per-authenticator recovery code table
	// were dropped by 20200728164433-refactor_authenticators.sql and never
	// recreated; recovery codes now live in _auth_recovery_code below.)
	"_auth_authenticator_oob",
	"_auth_authenticator_passkey",
	"_auth_authenticator_password",
	"_auth_authenticator_totp",
	"_auth_authenticator",

	// Other tables referencing _auth_user directly.
	"_auth_password_history",
	"_auth_oauth_authorization",
	"_auth_recovery_code",
	"_auth_verified_claim",

	// Roles and groups.
	"_auth_group_role",
	"_auth_user_role",
	"_auth_user_group",
	"_auth_group",
	"_auth_role",

	// Resources and scopes.
	"_auth_client_resource_scope",
	"_auth_client_resource",
	"_auth_resource_scope",
	"_auth_resource",

	// OAuth dynamic client registration, keyed only by app_id (no FK to
	// _auth_user).
	"_auth_oauth_client",
	"_auth_oauth_initial_access_token",

	// _auth_user must be last: every table above references it.
	"_auth_user",
}
