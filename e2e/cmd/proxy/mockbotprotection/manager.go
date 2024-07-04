package mockbotprotection

import (
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

type MockBotProtectionManager struct {
	Providers []Provider

	BotProtections []*MockBotProtection

	Server *http.Server
}

func NewMockBotProtectionManager() (*MockBotProtectionManager, error) {
	return &MockBotProtectionManager{
		Providers: SupportedProviders,
	}, nil
}

func (m *MockBotProtectionManager) Start(ln net.Listener) error {
	if m.Server != nil {
		return errors.New("server already started")
	}

	router := mux.NewRouter()

	m.Server = &http.Server{
		Addr:    ln.Addr().String(),
		Handler: router,
	}

	m.BotProtections = nil
	for _, provider := range m.Providers {
		pathPrefix := "/" + string(provider.Type)
		subrouter := router.PathPrefix(pathPrefix).Subrouter()
		botprotection := &MockBotProtection{
			Addr:     m.Addr() + pathPrefix,
			Provider: provider,
		}
		botprotection.Attach(subrouter)
		m.BotProtections = append(m.BotProtections, botprotection)
	}

	go func() {
		err := m.Server.Serve(ln)
		if err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	return nil
}

func (m *MockBotProtectionManager) Addr() string {
	if m.Server == nil {
		return ""
	}
	return fmt.Sprintf("http://%s", m.Server.Addr)
}

func (m *MockBotProtectionManager) Shutdown() {
	if m.Server != nil {
		m.Server.Close()
		m.Server = nil
	}
}

func (m *MockBotProtectionManager) GetBotProtection(providerType config.BotProtectionProviderType) *MockBotProtection {
	for _, bp := range m.BotProtections {
		if bp.Provider.Type == providerType {
			return bp
		}
	}
	return nil
}
