package mockbotprotection

import (
	"github.com/gorilla/mux"
)

type MockBotProtection struct {
	Provider Provider
	Addr     string
}

type Config struct {
}

func (m *MockBotProtection) Attach(router *mux.Router) {
	router.HandleFunc(VerifyEndpoint, m.Verify).Methods("POST")
}

func (m *MockBotProtection) VerifyEndpoint() string {
	return m.Addr + VerifyEndpoint
}
