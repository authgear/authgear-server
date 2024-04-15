package mockoidc

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type MockOIDCManager struct {
	Providers []Provider

	OIDCs []*MockOIDC

	Clock   Clock
	Server  *http.Server
	Keypair *Keypair
}

func NewMockOIDCManager() (*MockOIDCManager, error) {
	keypair, err := NewKeypair(nil)
	if err != nil {
		return nil, err
	}

	return &MockOIDCManager{
		Providers: SupportedProviders,
		Clock:     NewSystemClock(),
		Keypair:   keypair,
	}, nil
}

func (m *MockOIDCManager) Start(ln net.Listener) error {
	if m.Server != nil {
		return errors.New("server already started")
	}

	router := mux.NewRouter()

	m.Server = &http.Server{
		Addr:    ln.Addr().String(),
		Handler: router,
	}

	m.OIDCs = nil
	for _, provider := range m.Providers {
		pathPrefix := "/" + provider.Type
		subrouter := router.PathPrefix(pathPrefix).Subrouter()

		oidc := &MockOIDC{
			Addr:                          m.Addr() + pathPrefix,
			Provider:                      provider,
			ClientID:                      provider.Type,
			ClientSecret:                  "mock",
			Keypair:                       m.Keypair,
			Clock:                         m.Clock,
			AccessTTL:                     time.Duration(10) * time.Minute,
			RefreshTTL:                    time.Duration(60) * time.Minute,
			CodeChallengeMethodsSupported: []string{"plain", "S256"},
			SessionStore:                  NewSessionStore(),
		}
		oidc.Attach(subrouter)

		m.OIDCs = append(m.OIDCs, oidc)
	}

	go func() {
		err := m.Server.Serve(ln)
		if err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	return nil
}

func (m *MockOIDCManager) Addr() string {
	if m.Server == nil {
		return ""
	}
	return fmt.Sprintf("http://%s", m.Server.Addr)
}

func (m *MockOIDCManager) Shutdown() {
	if m.Server != nil {
		m.Server.Close()
		m.Server = nil
	}
}

func (m *MockOIDCManager) GetOIDC(alias string) *MockOIDC {
	for _, oidc := range m.OIDCs {
		if oidc.Provider.Type == alias {
			return oidc
		}
	}
	return nil
}
