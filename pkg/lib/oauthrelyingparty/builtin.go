package oauthrelyingparty

const (
	TypeGoogle     = "google"
	TypeFacebook   = "facebook"
	TypeGithub     = "github"
	TypeLinkedin   = "linkedin"
	TypeAzureADv2  = "azureadv2"
	TypeAzureADB2C = "azureadb2c"
	TypeADFS       = "adfs"
	TypeApple      = "apple"
	TypeWechat     = "wechat"
)

var BuiltinProviderTypes = []string{
	TypeGoogle,
	TypeFacebook,
	TypeGithub,
	TypeLinkedin,
	TypeAzureADv2,
	TypeAzureADB2C,
	TypeADFS,
	TypeApple,
	TypeWechat,
}
