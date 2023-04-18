package interaction

type OOBType string

const (
	OOBTypeSetupPrimary          OOBType = "setup-primary-oob"
	OOBTypeSetupSecondary        OOBType = "setup-secondary-oob"
	OOBTypeAuthenticatePrimary   OOBType = "authenticate-primary-oob"
	OOBTypeAuthenticateSecondary OOBType = "authenticate-secondary-oob"
)
