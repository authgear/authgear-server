package cmddatabase

// The order of the list is important because it defines the restoration order
// The referenced table must precede the referencing table
var tableNames []string = []string{
	"_auth_user",
	"_auth_authenticator",
	"_auth_authenticator_oob",
	"_auth_authenticator_passkey",
	"_auth_authenticator_password",
	"_auth_authenticator_totp",
	"_auth_identity",
	"_auth_identity_anonymous",
	"_auth_identity_biometric",
	"_auth_identity_login_id",
	"_auth_identity_oauth",
	"_auth_identity_passkey",
	"_auth_identity_siwe",
	"_auth_oauth_authorization",
	"_auth_password_history",
	"_auth_recovery_code",
	"_auth_verified_claim",
}
