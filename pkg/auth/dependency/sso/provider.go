package sso

type Provider interface {
	EncodeState(state State) (encodedState string, err error)
	DecodeState(encodedState string) (*State, error)

	EncodeSkygearAuthorizationCode(SkygearAuthorizationCode) (code string, err error)
	DecodeSkygearAuthorizationCode(code string) (*SkygearAuthorizationCode, error)
}
