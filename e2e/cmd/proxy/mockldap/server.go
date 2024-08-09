package mockldap

import "net"

type MockLDAPServer struct {
}

func NewMockLDAPServer() (*MockLDAPServer, error) {
	return &MockLDAPServer{}, nil
}

func (s *MockLDAPServer) Start(ln net.Listener) error {
	return nil
}

func (s *MockLDAPServer) Shutdown() error {
	return nil
}
