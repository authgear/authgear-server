package config

type AdminAPIAuth string

const (
	AdminAPIAuthNone AdminAPIAuth = "none"
	AdminAPIAuthJWT  AdminAPIAuth = "jwt"
)
