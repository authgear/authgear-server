package mockldap

import (
	"errors"

	ldapserver "github.com/vjeantet/ldapserver"
)

type MockLDAPServer struct {
	server *ldapserver.Server
}

func NewMockLDAPServer() (*MockLDAPServer, error) {
	return &MockLDAPServer{}, nil
}

func (s *MockLDAPServer) Start(addr string) error {
	if s.server != nil {
		return errors.New("server already started")
	}

	handler, err := NewLDAPRouteHandler()
	if err != nil {
		return err
	}

	server := ldapserver.NewServer()
	routes := ldapserver.NewRouteMux()
	routes.Bind(handler.HandleBind)
	routes.Search(handler.HandleSearch)
	server.Handle(routes)
	s.server = server

	go func() {
		err := s.server.ListenAndServe(addr)
		if err != nil {
			panic(err)
		}
	}()
	return nil
}

func (s *MockLDAPServer) Shutdown() error {
	if s.server == nil {
		return nil
	}
	s.server.Stop()
	return nil
}
