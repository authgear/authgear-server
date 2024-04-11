package mockoidc

import (
	"time"

	"github.com/gorilla/mux"
)

type MockOIDC struct {
	Provider     Provider
	ClientID     string
	ClientSecret string

	AccessTTL  time.Duration
	RefreshTTL time.Duration

	CodeChallengeMethodsSupported []string

	Addr         string
	Keypair      *Keypair
	SessionStore *SessionStore

	Clock Clock
}

type Config struct {
	ClientID     string
	ClientSecret string
	Issuer       string

	AccessTTL  time.Duration
	RefreshTTL time.Duration

	CodeChallengeMethodsSupported []string
}

func (m *MockOIDC) Attach(router *mux.Router) {
	router.HandleFunc(AuthorizationEndpoint, m.Authorize).Methods("GET")
	router.HandleFunc(TokenEndpoint, m.Token).Methods("POST")
	router.HandleFunc(UserinfoEndpoint, m.Userinfo).Methods("GET")
	router.HandleFunc(JWKSEndpoint, m.JWKS).Methods("GET")
	router.HandleFunc(DiscoveryEndpoint, m.Discovery).Methods("GET")
}

func (m *MockOIDC) Config() *Config {
	return &Config{
		Issuer:                        m.Provider.Issuer,
		ClientID:                      m.ClientID,
		ClientSecret:                  m.ClientSecret,
		CodeChallengeMethodsSupported: m.CodeChallengeMethodsSupported,
		AccessTTL:                     m.AccessTTL,
		RefreshTTL:                    m.RefreshTTL,
	}
}

func (m *MockOIDC) Issuer() string {
	return m.Addr + IssuerBase
}

func (m *MockOIDC) DiscoveryEndpoint() string {
	return m.Addr + DiscoveryEndpoint
}

func (m *MockOIDC) AuthorizationEndpoint() string {
	return m.Addr + AuthorizationEndpoint
}

func (m *MockOIDC) TokenEndpoint() string {
	return m.Addr + TokenEndpoint
}

func (m *MockOIDC) UserinfoEndpoint() string {
	return m.Addr + UserinfoEndpoint
}

func (m *MockOIDC) JWKSEndpoint() string {
	return m.Addr + JWKSEndpoint
}
